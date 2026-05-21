package signaling

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/openimsdk/openim-sdk-core/v3/internal/interaction"
	"github.com/openimsdk/openim-sdk-core/v3/open_im_sdk_callback"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/db_interface"
	pConstant "github.com/openimsdk/protocol/constant"
	"github.com/openimsdk/protocol/rtc"
	"github.com/openimsdk/protocol/sdkws"
	"github.com/openimsdk/tools/log"
	"github.com/openimsdk/tools/utils/datautil"
	"github.com/openimsdk/tools/utils/jsonutil"
	"google.golang.org/protobuf/proto"
)

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

	// inviteTimers 本端邀请超时定时器，key: roomID
	inviteTimers sync.Map

	// detailCache 为 GetLocalSignalCallRecordDetail 提供 LRU + TTL=5min 缓存
	detailCache *detailCache
}

func NewSignaling(longConnMgr *interaction.LongConnMgr, loginUserID string, platformID int32, db db_interface.DataBase) *Signaling {
	return &Signaling{
		loginUserID: loginUserID,
		platformID:  platformID,
		longConnMgr: longConnMgr,
		db:          db,
		detailCache: newDetailCache(),
	}
}

func (s *Signaling) SetListener(listener func() open_im_sdk_callback.OnSignalingListener) {
	s.listener = listener
}

// Close 清理定时器与缓存。
func (s *Signaling) Close() {
	s.inviteTimers.Range(func(key, value any) bool {
		it := value.(*inviteTimer)
		it.timer.Stop()
		s.inviteTimers.Delete(key)
		return true
	})
	if s.detailCache != nil {
		s.detailCache.Clear()
	}
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

// onInvitationTimeout 邀请超时回调：通知 UI。
func (s *Signaling) onInvitationTimeout(inv *rtc.InvitationInfo, direction int32) {
	ctx := context.Background()
	listener := s.listener()
	if listener != nil {
		log.ZDebug(ctx, "OnInvitationTimeout", "roomID", inv.RoomID, "direction", direction)
		listener.OnInvitationTimeout(jsonutil.StructToJsonString(inv))
	}
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
	default:
		log.ZWarn(ctx, "unhandled signaling payload type", nil, "type", fmt.Sprintf("%T", signalReq.Payload))
		return nil
	}
}

// handleInvite 被叫侧收到来电邀请：启动超时定时器，触发 UI 回调。
func (s *Signaling) handleInvite(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalInviteReq) error {
	if req.Invitation == nil {
		return nil
	}
	if datautil.Contain(s.loginUserID, req.Invitation.InviteeUserIDList...) {
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
		s.startInviteTimer(req.Invitation, constant.SignalCallDirectionMissed)
		log.ZDebug(ctx, "OnReceiveNewInvitation (group)", "invitation", req)
		listener.OnReceiveNewInvitation(jsonutil.StructToJsonString(req))
	}
	return nil
}

// handleAccept 主叫侧收到被叫接听通知：取消超时定时器。
func (s *Signaling) handleAccept(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalAcceptReq) error {
	if req.Invitation == nil {
		return nil
	}
	if req.Invitation.InviterUserID == s.loginUserID {
		s.cancelInviteTimer(req.Invitation.RoomID)

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

// handleReject 主叫侧收到被叫拒接通知：取消超时定时器。
func (s *Signaling) handleReject(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalRejectReq) error {
	if req.Invitation == nil {
		return nil
	}
	if req.Invitation.InviterUserID == s.loginUserID {
		s.cancelInviteTimer(req.Invitation.RoomID)

		log.ZDebug(ctx, "OnInviteeRejected", "reject", req)
		listener.OnInviteeRejected(jsonutil.StructToJsonString(req))
		return nil
	}
	if req.UserID == s.loginUserID && req.OpUserPlatformID != s.platformID {
		log.ZDebug(ctx, "OnInviteeRejectedByOtherDevice", "reject", req)
		listener.OnInviteeRejectedByOtherDevice(jsonutil.StructToJsonString(req))
	}
	return nil
}

// handleCancel 被叫侧收到主叫取消通知：取消超时定时器。
func (s *Signaling) handleCancel(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalCancelReq) error {
	if req.Invitation == nil {
		return nil
	}
	if datautil.Contain(s.loginUserID, req.Invitation.InviteeUserIDList...) {
		s.cancelInviteTimer(req.Invitation.RoomID)

		log.ZDebug(ctx, "OnInvitationCancelled", "cancel", req)
		listener.OnInvitationCancelled(jsonutil.StructToJsonString(req))
	}
	return nil
}

// handleHungUp 收到对端挂断通知：取消超时定时器。
func (s *Signaling) handleHungUp(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalHungUpReq) error {
	if req.Invitation == nil {
		return nil
	}
	if req.UserID != s.loginUserID {
		s.cancelInviteTimer(req.Invitation.RoomID)
		log.ZDebug(ctx, "OnHangUp", "hungUp", req)
		listener.OnHangUp(jsonutil.StructToJsonString(req))
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

