package signaling

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/openimsdk/openim-sdk-core/v3/internal/interaction"
	"github.com/openimsdk/openim-sdk-core/v3/open_im_sdk_callback"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/ccontext"
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
// mu 保护 connectMs/accepted 的并发写入（inviteMs/createdAt 写入后只读，无需保护）。
type roomTiming struct {
	mu        sync.Mutex
	inviteMs  int64     // 本端发起或收到邀请的毫秒时间戳（写入后只读）
	connectMs int64     // 接通时的毫秒时间戳（0 = 尚未接通）
	accepted  bool      // 本端已执行 Accept 或对端接听通知已到达（主叫侧）
	createdAt time.Time // 条目创建时间，用于定期清理（写入后只读）
}

// recordTask 异步写记录任务，包含构建 LocalSignalCallRecord 所需的全部信息。
type recordTask struct {
	inv          *rtc.InvitationInfo
	participant  *rtc.ParticipantMetaData
	status       int32
	direction    int32
	action       string // constant.SignalCallAction*
	inviteMs     int64  // 从 roomTimings 取得
	connectMs    int64  // 从 roomTimings 取得（0 = 未接通）
	endMs        int64  // 通话结束时间（调用处捕获，避免异步延迟误差）
	callDuration int64  // 客户端主动上报的通话时长（毫秒）；0 = 由时间戳推算
}

// inviteTimer 本端收到/发起邀请后的超时定时器
type inviteTimer struct {
	timer *time.Timer
	inv   *rtc.InvitationInfo
}

type Signaling struct {
	loginUserID  string
	platformID   int32
	longConnMgr  *interaction.LongConnMgr
	db           db_interface.DataBase
	globalConfig *ccontext.GlobalConfig
	listener     func() open_im_sdk_callback.OnSignalingListener

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

	// syncCallRecordsUserProfile 在落库后按最新好友/用户资料刷新通话记录展示名（避免信令快照覆盖）。
	syncCallRecordsUserProfile func(ctx context.Context, userIDs ...string)

	// cleanupDone 用于通知定时清理协程退出
	cleanupDone chan struct{}
}

func NewSignaling(longConnMgr *interaction.LongConnMgr, loginUserID string, platformID int32, db db_interface.DataBase, globalConfig *ccontext.GlobalConfig) *Signaling {
	s := &Signaling{
		loginUserID:  loginUserID,
		platformID:   platformID,
		longConnMgr:  longConnMgr,
		db:           db,
		globalConfig: globalConfig,
		recordCh:     make(chan recordTask, 32),
		done:         make(chan struct{}),
		detailCache:  newDetailCache(),
		cleanupDone:  make(chan struct{}),
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
		// 无论正常退出还是 panic 恢复，都必须关闭 done，
		// 否则 Close() 中的 <-s.done 会永久阻塞。
		close(s.done)
	}()

	for task := range s.recordCh {
		if s.db == nil || task.inv == nil {
			continue
		}
		ctx := s.backgroundCtx()

		calleeText := s.composeCalleeMatchText(ctx, task.inv, task.participant)
		inviteeNickname := s.resolveInviteeNickname(ctx, task.inv, task.participant)
		inviteeUID := extractPrimaryInviteeUID(task.inv)
		inviteeFaceURL := s.resolveInviteeFaceURL(ctx, task.inv, task.participant)
		inviterNickname := s.resolveInviterNickname(ctx, task.inv, task.participant)
		inviterFaceURL := s.resolveInviterFaceURL(ctx, task.inv, task.participant)
		groupName := s.resolveGroupName(ctx, task.inv, task.participant)
		inviteeIDsJSON := buildInviteeIDsJSON(task.inv)

		cfg := callRecordConfig{
			inv:             task.inv,
			status:          task.status,
			direction:       task.direction,
			action:          task.action,
			inviteMs:        task.inviteMs,
			connectMs:       task.connectMs,
			endMs:           task.endMs,
			callDuration:    task.callDuration,
			calleeMatchText: calleeText,
			inviteeNickname: inviteeNickname,
			inviteeUID:      inviteeUID,
			inviteeFaceURL:  inviteeFaceURL,
			inviteeIDsJSON:  inviteeIDsJSON,
			inviterNickname: inviterNickname,
			inviterFaceURL:  inviterFaceURL,
			groupName:       groupName,
			ownerUserID:     s.loginUserID,
		}
		lr := newLocalSignalCallRecord(cfg)
		if lr == nil {
			continue
		}

		log.ZDebug(ctx, "recordWorker syncCallRecordsUserProfile", "lr", lr)

		if err := s.db.BatchUpsertSignalCallRecords(ctx, []*model_struct.LocalSignalCallRecord{lr}); err != nil {
			log.ZWarn(ctx, "async persistLocalCallRecord failed", err, "sID", lr.SID)
			continue
		}

		log.ZDebug(ctx, "recordWorker syncCallRecordsUserProfile", "lr", lr)

		if s.syncCallRecordsUserProfile != nil {
			if uids := peerParticipantUserIDs(s.loginUserID, task.inv); len(uids) > 0 {
				log.ZDebug(ctx, "recordWorker syncCallRecordsUserProfile", "uids", uids)
				s.syncCallRecordsUserProfile(ctx, uids...)
			}
		} else {
			log.ZDebug(ctx, "recordWorker syncCallRecordsUserProfile is nil")
		}
	}
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

func (s *Signaling) SetSyncCallRecordsUserProfile(fn func(ctx context.Context, userIDs ...string)) {
	s.syncCallRecordsUserProfile = fn
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
		t := v.(*roomTiming)
		t.mu.Lock()
		t.connectMs = ms
		t.accepted = true
		t.mu.Unlock()
	} else {
		s.roomTimings.Store(roomID, &roomTiming{connectMs: ms, accepted: true, createdAt: time.Now()})
	}
}

