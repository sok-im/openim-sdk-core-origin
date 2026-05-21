package signaling

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/openimsdk/openim-sdk-core/v3/internal/interaction"
	"github.com/openimsdk/openim-sdk-core/v3/open_im_sdk_callback"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/db_interface"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	pConstant "github.com/openimsdk/protocol/constant"
	"github.com/openimsdk/protocol/rtc"
	"github.com/openimsdk/protocol/sdkws"
	"github.com/openimsdk/tools/log"
	"github.com/openimsdk/tools/utils/datautil"
	"github.com/openimsdk/tools/utils/jsonutil"
	"google.golang.org/protobuf/proto"
)

const roomTimingMaxAge = 10 * time.Minute

// roomTiming 追踪单次通话的关键时间戳（本端视角），以 roomID 为键存于 sync.Map。
type roomTiming struct {
	inviteMs  int64     // 本端发起或收到邀请的毫秒时间戳
	connectMs int64     // 接通时的毫秒时间戳（0 = 尚未接通）
	createdAt time.Time // 条目创建时间，用于定期清理
}

// recordTask 异步写记录任务，包含构建 LocalSignalCallRecord 所需的全部信息。
type recordTask struct {
	inv         *rtc.InvitationInfo
	participant *rtc.ParticipantMetaData
	status      int32
	direction   int32
	inviteMs    int64 // 从 roomTimings 取得
	connectMs   int64 // 从 roomTimings 取得（0 = 未接通）
	endMs       int64 // 通话结束时间（调用处捕获，避免异步延迟误差）
}

// inviteTimer 本端收到/发起邀请后的超时定时器
type inviteTimer struct {
	timer     *time.Timer
	inv       *rtc.InvitationInfo
	direction int32 // 主叫 or 被叫，超时后用于确定记录方向
}

type Signaling struct {
	loginUserID string
	platformID  int32
	longConnMgr *interaction.LongConnMgr
	db          db_interface.DataBase
	listener    func() open_im_sdk_callback.OnSignalingListener

	// roomTimings 记录每个房间的本端时间戳（inviteMs / connectMs）。
	// key: roomID (string) → value: *roomTiming
	roomTimings sync.Map

	// inviteTimers 本端邀请超时定时器，key: roomID
	inviteTimers sync.Map

	// recordCh 用于异步落库通话记录，避免阻塞信令通知路径
	recordCh chan recordTask
	done     chan struct{}

	// detailCache 为 GetLocalSignalCallRecordDetail 提供 LRU + TTL=5min 缓存
	detailCache *detailCache

	// cleanupDone 用于通知定时清理协程退出
	cleanupDone chan struct{}
}

func NewSignaling(longConnMgr *interaction.LongConnMgr, loginUserID string, platformID int32, db db_interface.DataBase) *Signaling {
	s := &Signaling{
		loginUserID: loginUserID,
		platformID:  platformID,
		longConnMgr: longConnMgr,
		db:          db,
		recordCh:    make(chan recordTask, 32),
		done:        make(chan struct{}),
		detailCache: newDetailCache(),
		cleanupDone: make(chan struct{}),
	}
	go s.recordWorker()
	go s.roomTimingsCleanup()
	return s
}

// recordWorker 异步处理通话记录入库。
func (s *Signaling) recordWorker() {
	defer func() {
		if r := recover(); r != nil {
			log.ZError(context.Background(), "recordWorker panic", nil, "recover", r)
		}
	}()

	for task := range s.recordCh {
		if s.db == nil || task.inv == nil {
			continue
		}
		ctx := context.Background()

		calleeText := s.composeCalleeMatchText(ctx, task.inv, task.participant)
		inviteeNickname := s.resolveInviteeNickname(ctx, task.inv, task.participant)
		inviterNickname := s.resolveInviterNickname(ctx, task.inv, task.participant)
		inviterFaceURL := s.resolveInviterFaceURL(ctx, task.inv, task.participant)
		groupName := s.resolveGroupName(ctx, task.inv, task.participant)
		inviteeIDsJSON := buildInviteeIDsJSON(task.inv)

		cfg := callRecordConfig{
			inv:             task.inv,
			status:          task.status,
			direction:       task.direction,
			inviteMs:        task.inviteMs,
			connectMs:       task.connectMs,
			endMs:           task.endMs,
			calleeMatchText: calleeText,
			inviteeNickname: inviteeNickname,
			inviteeIDsJSON:  inviteeIDsJSON,
			inviterNickname: inviterNickname,
			inviterFaceURL:  inviterFaceURL,
			groupName:       groupName,
		}
		lr := newLocalSignalCallRecord(cfg)
		if lr == nil {
			continue
		}
		if err := s.db.BatchUpsertSignalCallRecords(ctx, []*model_struct.LocalSignalCallRecord{lr}); err != nil {
			log.ZWarn(ctx, "async persistLocalCallRecord failed", err, "sID", lr.SID)
		}
	}
	close(s.done)
}

