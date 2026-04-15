package signaling

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openimsdk/openim-sdk-core/v3/internal/interaction"
	"github.com/openimsdk/openim-sdk-core/v3/open_im_sdk_callback"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/db_interface"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/sdkerrs"
	pConstant "github.com/openimsdk/protocol/constant"
	"github.com/openimsdk/protocol/rtc"
	"github.com/openimsdk/protocol/sdkws"
	"github.com/openimsdk/tools/log"
	"github.com/openimsdk/tools/utils/datautil"
	"github.com/openimsdk/tools/utils/jsonutil"
	"google.golang.org/protobuf/proto"
)

type Signaling struct {
	loginUserID string
	platformID  int32
	longConnMgr *interaction.LongConnMgr
	db          db_interface.DataBase
	listener    func() open_im_sdk_callback.OnSignalingListener

	// recordCh 用于异步落库通话记录，避免阻塞信令通知路径
	recordCh chan recordTask
	done     chan struct{}

	// detailCache 为 GetLocalSignalCallRecordDetail 提供 LRU + TTL=5min 缓存
	detailCache *detailCache
}

// recordTask 异步写记录任务
type recordTask struct {
	inv         *rtc.InvitationInfo
	participant *rtc.ParticipantMetaData
	dialStatus  int32
}

func NewSignaling(longConnMgr *interaction.LongConnMgr, loginUserID string, platformID int32, db db_interface.DataBase) *Signaling {
	s := &Signaling{
		loginUserID: loginUserID,
		platformID:  platformID,
		longConnMgr: longConnMgr,
		db:          db,
		recordCh:    make(chan recordTask, 32), // 缓冲32条，避免瞬时高峰阻塞
		done:        make(chan struct{}),
		detailCache: newDetailCache(),
	}

	// 启动异步写记录协程，避免阻塞信令通知路径
	go s.recordWorker()

	return s
}

// recordWorker 异步处理通话记录入库
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
		lr := newLocalSignalCallRecord(task.inv, time.Now().UnixMilli(), task.dialStatus, calleeText)
		if lr == nil {
			continue
		}
		if err := s.db.BatchUpsertSignalCallRecords(ctx, []*model_struct.LocalSignalCallRecord{lr}); err != nil {
			log.ZWarn(ctx, "async persistLocalCallRecord failed", err, "sID", lr.SID)
		}
	}
	close(s.done)
}

func (s *Signaling) SetListener(listener func() open_im_sdk_callback.OnSignalingListener) {
	s.listener = listener
}

// Close 清理异步 worker
func (s *Signaling) Close() {
	if s.recordCh != nil {
		close(s.recordCh)
	}
	if s.detailCache != nil {
		s.detailCache.Clear()
	}
	<-s.done // 等待 worker 退出
}

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
	default:
	}
	log.ZError(ctx, "unhandled signaling notification", nil, "contentType", msg.ContentType)
	return sdkerrs.New(pConstant.SignalingNotificationEnd, "unhandled signaling notification", fmt.Sprintf("contentType: %v", msg.ContentType)).Wrap()
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
	default:
		log.ZError(ctx, "unhandled signaling payload type", nil, "type", signalReq.Payload)
	}
	return sdkerrs.New(pConstant.SignalingNotificationEnd, "unhandled signaling payload type", fmt.Sprintf("type: %T", signalReq.Payload)).Wrap()
}

func (s *Signaling) handleInvite(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalInviteReq) error {
	if req.Invitation == nil {
		return nil
	}
	if datautil.Contain(s.loginUserID, req.Invitation.InviteeUserIDList...) {
		log.ZDebug(ctx, "OnReceiveNewInvitation", "invitation", req)
		listener.OnReceiveNewInvitation(jsonutil.StructToJsonString(req))
	}
	return nil
}

func (s *Signaling) handleInviteInGroup(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalInviteInGroupReq) error {
	if req.Invitation == nil {
		return nil
	}
	if datautil.Contain(s.loginUserID, req.Invitation.InviteeUserIDList...) {
		log.ZDebug(ctx, "OnReceiveNewInvitation (group)", "invitation", req)
		listener.OnReceiveNewInvitation(jsonutil.StructToJsonString(req))
	}
	return nil
}

func (s *Signaling) handleAccept(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalAcceptReq) error {
	if req.Invitation == nil {
		return nil
	}
	if req.Invitation.InviterUserID == s.loginUserID {
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

func (s *Signaling) handleReject(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalRejectReq) error {
	if req.Invitation == nil {
		return nil
	}
	if req.Invitation.InviterUserID == s.loginUserID {
		log.ZDebug(ctx, "OnInviteeRejected", "reject", req)
		listener.OnInviteeRejected(jsonutil.StructToJsonString(req))
		s.persistLocalCallRecord(ctx, req.Invitation, req.Participant, constant.SignalCallDialStatusNotConnected)
		return nil
	}
	if req.UserID == s.loginUserID && req.OpUserPlatformID != s.platformID {
		log.ZDebug(ctx, "OnInviteeRejectedByOtherDevice", "reject", req)
		listener.OnInviteeRejectedByOtherDevice(jsonutil.StructToJsonString(req))
	}
	return nil
}

func (s *Signaling) handleCancel(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalCancelReq) error {
	if req.Invitation == nil {
		return nil
	}
	if datautil.Contain(s.loginUserID, req.Invitation.InviteeUserIDList...) {
		log.ZDebug(ctx, "OnInvitationCancelled", "cancel", req)
		listener.OnInvitationCancelled(jsonutil.StructToJsonString(req))
		s.persistLocalCallRecord(ctx, req.Invitation, req.Participant, constant.SignalCallDialStatusNotConnected)
	}
	return nil
}

func (s *Signaling) handleHungUp(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalHungUpReq) error {
	if req.Invitation == nil {
		return nil
	}
	if req.UserID != s.loginUserID {
		log.ZDebug(ctx, "OnHangUp", "hungUp", req)
		listener.OnHangUp(jsonutil.StructToJsonString(req))
		s.persistLocalCallRecord(ctx, req.Invitation, nil, constant.SignalCallDialStatusConnected)
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

// persistLocalCallRecord 异步投递写任务，避免阻塞信令通知路径
func (s *Signaling) persistLocalCallRecord(ctx context.Context, inv *rtc.InvitationInfo, participant *rtc.ParticipantMetaData, dialStatus int32) {
	if s.db == nil || inv == nil || s.recordCh == nil {
		return
	}
	select {
	case s.recordCh <- recordTask{
		inv:         inv,
		participant: participant,
		dialStatus:  dialStatus,
	}:
	case <-s.done:
		// worker 已退出
		return
	default:
		log.ZWarn(ctx, "recordCh is full, drop call record", nil, "roomID", inv.RoomID)
	}
}

// composeCalleeMatchText 被叫检索串：信令中的 invitee ID + participant 昵称 + 本地好友昵称/备注。
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
		if f.Nickname != "" {
			extras = append(extras, f.Nickname)
		}
		if f.Remark != "" {
			extras = append(extras, f.Remark)
		}
	}
	if len(extras) == 0 {
		return base
	}
	return strings.TrimSpace(base + " " + strings.Join(extras, " "))
}