// peekTiming 读取 roomID 对应的时间戳（不删除）。
func (s *Signaling) peekTiming(roomID string) (inviteMs, connectMs int64) {
	if v, ok := s.roomTimings.Load(roomID); ok {
		t := v.(*roomTiming)
		t.mu.Lock()
		inviteMs, connectMs = t.inviteMs, t.connectMs
		t.mu.Unlock()
	}
	return
}

// peekTimingForRecord 读取 roomID 对应时间戳（不删除），供 Cancel 等事件判断是否已接听。
func (s *Signaling) peekTimingForRecord(roomID string) (inviteMs, connectMs int64, accepted bool, ok bool) {
	if v, ok2 := s.roomTimings.Load(roomID); ok2 {
		t := v.(*roomTiming)
		t.mu.Lock()
		inviteMs, connectMs, accepted = t.inviteMs, t.connectMs, t.accepted
		t.mu.Unlock()
		ok = inviteMs > 0 || connectMs > 0
	}
	return
}

// popTiming 读取并删除 roomID 对应的时间戳记录；若不存在则返回零值。
func (s *Signaling) popTiming(roomID string) (inviteMs, connectMs int64, accepted bool) {
	if v, ok := s.roomTimings.LoadAndDelete(roomID); ok {
		t := v.(*roomTiming)
		t.mu.Lock()
		inviteMs, connectMs, accepted = t.inviteMs, t.connectMs, t.accepted
		t.mu.Unlock()
		return
	}
	return 0, 0, false
}

// popTimingForRecord 消费时间戳并判断是否可以落库；已被其它终结事件消费时返回 ok=false。
func (s *Signaling) popTimingForRecord(roomID string) (inviteMs, connectMs int64, accepted bool, ok bool) {
	inviteMs, connectMs, accepted = s.popTiming(roomID)
	ok = inviteMs > 0 || connectMs > 0
	return
}

func callWasConnected(connectMs int64) bool {
	return connectMs > 0
}

// callRecordStatusFromTiming 仅在本端已接听且存在接通时间时记为已接听，避免超时/取消后 HungUp 误记为已接。
func callRecordStatusFromTiming(connectMs int64, accepted bool) int32 {
	if accepted && callWasConnected(connectMs) {
		return constant.SignalCallStatusAnswered
	}
	return constant.SignalCallStatusNotConnected
}

// resolveCallRecordDirection 根据本端在邀请中的角色与通话状态计算 direction。
// 主叫：Outgoing；被叫且已接通：Incoming；被叫且未接通：Missed。
func (s *Signaling) resolveCallRecordDirection(inv *rtc.InvitationInfo, status int32) int32 {
	if inv == nil {
		return constant.SignalCallDirectionUnknown
	}
	if inv.InviterUserID == s.loginUserID {
		return constant.SignalCallDirectionOutgoing
	}
	if status == constant.SignalCallStatusAnswered {
		return constant.SignalCallDirectionIncoming
	}
	return constant.SignalCallDirectionMissed
}

// ── 邀请超时定时器 ────────────────────────────────────────────────────────────