// roomTimingsCleanup 定期清理过期的 roomTimings 条目，防止因异常断连导致的内存泄漏。
func (s *Signaling) roomTimingsCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			s.roomTimings.Range(func(key, value any) bool {
				t := value.(*roomTiming)
				if now.Sub(t.createdAt) > roomTimingMaxAge {
					s.roomTimings.Delete(key)
				}
				return true
			})
		case <-s.cleanupDone:
			return
		}
	}
}

func (s *Signaling) SetListener(listener func() open_im_sdk_callback.OnSignalingListener) {
	s.listener = listener
}

// Close 清理异步 worker、定时器与缓存。
func (s *Signaling) Close() {
	// 停止定时清理协程
	close(s.cleanupDone)

	// 停止所有邀请超时定时器
	s.inviteTimers.Range(func(key, value any) bool {
		it := value.(*inviteTimer)
		it.timer.Stop()
		s.inviteTimers.Delete(key)
		return true
	})

	if s.recordCh != nil {
		close(s.recordCh)
	}
	if s.detailCache != nil {
		s.detailCache.Clear()
	}
	<-s.done
}

// ── 房间时间戳辅助 ────────────────────────────────────────────────────────────

func (s *Signaling) storeInviteTime(roomID string, ms int64) {
	t := &roomTiming{inviteMs: ms, createdAt: time.Now()}
	s.roomTimings.Store(roomID, t)
}

func (s *Signaling) storeConnectTime(roomID string, ms int64) {
	if v, ok := s.roomTimings.Load(roomID); ok {
		v.(*roomTiming).connectMs = ms
	} else {
		s.roomTimings.Store(roomID, &roomTiming{connectMs: ms, createdAt: time.Now()})
	}
}

// popTiming 读取并删除 roomID 对应的时间戳记录；若不存在则返回零值。
func (s *Signaling) popTiming(roomID string) (inviteMs, connectMs int64) {
	if v, ok := s.roomTimings.LoadAndDelete(roomID); ok {
		t := v.(*roomTiming)
		return t.inviteMs, t.connectMs
	}
	return 0, 0
}

// ── 邀请超时定时器 ────────────────────────────────────────────────────────────

// startInviteTimer 启动本端邀请超时定时器。超时后触发 OnInvitationTimeout 回调并写本地记录。
func (s *Signaling) startInviteTimer(inv *rtc.InvitationInfo, direction int32) {
	if inv == nil || inv.RoomID == "" || inv.Timeout <= 0 {
		return
	}

	timeout := time.Duration(inv.Timeout) * time.Second
	it := &inviteTimer{
		inv:       inv,
		direction: direction,
	}
	it.timer = time.AfterFunc(timeout, func() {
		s.inviteTimers.Delete(inv.RoomID)
		s.onInvitationTimeout(inv, direction)
	})
	// 如已有旧定时器（如重复邀请），先停止
	if old, loaded := s.inviteTimers.LoadAndDelete(inv.RoomID); loaded {
		old.(*inviteTimer).timer.Stop()
	}
	s.inviteTimers.Store(inv.RoomID, it)
}

// cancelInviteTimer 取消邀请超时定时器（收到 Accept/Reject/Cancel/HungUp 时调用）。
func (s *Signaling) cancelInviteTimer(roomID string) {
	if v, loaded := s.inviteTimers.LoadAndDelete(roomID); loaded {
		v.(*inviteTimer).timer.Stop()
	}
}

