package signaling

import (
	"context"
	"strings"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/api"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/sdkerrs"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/utils"
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
	if err != nil {
		return err
	}
	// 异步写本地记录（被叫拒接 → 未拨通）
	if signalRejectReq.Invitation != nil {
		s.persistLocalCallRecord(ctx, signalRejectReq.Invitation, signalRejectReq.Participant, constant.SignalCallDialStatusNotConnected)
	}
	return nil
}

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
	// 异步写本地记录（取消 → 未拨通）
	if signalCancelReq.Invitation != nil {
		s.persistLocalCallRecord(ctx, signalCancelReq.Invitation, signalCancelReq.Participant, constant.SignalCallDialStatusNotConnected)
	}
	return nil
}

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
	// 异步写本地记录（挂断 → 已拨通）
	if signalHungUpReq != nil && signalHungUpReq.Invitation != nil {
		s.persistLocalCallRecord(ctx, signalHungUpReq.Invitation, nil, constant.SignalCallDialStatusConnected)
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

// GetSignalInvitationRecords 仍可调 RTC 服务端接口（与本地通话记录无关）；本地列表仅来自客户端写入，请用 SearchLocalSignalCallRecords。
func (s *Signaling) GetSignalInvitationRecords(ctx context.Context, req *rtc.GetSignalInvitationRecordsReq) (*rtc.GetSignalInvitationRecordsResp, error) {
	return api.GetSignalInvitationRecords.Invoke(ctx, req)
}

// SearchLocalSignalCallRecords 查询仅由本机产生的本地通话记录；params.dialStatus：0=全部，1=未拨通，2=已拨通；params.userName 仅按被叫 userID 模糊查（callee_match_text）。
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
	list, err := s.db.SearchSignalCallRecords(ctx, params.Offset, params.Count, params.SessionType, params.DialStatus, params.StartTime, params.EndTime, params.Keyword, params.UserName)
	if err != nil {
		return nil, err
	}
	out := make([]*sdk_struct.SignalCallRecordWithDialStatus, 0, len(list))
	for _, l := range list {
		out = append(out, &sdk_struct.SignalCallRecordWithDialStatus{
			Record:     localToSignalRecord(l),
			DialStatus: l.DialStatus,
		})
	}
	return out, nil
}

// GetLocalSignalCallRecordDetail 按本地记录主键 sID 查询单条通话详情。
// 使用 LRU + TTL=5min 缓存加速重复查询。
func (s *Signaling) GetLocalSignalCallRecordDetail(ctx context.Context, sID string) (*sdk_struct.SignalCallRecordWithDialStatus, error) {
	if s.db == nil {
		return nil, sdkerrs.ErrSdkInternal.WrapMsg("db not initialized")
	}
	if strings.TrimSpace(sID) == "" {
		return nil, sdkerrs.ErrArgs.WrapMsg("sID is empty")
	}

	// 1. 查缓存
	if item, ok := s.detailCache.Get(sID); ok {
		return item, nil
	}

	// 2. 查 DB
	rec, err := s.db.GetSignalCallRecordBySID(ctx, sID)
	if err != nil {
		return nil, err
	}

	result := &sdk_struct.SignalCallRecordWithDialStatus{
		Record:     localToSignalRecord(rec),
		DialStatus: rec.DialStatus,
	}

	// 3. 写入缓存
	s.detailCache.Set(sID, result)

	return result, nil
}

// DeleteLocalSignalCallRecords 删除本地通话记录（按 sID 列表，支持批量删除）。
func (s *Signaling) DeleteLocalSignalCallRecords(ctx context.Context, sIDs []string) error {
	if s.db == nil {
		return sdkerrs.ErrSdkInternal.WrapMsg("db not initialized")
	}
	if len(sIDs) == 0 {
		return nil
	}
	return s.db.DeleteSignalCallRecords(ctx, sIDs)
}

// DeleteSignalRecords 删除服务端通话记录（同时可选择删除本地记录）。
func (s *Signaling) DeleteSignalRecords(ctx context.Context, req *rtc.DeleteSignalRecordsReq) error {
	if err := api.DeleteSignalRecords.Execute(ctx, req); err != nil {
		return err
	}
	// 可选：同时删除本地记录
	if s.db != nil && len(req.SIDs) > 0 {
		if err := s.DeleteLocalSignalCallRecords(ctx, req.SIDs); err != nil {
			log.ZWarn(ctx, "delete local call records after server delete failed", err)
			// 不返回错误，服务端已删除成功
		}
	}
	return nil
}

// ClearAllLocalSignalCallRecords 清空本地所有音视频通话记录（推荐用于 UI「清空列表」按钮）。
func (s *Signaling) ClearAllLocalSignalCallRecords(ctx context.Context) error {
	if s.db == nil {
		return sdkerrs.ErrSdkInternal.WrapMsg("db not initialized")
	}
	return s.db.ClearAllSignalCallRecords(ctx)
}

// ClearAllSignalCallRecords 清空服务端 + 本地所有记录（当前服务端暂无 ClearAll 接口，未来可扩展）。
func (s *Signaling) ClearAllSignalCallRecords(ctx context.Context) error {
	// TODO: 如果服务端提供 ClearAll 接口，可在此先调用
	// 目前仅清本地
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
	if invitation.RoomID == "" {
		invitation.RoomID = "room-" + utils.OperationIDGenerator()
	}
	if invitation.Timeout == 0 {
		invitation.Timeout = defaultTimeout
	}
}

func InviteReqToJson(req *rtc.SignalInviteReq) string {
	return jsonutil.StructToJsonString(req)
}
