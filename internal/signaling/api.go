package signaling

import (
	"context"
	"strings"
	"time"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/api"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/sdkerrs"
	"github.com/openimsdk/openim-sdk-core/v3/sdk_struct"
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

// Invite 主叫侧发起邀请：记录拨打开始时间，启动超时定时器。
func (s *Signaling) Invite(ctx context.Context, signalInviteReq *rtc.SignalInviteReq) (*rtc.SignalInviteResp, error) {
	s.fillInviteDefaults(signalInviteReq.Invitation)
	signalInviteReq.UserID = s.loginUserID

	if signalInviteReq.Invitation != nil && signalInviteReq.Invitation.RoomID != "" {
		inviteMs := signalInviteReq.Invitation.InitiateTime
		if inviteMs <= 0 {
			inviteMs = time.Now().UnixMilli()
		}
		s.storeInviteTime(signalInviteReq.Invitation.RoomID, inviteMs)
	}

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

	// 邀请发送成功后启动超时定时器（主叫侧）
	if signalInviteReq.Invitation != nil {
		s.startInviteTimer(signalInviteReq.Invitation, constant.SignalCallDirectionOutgoing)
	}

	if inviteResp := resp.GetInvite(); inviteResp != nil {
		log.ZInfo(ctx, "Invite success", "liveURL", inviteResp.LiveURL, "roomID", inviteResp.RoomID)
		return inviteResp, nil
	}
	return &rtc.SignalInviteResp{}, nil
}

// InviteInGroup 主叫侧发起群组邀请：同样记录拨出时间并启动超时定时器。
func (s *Signaling) InviteInGroup(ctx context.Context, signalInviteInGroupReq *rtc.SignalInviteInGroupReq) (*rtc.SignalInviteInGroupResp, error) {
	s.fillInviteDefaults(signalInviteInGroupReq.Invitation)
	signalInviteInGroupReq.UserID = s.loginUserID

	if signalInviteInGroupReq.Invitation != nil && signalInviteInGroupReq.Invitation.RoomID != "" {
		inviteMs := signalInviteInGroupReq.Invitation.InitiateTime
		if inviteMs <= 0 {
			inviteMs = time.Now().UnixMilli()
		}
		s.storeInviteTime(signalInviteInGroupReq.Invitation.RoomID, inviteMs)
	}

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

	if signalInviteInGroupReq.Invitation != nil {
		s.startInviteTimer(signalInviteInGroupReq.Invitation, constant.SignalCallDirectionOutgoing)
	}

	if inviteResp := resp.GetInviteInGroup(); inviteResp != nil {
		log.ZInfo(ctx, "InviteInGroup success", "liveURL", inviteResp.LiveURL, "roomID", inviteResp.RoomID)
		return inviteResp, nil
	}
	return &rtc.SignalInviteInGroupResp{}, nil
}

// Accept 被叫侧接听：取消超时定时器，记录接通时间。
func (s *Signaling) Accept(ctx context.Context, signalAcceptReq *rtc.SignalAcceptReq) (*rtc.SignalAcceptResp, error) {
	signalAcceptReq.UserID = s.loginUserID
	signalAcceptReq.OpUserPlatformID = s.platformID

	if signalAcceptReq.Invitation != nil && signalAcceptReq.Invitation.RoomID != "" {
		s.cancelInviteTimer(signalAcceptReq.Invitation.RoomID)
		s.storeConnectTime(signalAcceptReq.Invitation.RoomID, time.Now().UnixMilli())
	}

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

// Reject 被叫侧拒接：取消超时定时器，写未拨通记录（incoming 方向）。
func (s *Signaling) Reject(ctx context.Context, signalRejectReq *rtc.SignalRejectReq) error {
	signalRejectReq.UserID = s.loginUserID
	signalRejectReq.OpUserPlatformID = s.platformID

	req := &rtc.SignalReq{
		Payload: &rtc.SignalReq_Reject{
			Reject: signalRejectReq,
		},
	}
	_, err := s.signalingRequest(ctx, req)
	if err != nil {
		return err
	}
	if signalRejectReq.Invitation != nil {
		s.cancelInviteTimer(signalRejectReq.Invitation.RoomID)
		inviteMs, connectMs := s.popTiming(signalRejectReq.Invitation.RoomID)
		s.persistLocalCallRecord(ctx,
			signalRejectReq.Invitation,
			signalRejectReq.Participant,
			constant.SignalCallDialStatusNotConnected,
			constant.SignalCallDirectionIncoming,
			inviteMs, connectMs, time.Now().UnixMilli())
	}
	return nil
}

// Cancel 主叫侧取消：取消超时定时器，写未拨通记录（outgoing 方向）。
func (s *Signaling) Cancel(ctx context.Context, signalCancelReq *rtc.SignalCancelReq) error {
	signalCancelReq.UserID = s.loginUserID

	req := &rtc.SignalReq{
		Payload: &rtc.SignalReq_Cancel{
			Cancel: signalCancelReq,
		},
	}
	_, err := s.signalingRequest(ctx, req)
	if err != nil {
		return err
	}
	if signalCancelReq.Invitation != nil {
		s.cancelInviteTimer(signalCancelReq.Invitation.RoomID)
		inviteMs, connectMs := s.popTiming(signalCancelReq.Invitation.RoomID)
		s.persistLocalCallRecord(ctx,
			signalCancelReq.Invitation,
			signalCancelReq.Participant,
			constant.SignalCallDialStatusNotConnected,
			constant.SignalCallDirectionOutgoing,
			inviteMs, connectMs, time.Now().UnixMilli())
	}
	return nil
}

// HungUp 任意一方挂断：取消超时定时器，写已拨通记录。
func (s *Signaling) HungUp(ctx context.Context, signalHungUpReq *rtc.SignalHungUpReq) error {
	signalHungUpReq.UserID = s.loginUserID

	req := &rtc.SignalReq{
		Payload: &rtc.SignalReq_HungUp{
			HungUp: signalHungUpReq,
		},
	}
	_, err := s.signalingRequest(ctx, req)
	if err != nil {
		return err
	}
	if signalHungUpReq.Invitation != nil {
		s.cancelInviteTimer(signalHungUpReq.Invitation.RoomID)
		direction := constant.SignalCallDirectionOutgoing
		if signalHungUpReq.Invitation.InviterUserID != s.loginUserID {
			direction = constant.SignalCallDirectionIncoming
		}
		inviteMs, connectMs := s.popTiming(signalHungUpReq.Invitation.RoomID)
		s.persistLocalCallRecord(ctx,
			signalHungUpReq.Invitation,
			nil,
			constant.SignalCallDialStatusConnected,
			direction,
			inviteMs, connectMs, time.Now().UnixMilli())
	}
	return nil
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

// GetSignalInvitationRecords 调用服务端历史接口（与本地通话记录独立）。
func (s *Signaling) GetSignalInvitationRecords(ctx context.Context, req *rtc.GetSignalInvitationRecordsReq) (*rtc.GetSignalInvitationRecordsResp, error) {
	return api.GetSignalInvitationRecords.Invoke(ctx, req)
}

// SearchLocalSignalCallRecords 查询本地通话记录列表。
func (s *Signaling) SearchLocalSignalCallRecords(ctx context.Context, params *sdk_struct.SearchLocalSignalCallRecordsParams) ([]*sdk_struct.SignalCallRecordWithDialStatus, error) {
	if s.db == nil {
		return nil, sdkerrs.ErrSdkInternal.WrapMsg("db not initialized")
	}
	if params == nil {
		params = &sdk_struct.SearchLocalSignalCallRecordsParams{}
	}
	if params.Count <= 0 {
		params.Count = 20
	}
	list, err := s.db.SearchSignalCallRecords(ctx,
		params.Offset, params.Count,
		params.SessionType,
		params.DialStatus,
		params.Direction,
		params.StartTime, params.EndTime,
		params.Keyword, params.UserName)
	if err != nil {
		return nil, err
	}
	out := make([]*sdk_struct.SignalCallRecordWithDialStatus, 0, len(list))
	for _, l := range list {
		out = append(out, &sdk_struct.SignalCallRecordWithDialStatus{
			Record:              localToSignalRecord(l),
			DialStatus:          l.DialStatus,
			Direction:           l.Direction,
			ConnectTime:         l.ConnectTime,
			DialDuration:        l.DialDuration,
			CallDuration:        l.CallDuration,
			InviteeUserNickname: l.InviteeUserNickname,
		})
	}
	return out, nil
}

// GetLocalSignalCallRecordDetail 按 sID 查询单条通话详情（LRU+TTL=5min 缓存）。
func (s *Signaling) GetLocalSignalCallRecordDetail(ctx context.Context, sID string) (*sdk_struct.SignalCallRecordWithDialStatus, error) {
	if s.db == nil {
		return nil, sdkerrs.ErrSdkInternal.WrapMsg("db not initialized")
	}
	if strings.TrimSpace(sID) == "" {
		return nil, sdkerrs.ErrArgs.WrapMsg("sID is empty")
	}

	if item, ok := s.detailCache.Get(sID); ok {
		return item, nil
	}

	rec, err := s.db.GetSignalCallRecordBySID(ctx, sID)
	if err != nil {
		return nil, err
	}

	result := &sdk_struct.SignalCallRecordWithDialStatus{
		Record:              localToSignalRecord(rec),
		DialStatus:          rec.DialStatus,
		Direction:           rec.Direction,
		ConnectTime:         rec.ConnectTime,
		DialDuration:        rec.DialDuration,
		CallDuration:        rec.CallDuration,
		InviteeUserNickname: rec.InviteeUserNickname,
	}
	s.detailCache.Set(sID, result)
	return result, nil
}

// DeleteLocalSignalCallRecords 删除本地通话记录（按 sID 列表，支持批量）。
func (s *Signaling) DeleteLocalSignalCallRecords(ctx context.Context, sIDs []string) error {
	if s.db == nil {
		return sdkerrs.ErrSdkInternal.WrapMsg("db not initialized")
	}
	if len(sIDs) == 0 {
		return nil
	}
	if err := s.db.DeleteSignalCallRecords(ctx, sIDs); err != nil {
		return err
	}
	for _, sID := range sIDs {
		s.detailCache.Delete(sID)
	}
	return nil
}

// DeleteSignalRecords 删除服务端通话记录，并同步删除本地记录及缓存。
func (s *Signaling) DeleteSignalRecords(ctx context.Context, req *rtc.DeleteSignalRecordsReq) error {
	if err := api.DeleteSignalRecords.Execute(ctx, req); err != nil {
		return err
	}
	if s.db != nil && len(req.SIDs) > 0 {
		if err := s.DeleteLocalSignalCallRecords(ctx, req.SIDs); err != nil {
			log.ZWarn(ctx, "delete local call records after server delete failed", err)
		}
	}
	return nil
}

// ClearAllLocalSignalCallRecords 清空本地所有通话记录及缓存。
func (s *Signaling) ClearAllLocalSignalCallRecords(ctx context.Context) error {
	if s.db == nil {
		return sdkerrs.ErrSdkInternal.WrapMsg("db not initialized")
	}
	if err := s.db.ClearAllSignalCallRecords(ctx); err != nil {
		return err
	}
	s.detailCache.Clear()
	return nil
}

// ClearAllSignalCallRecords 清空服务端 + 本地所有记录。
func (s *Signaling) ClearAllSignalCallRecords(ctx context.Context) error {
	return s.ClearAllLocalSignalCallRecords(ctx)
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
	if invitation.Timeout == 0 {
		invitation.Timeout = defaultTimeout
	}
}

func InviteReqToJson(req *rtc.SignalInviteReq) string {
	return jsonutil.StructToJsonString(req)
}