// onInvitationTimeout 本端邀请超时：主叫侧上报服务端；被叫侧仅本地回调与落库。
func (s *Signaling) onInvitationTimeout(inv *rtc.InvitationInfo, direction int32) {
	if inv == nil {
		return
	}
	ctx := context.Background()
	if direction == constant.SignalCallDirectionOutgoing && inv.InviterUserID == s.loginUserID {
		if err := s.Timeout(ctx, &rtc.SignalTimeoutReq{Invitation: inv}); err != nil {
			log.ZWarn(ctx, "Timeout request failed on local timer", err, "roomID", inv.RoomID)
		}
		return
	}
	s.notifyInvitationTimeout(ctx, inv, constant.SignalCallDirectionMissed)
}

// notifyInvitationTimeout 被叫侧超时未接：通知 UI 并写本地未接通记录。
func (s *Signaling) notifyInvitationTimeout(ctx context.Context, inv *rtc.InvitationInfo, recordDir int32) {
	s.cancelInviteTimer(inv.RoomID)

	listener := s.listener()
	if listener != nil {
		log.ZDebug(ctx, "OnInvitationTimeout", "roomID", inv.RoomID, "direction", recordDir)
		listener.OnInvitationTimeout(jsonutil.StructToJsonString(inv))
	}

	inviteMs, connectMs := s.popTiming(inv.RoomID)
	s.persistLocalCallRecord(ctx, inv, nil,
		constant.SignalCallStatusNotConnected,
		recordDir,
		inviteMs, connectMs, time.Now().UnixMilli())
}

// ── 通知路由 ────────────────────────────────────────────────────────────────

func (s *Signaling) DoNotification(ctx context.Context, msg *sdkws.MsgData) {
	if err := s.doNotification(ctx, msg); err != nil {
		log.ZError(ctx, "DoSignalingNotification failed", err, "contentType", msg.ContentType)
	}
}

func (s *Signaling) doNotification(ctx context.Context, msg *sdkws.MsgData) error {
	switch msg.ContentType {
	case pConstant.SignalingNotification:
		return s.handleSignalingNotification(ctx, msg)
	case pConstant.RoomParticipantsConnectedNotification:
		return s.handleRoomParticipantConnected(ctx, msg)
	case pConstant.RoomParticipantsDisconnectedNotification:
		return s.handleRoomParticipantDisconnected(ctx, msg)
	case pConstant.StreamChangedNotification, pConstant.CustomSignalNotification:
		log.ZDebug(ctx, "ignoring signaling notification", "contentType", msg.ContentType)
		return nil
	default:
		log.ZWarn(ctx, "unhandled signaling notification", nil, "contentType", msg.ContentType)
		return nil
	}
}

func (s *Signaling) handleSignalingNotification(ctx context.Context, msg *sdkws.MsgData) error {
	var signalReq rtc.SignalReq
	if err := proto.Unmarshal(msg.Content, &signalReq); err != nil {
		return err
	}
	listener := s.listener()
	if listener == nil {
		log.ZWarn(ctx, "signaling listener is nil, skipping notification", nil)
		return nil
	}

	log.ZDebug(ctx, "handleSignalingNotification", "signalReq", &signalReq)

	switch payload := signalReq.Payload.(type) {
	case *rtc.SignalReq_Invite:
		return s.handleInvite(ctx, listener, payload.Invite)
	case *rtc.SignalReq_InviteInGroup:
		return s.handleInviteInGroup(ctx, listener, payload.InviteInGroup)
	case *rtc.SignalReq_Accept:
		return s.handleAccept(ctx, listener, payload.Accept)
	case *rtc.SignalReq_Reject:
		return s.handleReject(ctx, listener, payload.Reject)
	case *rtc.SignalReq_Cancel:
		return s.handleCancel(ctx, listener, payload.Cancel)
	case *rtc.SignalReq_HungUp:
		return s.handleHungUp(ctx, listener, payload.HungUp)
	case *rtc.SignalReq_Timeout:
		return s.handleTimeout(ctx, listener, payload.Timeout)
	default:
		log.ZWarn(ctx, "unhandled signaling payload type", nil, "type", fmt.Sprintf("%T", signalReq.Payload))
		return nil
	}
}