// startInviteTimer 启动本端邀请超时定时器。超时后触发 OnInvitationTimeout 回调并写本地记录。
func (s *Signaling) startInviteTimer(inv *rtc.InvitationInfo) {
	if inv == nil || inv.RoomID == "" || inv.Timeout <= 0 {
		return
	}

	timeout := time.Duration(inv.Timeout) * time.Second
	it := &inviteTimer{inv: inv}
	it.timer = time.AfterFunc(timeout, func() {
		s.inviteTimers.Delete(inv.RoomID)
		s.onInvitationTimeout(inv)
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

// backgroundCtx 为定时器/异步协程提供带 GlobalConfig 的 context。
// 注意：当 globalConfig 为 nil 时返回裸 context.Background()，调用方在执行需要
// GlobalConfig（如 signalingRequest/ccontext.Info）的操作前必须先检查 globalConfig。
func (s *Signaling) backgroundCtx() context.Context {
	if s.globalConfig == nil {
		return context.Background()
	}
	ctx := ccontext.WithInfo(context.Background(), s.globalConfig)
	return ccontext.WithOperationID(ctx, fmt.Sprintf("signaling_timer_%d", time.Now().UnixNano()))
}

// onInvitationTimeout 本端邀请超时：主叫侧上报服务端；被叫侧仅本地回调与落库。
func (s *Signaling) onInvitationTimeout(inv *rtc.InvitationInfo) {
	if inv == nil {
		return
	}
	ctx := s.backgroundCtx()
	if inv.InviterUserID == s.loginUserID {
		// globalConfig 为 nil 时无法构造认证 context，跳过需要网络请求的 Timeout 调用，
		// 避免 ccontext.Info panic（interface {} is nil, not *ccontext.GlobalConfig）。
		if s.globalConfig == nil {
			log.ZWarn(ctx, "onInvitationTimeout: globalConfig is nil, skipping server Timeout notify", nil, "roomID", inv.RoomID)
			return
		}
		if err := s.Timeout(ctx, &rtc.SignalTimeoutReq{Invitation: inv}); err != nil {
			log.ZWarn(ctx, "Timeout request failed on local timer", err, "roomID", inv.RoomID)
		}
		return
	}
	s.notifyInvitationTimeout(ctx, inv)
}

// notifyInvitationTimeout 被叫侧超时未接：通知 UI 并写本地未接通记录。
func (s *Signaling) notifyInvitationTimeout(ctx context.Context, inv *rtc.InvitationInfo) {
	s.cancelInviteTimer(inv.RoomID)

	listener := s.listener()
	if listener != nil {
		log.ZDebug(ctx, "OnInvitationTimeout", "roomID", inv.RoomID)
		listener.OnInvitationTimeout(jsonutil.StructToJsonString(inv))
	}

	inviteMs, _, _, ok := s.popTimingForRecord(inv.RoomID)
	if !ok {
		return
	}
	s.persistLocalCallRecord(ctx, inv, nil,
		constant.SignalCallStatusNotConnected,
		constant.SignalCallActionTimeout,
		inviteMs, 0, time.Now().UnixMilli(), 0)
	log.ZInfo(ctx, "persistLocalCallRecord", "Invitation", inv, "status", constant.SignalCallStatusNotConnected, "action", constant.SignalCallActionTimeout)

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
	case pConstant.CustomSignalNotification:
		return s.handleCustomSignalNotification(ctx, msg)
	case pConstant.StreamChangedNotification:
		log.ZDebug(ctx, "ignoring signaling notification", "contentType", msg.ContentType)
		return nil
	default:
		log.ZWarn(ctx, "unhandled signaling notification", nil, "contentType", msg.ContentType)
		return nil
	}
}

const (
	groupCallStatusPayloadType = "groupCallStatus"
	groupCallStatusStarted     = "started"
	groupCallStatusEnded       = "ended"
)

// groupCallStatusPayload mirrors the server broadcast in internal/rpc/rtc/signal.go.
type groupCallStatusPayload struct {
	Type          string `json:"type"`
	Status        string `json:"status"`
	GroupID       string `json:"groupID"`
	RoomID        string `json:"roomID"`
	MediaType     string `json:"mediaType"`
	InviterUserID string `json:"inviterUserID"`
}

// handleCustomSignalNotification handles CustomSignalNotification (1605).
// Currently supports group-wide "call in progress" banners for non-invited members.
func (s *Signaling) handleCustomSignalNotification(ctx context.Context, msg *sdkws.MsgData) error {
	var payload groupCallStatusPayload
	if err := jsonutil.JsonUnmarshal(msg.Content, &payload); err != nil {
		return err
	}
	if payload.Type != groupCallStatusPayloadType {
		log.ZDebug(ctx, "ignore unknown custom signal payload", "type", payload.Type)
		return nil
	}
	if payload.GroupID == "" || payload.RoomID == "" {
		log.ZWarn(ctx, "invalid groupCallStatus payload", nil, "payload", payload)
		return nil
	}
	switch payload.Status {
	case groupCallStatusStarted, groupCallStatusEnded:
	default:
		log.ZWarn(ctx, "unknown groupCallStatus", nil, "status", payload.Status)
		return nil
	}

	listener := s.listener()
	if listener == nil {
		log.ZWarn(ctx, "signaling listener is nil, skipping groupCallStatus", nil)
		return nil
	}

	log.ZDebug(ctx, "OnGroupCallStatusChanged", "payload", payload)
	listener.OnGroupCallStatusChanged(jsonutil.StructToJsonString(payload))
	return nil
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
		s.startInviteTimer(req.Invitation)
		log.ZInfo(ctx, "handleInvite startInviteTimer", "roomID", req.Invitation.RoomID)
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
		s.startInviteTimer(req.Invitation)

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

		inviteMs, _, _, ok := s.popTimingForRecord(req.Invitation.RoomID)
		if ok {
			s.persistLocalCallRecord(ctx, req.Invitation, req.Participant,
				constant.SignalCallStatusNotConnected,
				constant.SignalCallActionReject,
				inviteMs, 0, time.Now().UnixMilli(), 0)
			log.ZInfo(ctx, "persistLocalCallRecord", "Invitation", req.Invitation, "status", constant.SignalCallStatusNotConnected, "action", constant.SignalCallActionReject)
		}
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
		s.notifyInvitationTimeout(ctx, req.Invitation)
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

		// 主叫挂断时服务端可能同时推送 Cancel + HungUp；已接听则仅由 HungUp 落库，避免两条未接记录。
		if _, connectMs, accepted, _ := s.peekTimingForRecord(req.Invitation.RoomID); callRecordStatusFromTiming(connectMs, accepted) == constant.SignalCallStatusAnswered {
			return nil
		}
		inviteMs, _, _, ok := s.popTimingForRecord(req.Invitation.RoomID)
		if !ok {
			return nil
		}
		s.persistLocalCallRecord(ctx, req.Invitation, req.Participant,
			constant.SignalCallStatusNotConnected,
			constant.SignalCallActionCancel,
			inviteMs, 0, time.Now().UnixMilli(), 0)
		log.ZInfo(ctx, "persistLocalCallRecord", "Invitation", req.Invitation, "status", constant.SignalCallStatusNotConnected, "action", constant.SignalCallActionCancel)
	}
	return nil
}

// handleHungUp 收到对端挂断通知：取消超时定时器，写通话记录。
// 若 connectMs=0（对端在接听前挂断），记录为未接通；否则记录为已接听。
func (s *Signaling) handleHungUp(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalHungUpReq) error {
	if req.Invitation == nil {
		return nil
	}
	if req.UserID != s.loginUserID {
		s.cancelInviteTimer(req.Invitation.RoomID)

		log.ZInfo(ctx, "OnHangUp", "hungUp", req, "loginUserID", s.loginUserID)
		listener.OnHangUp(jsonutil.StructToJsonString(req))

		inviteMs, connectMs, accepted, ok := s.popTimingForRecord(req.Invitation.RoomID)
		if !ok {
			return nil
		}
		status := callRecordStatusFromTiming(connectMs, accepted)
		s.persistLocalCallRecord(ctx, req.Invitation, nil,
			status,
			constant.SignalCallActionHungUp,
			inviteMs, connectMs, time.Now().UnixMilli(), req.CallDuration)
		log.ZInfo(ctx, "persistLocalCallRecord", "Invitation", req.Invitation,
			"status", status, "action", constant.SignalCallActionHungUp, "callDuration", req.CallDuration)
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
// direction 由本端角色与 status 在内部统一计算，调用方勿再传入。
// callDuration > 0 时直接用作通话时长（毫秒），否则由时间戳推算。
// 群音视频通话不落本地通话记录。
func (s *Signaling) persistLocalCallRecord(
	ctx context.Context,
	inv *rtc.InvitationInfo,
	participant *rtc.ParticipantMetaData,
	status int32,
	action string,
	inviteMs, connectMs, endMs int64,
	callDuration int64,
) {
	if s.db == nil || inv == nil || s.recordCh == nil || isGroupChatCall(inv) {
		return
	}
	direction := s.resolveCallRecordDirection(inv, status)
	select {
	case s.recordCh <- recordTask{
		inv:          inv,
		participant:  participant,
		status:       status,
		direction:    direction,
		action:       action,
		inviteMs:     inviteMs,
		connectMs:    connectMs,
		endMs:        endMs,
		callDuration: callDuration,
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
	return s.appendParticipantSearchTokens(ctx, inv, base)
}

// appendParticipantSearchTokens 将主叫/被叫在好友表、用户表中的展示名写入 callee_match_text，便于按姓名检索。
func (s *Signaling) appendParticipantSearchTokens(ctx context.Context, inv *rtc.InvitationInfo, base string) string {
	if s.db == nil || inv == nil {
		return base
	}
	userIDs := make([]string, 0, len(inv.InviteeUserIDList)+1)
	if uid := strings.TrimSpace(inv.InviterUserID); uid != "" {
		userIDs = append(userIDs, uid)
	}
	for _, uid := range inv.InviteeUserIDList {
		if uid = strings.TrimSpace(uid); uid != "" {
			userIDs = append(userIDs, uid)
		}
	}
	if len(userIDs) == 0 {
		return base
	}

	friendByID := make(map[string]*model_struct.LocalFriend)
	friends, err := s.db.GetFriendInfoList(ctx, userIDs)
	if err == nil {
		for _, f := range friends {
			if f != nil && f.FriendUserID != "" {
				friendByID[f.FriendUserID] = f
			}
		}
	}

	var tokens []string
	for _, uid := range userIDs {
		if f := friendByID[uid]; f != nil {
			tokens = append(tokens, searchableFriendTokens(f)...)
			continue
		}
		user, err := s.db.GetLoginUser(ctx, uid)
		if err == nil && user != nil {
			tokens = append(tokens, searchableUserTokens(user)...)
		}
	}
	return appendUniqueSearchTokens(base, tokens)
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
		user, err := s.db.GetLoginUser(ctx, userID)
		if err == nil && user != nil {
			if name := user.DisplayName(); name != "" {
				return name
			}
		}
	}
	return nicknameFromParticipant(userID, p)
}

// resolveInviteeNickname 被叫展示名：群通话为群名；单聊为好友 remark > firstName+lastName > nickname。
func (s *Signaling) resolveInviteeNickname(ctx context.Context, inv *rtc.InvitationInfo, p *rtc.ParticipantMetaData) string {
	if isGroupChatCall(inv) {
		return s.resolveGroupName(ctx, inv, p)
	}
	uid := extractPrimaryInviteeUID(inv)
	if uid == "" {
		return ""
	}
	if name := s.resolve1v1UserDisplayName(ctx, uid, p); name != "" {
		return name
	}
	return extractInviteeNickname(inv, p)
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
	if s.db != nil && inv != nil && inv.InviterUserID != "" {
		friends, err := s.db.GetFriendInfoList(ctx, []string{inv.InviterUserID})
		if err == nil && len(friends) > 0 && friends[0].FaceURL != "" {
			return friends[0].FaceURL
		}
	}
	if faceURL := extractInviterFaceURL(inv, p); faceURL != "" {
		return faceURL
	}
	return ""
}

func (s *Signaling) resolveInviteeFaceURL(ctx context.Context, inv *rtc.InvitationInfo, p *rtc.ParticipantMetaData) string {
	if isGroupChatCall(inv) {
		return s.resolveGroupFaceURL(ctx, inv, p)
	}
	uid := extractPrimaryInviteeUID(inv)
	if s.db != nil && uid != "" {
		friends, err := s.db.GetFriendInfoList(ctx, []string{uid})
		if err == nil && len(friends) > 0 && friends[0].FaceURL != "" {
			return friends[0].FaceURL
		}
	}
	if faceURL := extractInviteeFaceURL(inv, p); faceURL != "" {
		return faceURL
	}
	return ""
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

func (s *Signaling) resolveGroupFaceURL(ctx context.Context, inv *rtc.InvitationInfo, p *rtc.ParticipantMetaData) string {
	if faceURL := extractGroupFaceURL(p); faceURL != "" {
		return faceURL
	}
	if s.db == nil || inv == nil || inv.GroupID == "" {
		return ""
	}
	g, err := s.db.GetGroupInfoByGroupID(ctx, inv.GroupID)
	if err != nil || g == nil {
		return ""
	}
	return g.FaceURL
}
