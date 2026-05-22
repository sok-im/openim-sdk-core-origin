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
		log.ZError(ctx, "Invite failed", err, "req", req)
		return nil, err
	}

	// 邀请发送成功后启动超时定时器（主叫侧）
	if signalInviteReq.Invitation != nil {
		s.startInviteTimer(signalInviteReq.Invitation, constant.SignalCallDirectionOutgoing)
	}

	log.ZInfo(ctx, "Invite success", "req", req, "resp", resp)

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

	log.ZInfo(ctx, "InviteInGroup success", "req", req, "resp", resp)

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

	log.ZInfo(ctx, "Accept success", "req", req, "resp", resp)

	// 接听阶段仅记录接通时间，不落库；通话结束时由 HungUp / 对端 HungUp 通知写入唯一一条记录。

	if acceptResp := resp.GetAccept(); acceptResp != nil {
		return acceptResp, nil
	}
	return &rtc.SignalAcceptResp{}, nil
}

// Reject 被叫侧拒接：取消超时定时器，写未拨通记录（missed 方向）。
func (s *Signaling) Reject(ctx context.Context, signalRejectReq *rtc.SignalRejectReq) error {
	signalRejectReq.UserID = s.loginUserID
	signalRejectReq.OpUserPlatformID = s.platformID

	req := &rtc.SignalReq{
		Payload: &rtc.SignalReq_Reject{
			Reject: signalRejectReq,
		},
	}
	resp, err := s.signalingRequest(ctx, req)
	if err != nil {
		return err
	}

	log.ZInfo(ctx, "Reject success", "req", req, "resp", resp)

	if signalRejectReq.Invitation != nil {
		s.cancelInviteTimer(signalRejectReq.Invitation.RoomID)
		inviteMs, connectMs, ok := s.popTimingForRecord(signalRejectReq.Invitation.RoomID)
		if ok {
			s.persistLocalCallRecord(ctx,
				signalRejectReq.Invitation,
				signalRejectReq.Participant,
				constant.SignalCallStatusNotConnected,
				constant.SignalCallDirectionMissed,
				constant.SignalCallActionReject,
				inviteMs, connectMs, time.Now().UnixMilli())
		}
	}
	return nil
}