// handleInvite 被叫侧收到来电邀请：记录来电开始时间，启动超时定时器，触发 UI 回调。
func (s *Signaling) handleInvite(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalInviteReq) error {
	if req.Invitation == nil {
		return nil
	}
	if datautil.Contain(s.loginUserID, req.Invitation.InviteeUserIDList...) {
		inviteMs := time.Now().UnixMilli()
		if req.Invitation.InitiateTime > 0 {
			inviteMs = req.Invitation.InitiateTime
		}
		s.storeInviteTime(req.Invitation.RoomID, inviteMs)
		s.startInviteTimer(req.Invitation, constant.SignalCallDirectionMissed)

		log.ZDebug(ctx, "OnReceiveNewInvitation", "invitation", req)
		listener.OnReceiveNewInvitation(jsonutil.StructToJsonString(req))
	}
	return nil
}

// handleInviteInGroup 被叫侧收到群组来电邀请。
func (s *Signaling) handleInviteInGroup(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalInviteInGroupReq) error {
	if req.Invitation == nil {
		return nil
	}
	if datautil.Contain(s.loginUserID, req.Invitation.InviteeUserIDList...) {
		inviteMs := time.Now().UnixMilli()
		if req.Invitation.InitiateTime > 0 {
			inviteMs = req.Invitation.InitiateTime
		}
		s.storeInviteTime(req.Invitation.RoomID, inviteMs)
		s.startInviteTimer(req.Invitation, constant.SignalCallDirectionMissed)

		log.ZDebug(ctx, "OnReceiveNewInvitation (group)", "invitation", req)
		listener.OnReceiveNewInvitation(jsonutil.StructToJsonString(req))
	}
	return nil
}

// handleAccept 主叫侧收到被叫接听通知：取消超时定时器，记录接通时间。
func (s *Signaling) handleAccept(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalAcceptReq) error {
	if req.Invitation == nil {
		return nil
	}
	if req.Invitation.InviterUserID == s.loginUserID {
		s.cancelInviteTimer(req.Invitation.RoomID)
		s.storeConnectTime(req.Invitation.RoomID, time.Now().UnixMilli())

		log.ZDebug(ctx, "OnInviteeAccepted", "accept", req)
		listener.OnInviteeAccepted(jsonutil.StructToJsonString(req))
		return nil
	}
	if req.UserID == s.loginUserID && req.OpUserPlatformID != s.platformID {
		log.ZDebug(ctx, "OnInviteeAcceptedByOtherDevice", "accept", req)
		listener.OnInviteeAcceptedByOtherDevice(jsonutil.StructToJsonString(req))
	}
	return nil
}

// handleReject 主叫侧收到被叫拒接通知：取消超时定时器，写未拨通记录（主叫-outgoing）。
func (s *Signaling) handleReject(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalRejectReq) error {
	if req.Invitation == nil {
		return nil
	}
	if req.Invitation.InviterUserID == s.loginUserID {
		s.cancelInviteTimer(req.Invitation.RoomID)

		log.ZDebug(ctx, "OnInviteeRejected", "reject", req)
		listener.OnInviteeRejected(jsonutil.StructToJsonString(req))

		inviteMs, connectMs := s.popTiming(req.Invitation.RoomID)
		s.persistLocalCallRecord(ctx, req.Invitation, req.Participant,
			constant.SignalCallStatusNotConnected,
			constant.SignalCallDirectionOutgoing,
			inviteMs, connectMs, time.Now().UnixMilli())
		return nil
	}
	if req.UserID == s.loginUserID && req.OpUserPlatformID != s.platformID {
		log.ZDebug(ctx, "OnInviteeRejectedByOtherDevice", "reject", req)
		listener.OnInviteeRejectedByOtherDevice(jsonutil.StructToJsonString(req))
	}
	return nil
}

