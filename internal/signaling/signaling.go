package signaling

import (
	"context"
	"fmt"

	"github.com/openimsdk/openim-sdk-core/v3/internal/interaction"
	"github.com/openimsdk/openim-sdk-core/v3/open_im_sdk_callback"
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
	listener    func() open_im_sdk_callback.OnSignalingListener
}

func NewSignaling(longConnMgr *interaction.LongConnMgr, loginUserID string, platformID int32) *Signaling {
	return &Signaling{
		loginUserID: loginUserID,
		platformID:  platformID,
		longConnMgr: longConnMgr,
	}
}

func (s *Signaling) SetListener(listener func() open_im_sdk_callback.OnSignalingListener) {
	s.listener = listener
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
	}
	return nil
}

func (s *Signaling) handleHungUp(ctx context.Context, listener open_im_sdk_callback.OnSignalingListener, req *rtc.SignalHungUpReq) error {
	if req.UserID != s.loginUserID {
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
