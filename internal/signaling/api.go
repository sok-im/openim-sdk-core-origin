package signaling

import (
	"context"

	"github.com/google/uuid"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/api"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/protocol/rtc"
	"github.com/openimsdk/tools/log"
	"github.com/openimsdk/tools/utils/jsonutil"
)

const defaultTimeout = 30

func (s *Signaling) signalingRequest(ctx context.Context, req *rtc.SignalReq) (*rtc.SignalResp, error) {
	var resp rtc.SignalResp
	if err := s.longConnMgr.SendReqWaitResp(ctx, req, constant.SendSignalMsg, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *Signaling) Invite(ctx context.Context, signalInviteReq *rtc.SignalInviteReq) (*rtc.SignalInviteResp, error) {
	s.fillInviteDefaults(signalInviteReq.Invitation)
	signalInviteReq.UserID = s.loginUserID

	req := &rtc.SignalReq{
		Payload: &rtc.SignalReq_Invite{
			Invite: signalInviteReq,
		},
	}
	resp, err := s.signalingRequest(ctx, req)
	if err != nil {
		log.ZError(ctx, "Invite failed", err)
		return nil, err
	}
	if inviteResp := resp.GetInvite(); inviteResp != nil {
		log.ZInfo(ctx, "Invite success", "liveURL", inviteResp.LiveURL, "roomID", inviteResp.RoomID)
		return inviteResp, nil
	}
	return &rtc.SignalInviteResp{}, nil
}

func (s *Signaling) InviteInGroup(ctx context.Context, signalInviteInGroupReq *rtc.SignalInviteInGroupReq) (*rtc.SignalInviteInGroupResp, error) {
	s.fillInviteDefaults(signalInviteInGroupReq.Invitation)
	signalInviteInGroupReq.UserID = s.loginUserID

	req := &rtc.SignalReq{
		Payload: &rtc.SignalReq_InviteInGroup{
			InviteInGroup: signalInviteInGroupReq,
		},
	}
	resp, err := s.signalingRequest(ctx, req)
	if err != nil {
		log.ZError(ctx, "InviteInGroup failed", err)
		return nil, err
	}
	if inviteResp := resp.GetInviteInGroup(); inviteResp != nil {
		log.ZInfo(ctx, "InviteInGroup success", "liveURL", inviteResp.LiveURL, "roomID", inviteResp.RoomID)
		return inviteResp, nil
	}
	return &rtc.SignalInviteInGroupResp{}, nil
}

func (s *Signaling) Accept(ctx context.Context, signalAcceptReq *rtc.SignalAcceptReq) (*rtc.SignalAcceptResp, error) {
	signalAcceptReq.UserID = s.loginUserID
	signalAcceptReq.OpUserPlatformID = s.platformID

	req := &rtc.SignalReq{
		Payload: &rtc.SignalReq_Accept{
			Accept: signalAcceptReq,
		},
	}
	resp, err := s.signalingRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	if acceptResp := resp.GetAccept(); acceptResp != nil {
		return acceptResp, nil
	}
	return &rtc.SignalAcceptResp{}, nil
}

func (s *Signaling) Reject(ctx context.Context, signalRejectReq *rtc.SignalRejectReq) error {
	signalRejectReq.UserID = s.loginUserID
	signalRejectReq.OpUserPlatformID = s.platformID

	req := &rtc.SignalReq{
		Payload: &rtc.SignalReq_Reject{
			Reject: signalRejectReq,
		},
	}
	_, err := s.signalingRequest(ctx, req)
	return err
}

func (s *Signaling) Cancel(ctx context.Context, signalCancelReq *rtc.SignalCancelReq) error {
	signalCancelReq.UserID = s.loginUserID

	req := &rtc.SignalReq{
		Payload: &rtc.SignalReq_Cancel{
			Cancel: signalCancelReq,
		},
	}
	_, err := s.signalingRequest(ctx, req)
	return err
}

func (s *Signaling) HungUp(ctx context.Context, signalHungUpReq *rtc.SignalHungUpReq) error {
	signalHungUpReq.UserID = s.loginUserID

	req := &rtc.SignalReq{
		Payload: &rtc.SignalReq_HungUp{
			HungUp: signalHungUpReq,
		},
	}
	_, err := s.signalingRequest(ctx, req)
	return err
}

func (s *Signaling) GetTokenByRoomID(ctx context.Context, signalGetTokenReq *rtc.SignalGetTokenByRoomIDReq) (*rtc.SignalGetTokenByRoomIDResp, error) {
	signalGetTokenReq.UserID = s.loginUserID

	req := &rtc.SignalReq{
		Payload: &rtc.SignalReq_GetTokenByRoomID{
			GetTokenByRoomID: signalGetTokenReq,
		},
	}
	resp, err := s.signalingRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	if tokenResp := resp.GetGetTokenByRoomID(); tokenResp != nil {
		return tokenResp, nil
	}
	return &rtc.SignalGetTokenByRoomIDResp{}, nil
}

func (s *Signaling) GetRoomByGroupID(ctx context.Context, req *rtc.SignalGetRoomByGroupIDReq) (*rtc.SignalGetRoomByGroupIDResp, error) {
	return api.SignalGetRoomByGroupID.Invoke(ctx, req)
}

func (s *Signaling) GetSignalInvitationInfoStartApp(ctx context.Context, req *rtc.GetSignalInvitationInfoStartAppReq) (*rtc.GetSignalInvitationInfoStartAppResp, error) {
	req.UserID = s.loginUserID
	return api.GetSignalInvitationInfoStartApp.Invoke(ctx, req)
}

func (s *Signaling) GetSignalInvitationRecords(ctx context.Context, req *rtc.GetSignalInvitationRecordsReq) (*rtc.GetSignalInvitationRecordsResp, error) {
	return api.GetSignalInvitationRecords.Invoke(ctx, req)
}

func (s *Signaling) SendCustomSignal(ctx context.Context, req *rtc.SignalSendCustomSignalReq) error {
	return api.SignalSendCustomSignal.Execute(ctx, req)
}

func (s *Signaling) fillInviteDefaults(invitation *rtc.InvitationInfo) {
	if invitation == nil {
		return
	}
	invitation.InviterUserID = s.loginUserID
	invitation.PlatformID = s.platformID
	if invitation.RoomID == "" {
		invitation.RoomID = "room-" + uuid.New().String()
	}
	if invitation.Timeout == 0 {
		invitation.Timeout = defaultTimeout
	}
}

func InviteReqToJson(req *rtc.SignalInviteReq) string {
	return jsonutil.StructToJsonString(req)
}