// handleTimeout 被叫侧收到主叫超时未接通通知。
func (s *Signaling) handleTimeout(ctx context.Context, _ open_im_sdk_callback.OnSignalingListener, req *rtc.SignalTimeoutReq) error {
	if req.Invitation == nil {
		return nil
	}
	if req.Invitation.InviterUserID == s.loginUserID {
		return nil
	}
	if datautil.Contain(s.loginUserID, req.Invitation.InviteeUserIDList...) {
		s.notifyInvitationTimeout(ctx, req.Invitation, constant.SignalCallDirectionMissed)
	}
	return nil
}

// handleCancel 被叫侧收到主叫取消通知：取消超时定时器，写未接来电记录（被叫-missed）。
func (s *Signaling) handleCancel(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalCancelReq) error {
	if req.Invitation == nil {
		return nil
	}
	if datautil.Contain(s.loginUserID, req.Invitation.InviteeUserIDList...) {
		s.cancelInviteTimer(req.Invitation.RoomID)

		log.ZDebug(ctx, "OnInvitationCancelled", "cancel", req)
		listener.OnInvitationCancelled(jsonutil.StructToJsonString(req))

		inviteMs, connectMs := s.popTiming(req.Invitation.RoomID)
		s.persistLocalCallRecord(ctx, req.Invitation, req.Participant,
			constant.SignalCallStatusNotConnected,
			constant.SignalCallDirectionMissed,
			inviteMs, connectMs, time.Now().UnixMilli())
	}
	return nil
}

// handleHungUp 收到对端挂断通知：取消超时定时器，写已拨通记录。
func (s *Signaling) handleHungUp(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalHungUpReq) error {
	if req.Invitation == nil {
		return nil
	}
	if req.UserID != s.loginUserID {
		s.cancelInviteTimer(req.Invitation.RoomID)

		log.ZDebug(ctx, "OnHangUp", "hungUp", req)
		listener.OnHangUp(jsonutil.StructToJsonString(req))

		direction := constant.SignalCallDirectionIncoming
		if req.Invitation.InviterUserID == s.loginUserID {
			direction = constant.SignalCallDirectionOutgoing
		}
		inviteMs, connectMs := s.popTiming(req.Invitation.RoomID)
		s.persistLocalCallRecord(ctx, req.Invitation, nil,
			constant.SignalCallStatusAnswered,
			direction,
			inviteMs, connectMs, time.Now().UnixMilli())
	}
	return nil
}

func (s *Signaling) handleRoomParticipantConnected(ctx context.Context, msg *sdkws.MsgData) error {
	listener := s.listener()
	if listener == nil {
		return nil
	}
	var req rtc.SignalOnRoomParticipantConnectedReq
	if err := proto.Unmarshal(msg.Content, &req); err != nil {
		return err
	}
	log.ZDebug(ctx, "OnRoomParticipantConnected", "req", &req)
	listener.OnRoomParticipantConnected(jsonutil.StructToJsonString(&req))
	return nil
}

func (s *Signaling) handleRoomParticipantDisconnected(ctx context.Context, msg *sdkws.MsgData) error {
	listener := s.listener()
	if listener == nil {
		return nil
	}
	var req rtc.SignalOnRoomParticipantDisconnectedReq
	if err := proto.Unmarshal(msg.Content, &req); err != nil {
		return err
	}
	log.ZDebug(ctx, "OnRoomParticipantDisconnected", "req", &req)
	listener.OnRoomParticipantDisconnected(jsonutil.StructToJsonString(&req))
	return nil
}

// persistLocalCallRecord 异步投递写任务，避免阻塞信令通知路径。
func (s *Signaling) persistLocalCallRecord(
	ctx context.Context,
	inv *rtc.InvitationInfo,
	participant *rtc.ParticipantMetaData,
	status int32,
	direction int32,
	inviteMs, connectMs, endMs int64,
) {
	if s.db == nil || inv == nil || s.recordCh == nil {
		return
	}
	select {
	case s.recordCh <- recordTask{
		inv:         inv,
		participant: participant,
		status:      status,
		direction:   direction,
		inviteMs:    inviteMs,
		connectMs:   connectMs,
		endMs:       endMs,
	}:
	case <-s.done:
		return
	default:
		log.ZWarn(ctx, "recordCh is full, drop call record", nil, "roomID", inv.RoomID)
	}
}