// Timeout 主叫侧超时未接通：通知服务端并写未接通记录（outgoing 方向）。
func (s *Signaling) Timeout(ctx context.Context, signalTimeoutReq *rtc.SignalTimeoutReq) error {
	signalTimeoutReq.UserID = s.loginUserID

	if signalTimeoutReq.Invitation != nil {
		s.cancelInviteTimer(signalTimeoutReq.Invitation.RoomID)
	}

	/*
		req := &rtc.SignalReq{
			Payload: &rtc.SignalReq_Timeout{
				Timeout: signalTimeoutReq,
			},
		}
		_, err := s.signalingRequest(ctx, req)
		if err != nil {
			return err
		}
	*/

	if signalTimeoutReq.Invitation != nil && signalTimeoutReq.Invitation.InviterUserID == s.loginUserID {
		if _, connectMs := s.peekTiming(signalTimeoutReq.Invitation.RoomID); !callWasConnected(connectMs) {
			inviteMs, connectMs, ok := s.popTimingForRecord(signalTimeoutReq.Invitation.RoomID)
			if ok {
				s.persistLocalCallRecord(ctx,
					signalTimeoutReq.Invitation,
					nil,
					constant.SignalCallStatusNotConnected,
					constant.SignalCallDirectionOutgoing,
					constant.SignalCallActionTimeout,
					inviteMs, connectMs, time.Now().UnixMilli())
			}
		}

		if listener := s.listener(); listener != nil {
			listener.OnInvitationTimeout(jsonutil.StructToJsonString(signalTimeoutReq.Invitation))
		}
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
	resp, err := s.signalingRequest(ctx, req)
	if err != nil {
		return err
	}

	log.ZInfo(ctx, "Cancel success", "req", req, "resp", resp)

	if signalCancelReq.Invitation != nil {
		s.cancelInviteTimer(signalCancelReq.Invitation.RoomID)
		if _, connectMs := s.peekTiming(signalCancelReq.Invitation.RoomID); !callWasConnected(connectMs) {
			inviteMs, connectMs, ok := s.popTimingForRecord(signalCancelReq.Invitation.RoomID)
			if ok {
				s.persistLocalCallRecord(ctx,
					signalCancelReq.Invitation,
					signalCancelReq.Participant,
					constant.SignalCallStatusNotConnected,
					constant.SignalCallDirectionOutgoing,
					constant.SignalCallActionCancel,
					inviteMs, connectMs, time.Now().UnixMilli())
			}
		}
	}
	return nil
}

// HungUp 任意一方挂断：取消超时定时器，写通话记录。
// 若 connectMs=0（被叫未接通即挂断），记录为未接通；否则记录为已接听。
func (s *Signaling) HungUp(ctx context.Context, signalHungUpReq *rtc.SignalHungUpReq) error {
	signalHungUpReq.UserID = s.loginUserID

	req := &rtc.SignalReq{
		Payload: &rtc.SignalReq_HungUp{
			HungUp: signalHungUpReq,
		},
	}
	resp, err := s.signalingRequest(ctx, req)
	if err != nil {
		return err
	}

	log.ZInfo(ctx, "HungUp success", "req", req, "resp", resp)

	if signalHungUpReq.Invitation != nil {
		s.cancelInviteTimer(signalHungUpReq.Invitation.RoomID)
		direction := constant.SignalCallDirectionOutgoing
		if signalHungUpReq.Invitation.InviterUserID != s.loginUserID {
			direction = constant.SignalCallDirectionIncoming
		}
		inviteMs, connectMs, ok := s.popTimingForRecord(signalHungUpReq.Invitation.RoomID)
		if ok {
			status := constant.SignalCallStatusAnswered
			if !callWasConnected(connectMs) {
				status = constant.SignalCallStatusNotConnected
				if direction == constant.SignalCallDirectionIncoming {
					direction = constant.SignalCallDirectionMissed
				}
			}
			s.persistLocalCallRecord(ctx,
				signalHungUpReq.Invitation,
				nil,
				status,
				direction,
				constant.SignalCallActionHungUp,
				inviteMs, connectMs, time.Now().UnixMilli())
		}
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

func (s *Signaling) GetRoomByGroupID(ctx context.Context, groupID string) (*rtc.SignalGetRoomByGroupIDResp, error) {
	return api.SignalGetRoomByGroupID.Invoke(ctx, &rtc.SignalGetRoomByGroupIDReq{GroupID: groupID})
}

func (s *Signaling) GetSignalInvitationInfoStartApp(ctx context.Context, req *rtc.GetSignalInvitationInfoStartAppReq) (*rtc.GetSignalInvitationInfoStartAppResp, error) {
	req.UserID = s.loginUserID
	return api.GetSignalInvitationInfoStartApp.Invoke(ctx, req)
}

// GetSignalInvitationRecords 调用服务端历史接口（与本地通话记录独立）。
func (s *Signaling) GetSignalInvitationRecords(ctx context.Context, req *rtc.GetSignalInvitationRecordsReq) (*rtc.GetSignalInvitationRecordsResp, error) {
	return api.GetSignalInvitationRecords.Invoke(ctx, req)
}

// GetLocalCallRecords 按用户查询本地通话记录；status：0=全部 1=已接听 2=未接通。
func (s *Signaling) GetLocalCallRecords(ctx context.Context, params *sdk_struct.GetLocalCallRecordsParams) (*sdk_struct.GetLocalCallRecordsResp, error) {
	if s.db == nil {
		return nil, sdkerrs.ErrSdkInternal.WrapMsg("db not initialized")
	}
	if params == nil {
		params = &sdk_struct.GetLocalCallRecordsParams{}
	}
	if params.Count <= 0 {
		params.Count = 20
	}
	total, err := s.db.CountSignalCallRecordsByUser(ctx, params.UserID, params.Status, params.StartTime, params.EndTime)
	if err != nil {
		return nil, err
	}
	list, err := s.db.SearchSignalCallRecordsByUser(ctx, params.UserID, params.Status, params.Offset, params.Count, params.StartTime, params.EndTime)
	if err != nil {
		return nil, err
	}
	records := make([]*sdk_struct.SignalCallRecordWithDialStatus, 0, len(list))
	for _, l := range list {
		records = append(records, localRecordToSDK(l))
	}
	return &sdk_struct.GetLocalCallRecordsResp{Total: total, Records: records}, nil
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
		params.Status,
		params.Direction,
		params.StartTime, params.EndTime,
		params.Keyword, params.UserName, params.InviteeNickname, "", "")
	if err != nil {
		return nil, err
	}
	out := make([]*sdk_struct.SignalCallRecordWithDialStatus, 0, len(list))
	for _, l := range list {
		out = append(out, localRecordToSDK(l))
	}
	return out, nil
}

// GetLocalMissedCallRecords 查询本地未接来电（被叫未接：status=未接通，direction=被叫-未接）。
func (s *Signaling) GetLocalMissedCallRecords(ctx context.Context, params *sdk_struct.GetLocalMissedCallRecordsParams) (*sdk_struct.GetLocalCallRecordsResp, error) {
	if s.db == nil {
		return nil, sdkerrs.ErrSdkInternal.WrapMsg("db not initialized")
	}
	if params == nil {
		params = &sdk_struct.GetLocalMissedCallRecordsParams{}
	}
	if params.Count <= 0 {
		params.Count = 20
	}
	total, err := s.db.CountSignalCallRecords(ctx,
		params.SessionType,
		constant.SignalCallStatusNotConnected,
		constant.SignalCallDirectionMissed,
		params.StartTime, params.EndTime,
		params.Keyword, "", "", params.UserID, "")
	if err != nil {
		return nil, err
	}
	list, err := s.db.SearchSignalCallRecords(ctx,
		params.Offset, params.Count,
		params.SessionType,
		constant.SignalCallStatusNotConnected,
		constant.SignalCallDirectionMissed,
		params.StartTime, params.EndTime,
		params.Keyword, "", "", params.UserID, "")
	if err != nil {
		return nil, err
	}
	records := make([]*sdk_struct.SignalCallRecordWithDialStatus, 0, len(list))
	for _, l := range list {
		records = append(records, localRecordToSDK(l))
	}
	return &sdk_struct.GetLocalCallRecordsResp{Total: total, Records: records}, nil
}

// GetLocalAnsweredCallRecords 查询本地已接通通话（status=已接听，含主叫/被叫方向）。
func (s *Signaling) GetLocalAnsweredCallRecords(ctx context.Context, params *sdk_struct.GetLocalAnsweredCallRecordsParams) (*sdk_struct.GetLocalCallRecordsResp, error) {
	if s.db == nil {
		return nil, sdkerrs.ErrSdkInternal.WrapMsg("db not initialized")
	}
	if params == nil {
		params = &sdk_struct.GetLocalAnsweredCallRecordsParams{}
	}
	if params.Count <= 0 {
		params.Count = 20
	}
	total, err := s.db.CountSignalCallRecords(ctx,
		params.SessionType,
		constant.SignalCallStatusAnswered,
		0,
		params.StartTime, params.EndTime,
		params.Keyword, "", "", "", params.UserID)
	if err != nil {
		return nil, err
	}
	list, err := s.db.SearchSignalCallRecords(ctx,
		params.Offset, params.Count,
		params.SessionType,
		constant.SignalCallStatusAnswered,
		0,
		params.StartTime, params.EndTime,
		params.Keyword, "", "", "", params.UserID)
	if err != nil {
		return nil, err
	}
	records := make([]*sdk_struct.SignalCallRecordWithDialStatus, 0, len(list))
	for _, l := range list {
		records = append(records, localRecordToSDK(l))
	}
	return &sdk_struct.GetLocalCallRecordsResp{Total: total, Records: records}, nil
}

// GetLocalAllCallRecords 查询本地全部通话记录（含已接听、未接通等，不筛选 status/direction）。
func (s *Signaling) GetLocalAllCallRecords(ctx context.Context, params *sdk_struct.GetLocalAllCallRecordsParams) (*sdk_struct.GetLocalCallRecordsResp, error) {
	if s.db == nil {
		return nil, sdkerrs.ErrSdkInternal.WrapMsg("db not initialized")
	}
	if params == nil {
		params = &sdk_struct.GetLocalAllCallRecordsParams{}
	}
	if params.Count <= 0 {
		params.Count = 20
	}
	total, err := s.db.CountSignalCallRecords(ctx,
		params.SessionType,
		0, 0,
		params.StartTime, params.EndTime,
		params.Keyword, "", "", "", params.UserID)
	if err != nil {
		return nil, err
	}
	list, err := s.db.SearchSignalCallRecords(ctx,
		params.Offset, params.Count,
		params.SessionType,
		0, 0,
		params.StartTime, params.EndTime,
		params.Keyword, "", "", "", params.UserID)
	if err != nil {
		return nil, err
	}
	records := make([]*sdk_struct.SignalCallRecordWithDialStatus, 0, len(list))
	for _, l := range list {
		records = append(records, localRecordToSDK(l))
	}
	return &sdk_struct.GetLocalCallRecordsResp{Total: total, Records: records}, nil
}

// GetLocalCallRecordsByUserName 按用户名模糊查询本地通话记录。
func (s *Signaling) GetLocalCallRecordsByUserName(ctx context.Context, params *sdk_struct.GetLocalCallRecordsByUserNameParams) (*sdk_struct.GetLocalCallRecordsResp, error) {
	if s.db == nil {
		return nil, sdkerrs.ErrSdkInternal.WrapMsg("db not initialized")
	}
	if params == nil {
		params = &sdk_struct.GetLocalCallRecordsByUserNameParams{}
	}
	if strings.TrimSpace(params.UserName) == "" {
		return nil, sdkerrs.ErrArgs.WrapMsg("userName is empty")
	}
	if params.Count <= 0 {
		params.Count = 20
	}
	total, err := s.db.CountSignalCallRecords(ctx,
		params.SessionType,
		params.Status,
		params.Direction,
		params.StartTime, params.EndTime,
		"", params.UserName, "", "", "")
	if err != nil {
		return nil, err
	}
	list, err := s.db.SearchSignalCallRecords(ctx,
		params.Offset, params.Count,
		params.SessionType,
		params.Status,
		params.Direction,
		params.StartTime, params.EndTime,
		"", params.UserName, "", "", "")
	if err != nil {
		return nil, err
	}
	records := make([]*sdk_struct.SignalCallRecordWithDialStatus, 0, len(list))
	for _, l := range list {
		records = append(records, localRecordToSDK(l))
	}
	return &sdk_struct.GetLocalCallRecordsResp{Total: total, Records: records}, nil
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

	result := localRecordToSDK(rec)
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

// ClearAllSignalCallRecords 清空本地所有通话记录（仅本地，不涉及服务端）。
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