// ── 昵称解析辅助（带好友表兜底） ─────────────────────────────────────────────

func (s *Signaling) composeCalleeMatchText(ctx context.Context, inv *rtc.InvitationInfo, participant *rtc.ParticipantMetaData) string {
	base := buildCalleeMatchTextFromProto(inv, participant)
	return s.appendCalleeNicknamesFromFriends(ctx, inv, base)
}

func (s *Signaling) appendCalleeNicknamesFromFriends(ctx context.Context, inv *rtc.InvitationInfo, base string) string {
	if s.db == nil || inv == nil || len(inv.InviteeUserIDList) == 0 {
		return base
	}
	friends, err := s.db.GetFriendInfoList(ctx, inv.InviteeUserIDList)
	if err != nil || len(friends) == 0 {
		return base
	}
	var extras []string
	for _, f := range friends {
		name := f.ConversationShowName()
		if name != "" {
			extras = append(extras, name)
		}
		if !isSingleChatCall(inv) {
			if f.Nickname != "" {
				extras = append(extras, f.Nickname)
			}
			if f.Remark != "" {
				extras = append(extras, f.Remark)
			}
		}
	}
	if len(extras) == 0 {
		return base
	}
	return strings.TrimSpace(base + " " + strings.Join(extras, " "))
}

func isSingleChatCall(inv *rtc.InvitationInfo) bool {
	return inv != nil && inv.SessionType == constant.SingleChatType
}

// resolve1v1UserDisplayName 单聊展示名：好友 remark > firstName+lastName > nickname；非好友仅信令 nickname。
func (s *Signaling) resolve1v1UserDisplayName(ctx context.Context, userID string, p *rtc.ParticipantMetaData) string {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ""
	}
	if s.db != nil {
		friends, err := s.db.GetFriendInfoList(ctx, []string{userID})
		if err == nil && len(friends) > 0 {
			if name := friends[0].ConversationShowName(); name != "" {
				return name
			}
		}
	}
	return nicknameFromParticipant(userID, p)
}

func (s *Signaling) resolveInviteeNickname(ctx context.Context, inv *rtc.InvitationInfo, p *rtc.ParticipantMetaData) string {
	if inv == nil || len(inv.InviteeUserIDList) == 0 {
		return ""
	}
	firstUID := inv.InviteeUserIDList[0]
	if isSingleChatCall(inv) {
		return s.resolve1v1UserDisplayName(ctx, firstUID, p)
	}
	if nick := extractInviteeNickname(inv, p); nick != "" {
		return nick
	}
	return s.resolve1v1UserDisplayName(ctx, firstUID, p)
}

func (s *Signaling) resolveInviterNickname(ctx context.Context, inv *rtc.InvitationInfo, p *rtc.ParticipantMetaData) string {
	if inv == nil || inv.InviterUserID == "" {
		return ""
	}
	if isSingleChatCall(inv) {
		return s.resolve1v1UserDisplayName(ctx, inv.InviterUserID, p)
	}
	if nick := extractInviterNickname(inv, p); nick != "" {
		return nick
	}
	return s.resolve1v1UserDisplayName(ctx, inv.InviterUserID, p)
}

func (s *Signaling) resolveInviterFaceURL(ctx context.Context, inv *rtc.InvitationInfo, p *rtc.ParticipantMetaData) string {
	if faceURL := extractInviterFaceURL(inv, p); faceURL != "" {
		return faceURL
	}
	if s.db == nil || inv == nil || inv.InviterUserID == "" {
		return ""
	}
	friends, err := s.db.GetFriendInfoList(ctx, []string{inv.InviterUserID})
	if err != nil || len(friends) == 0 {
		return ""
	}
	return friends[0].FaceURL
}

func (s *Signaling) resolveGroupName(ctx context.Context, inv *rtc.InvitationInfo, p *rtc.ParticipantMetaData) string {
	if name := extractGroupName(p); name != "" {
		return name
	}
	if s.db == nil || inv == nil || inv.GroupID == "" {
		return ""
	}
	g, err := s.db.GetGroupInfoByGroupID(ctx, inv.GroupID)
	if err != nil || g == nil {
		return ""
	}
	return g.GroupName
}
