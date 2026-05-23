package group

import (
	"context"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/api"
	"github.com/openimsdk/protocol/group"
	"github.com/openimsdk/protocol/sdkws"
)

func (g *Group) getFullGroupMemberUserIDs(ctx context.Context, req *group.GetFullGroupMemberUserIDsReq) (*group.GetFullGroupMemberUserIDsResp, error) {
	return api.GetFullGroupMemberUserIDs.Invoke(ctx, req)
}

func (g *Group) getIncrementalJoinGroup(ctx context.Context, req *group.GetIncrementalJoinGroupReq) (*group.GetIncrementalJoinGroupResp, error) {
	return api.GetIncrementalJoinGroup.Invoke(ctx, req)
}

func (g *Group) getFullJoinGroupIDs(ctx context.Context, req *group.GetFullJoinGroupIDsReq) (*group.GetFullJoinGroupIDsResp, error) {
	return api.GetFullJoinedGroupIDs.Invoke(ctx, req)
}

func (g *Group) getIncrementalGroupMemberBatch(ctx context.Context, reqs []*group.GetIncrementalGroupMemberReq) (map[string]*group.GetIncrementalGroupMemberResp, error) {
	req := &group.BatchGetIncrementalGroupMemberReq{UserID: g.loginUserID, ReqList: reqs}
	return api.ExtractField(ctx, api.GetIncrementalGroupMemberBatch.Invoke, req, (*group.BatchGetIncrementalGroupMemberResp).GetRespList)
}

func (g *Group) createGroup(ctx context.Context, req *group.CreateGroupReq) (*group.CreateGroupResp, error) {
	return api.CreateGroup.Invoke(ctx, req)
}

func (g *Group) joinGroup(ctx context.Context, req *group.JoinGroupReq) error {
	return api.JoinGroup.Execute(ctx, req)
}

func (g *Group) quitGroup(ctx context.Context, groupID string) error {
	return api.QuitGroup.Execute(ctx, &group.QuitGroupReq{GroupID: groupID, UserID: g.loginUserID})
}

func (g *Group) dismissGroup(ctx context.Context, groupID string) error {
	return api.DismissGroup.Execute(ctx, &group.DismissGroupReq{GroupID: groupID})
}

func (g *Group) setGroupInfo(ctx context.Context, req *group.SetGroupInfoExReq) error {
	return api.SetGroupInfoEx.Execute(ctx, req)
}

func (g *Group) setSendMessageSetting(ctx context.Context, req *api.SetSendMessageSettingReq) error {
	return api.SetSendMessageSetting.Execute(ctx, req)
}

func (g *Group) getGroupSetting(ctx context.Context, groupID string) (*api.GetGroupSettingResp, error) {
	return api.GetGroupSetting.Invoke(ctx, &api.GetGroupSettingReq{GroupID: groupID})
}

func (g *Group) getSendMessageSetting(ctx context.Context, groupID string) (*api.GetSendMessageSettingResp, error) {
	resp, err := g.getGroupSetting(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return &api.GetSendMessageSettingResp{
		GroupID:      resp.GroupID,
		AllowSendMsg: resp.AllowSendMsg,
	}, nil
}

func (g *Group) setInviteSetting(ctx context.Context, req *api.SetInviteSettingReq) error {
	return api.SetInviteSetting.Execute(ctx, req)
}

func (g *Group) getInviteSetting(ctx context.Context, groupID string) (*api.GetInviteSettingResp, error) {
	resp, err := g.getGroupSetting(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return &api.GetInviteSettingResp{
		GroupID:        resp.GroupID,
		AllowAddMember: resp.AllowAddMember,
	}, nil
}

func (g *Group) setPinSetting(ctx context.Context, req *api.SetPinSettingReq) error {
	return api.SetPinSetting.Execute(ctx, req)
}

func (g *Group) getPinSetting(ctx context.Context, groupID string) (*api.GetPinSettingResp, error) {
	resp, err := g.getGroupSetting(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return &api.GetPinSettingResp{
		GroupID:     resp.GroupID,
		AllowPinMsg: resp.AllowPinMsg,
	}, nil
}

func (g *Group) setEditSetting(ctx context.Context, req *api.SetEditSettingReq) error {
	return api.SetEditSetting.Execute(ctx, req)
}

func (g *Group) setBurnSetting(ctx context.Context, req *api.SetBurnSettingReq) error {
	return api.SetBurnSetting.Execute(ctx, req)
}

func (g *Group) getBurnSetting(ctx context.Context, groupID string) (*api.GetBurnSettingResp, error) {
	resp, err := g.getGroupSetting(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return &api.GetBurnSettingResp{
		GroupID:         resp.GroupID,
		AllowMemberBurn: resp.AllowMemberBurn,
	}, nil
}

func (g *Group) getEditSetting(ctx context.Context, groupID string) (*api.GetEditSettingResp, error) {
	resp, err := g.getGroupSetting(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return &api.GetEditSettingResp{
		GroupID:            resp.GroupID,
		AllowEditGroupInfo: resp.AllowEditGroupInfo,
	}, nil
}

func (g *Group) setMsgBurnDuration(ctx context.Context, req *api.SetMsgBurnDurationReq) error {
	return api.SetMsgBurnDuration.Execute(ctx, req)
}

func (g *Group) getMsgBurnDuration(ctx context.Context, groupID string) (*api.GetMsgBurnDurationResp, error) {
	return api.GetMsgBurnDuration.Invoke(ctx, &api.GetMsgBurnDurationReq{GroupID: groupID})
}

func (g *Group) setGroupAnnouncement(ctx context.Context, req *api.SetGroupAnnouncementReq) error {
	return api.SetGroupAnnouncement.Execute(ctx, req)
}

func (g *Group) getGroupAnnouncement(ctx context.Context, groupID string) (*api.GetGroupAnnouncementResp, error) {
	return api.GetGroupAnnouncement.Invoke(ctx, &api.GetGroupAnnouncementReq{GroupID: groupID})
}

func (g *Group) setGroupMemberInfo(ctx context.Context, req *group.SetGroupMemberInfoReq) error {
	return api.SetGroupMemberInfo.Execute(ctx, req)
}

func (g *Group) kickGroupMember(ctx context.Context, req *group.KickGroupMemberReq) error {
	return api.KickGroupMember.Execute(ctx, req)
}

func (g *Group) transferGroup(ctx context.Context, req *group.TransferGroupOwnerReq) error {
	return api.TransferGroup.Execute(ctx, req)
}

func (g *Group) cancelMuteGroupMember(ctx context.Context, req *group.CancelMuteGroupMemberReq) error {
	return api.CancelMuteGroupMember.Execute(ctx, req)
}

func (g *Group) muteGroupMember(ctx context.Context, req *group.MuteGroupMemberReq) error {
	return api.MuteGroupMember.Execute(ctx, req)
}

func (g *Group) muteGroup(ctx context.Context, groupID string) error {
	return api.MuteGroup.Execute(ctx, &group.MuteGroupReq{GroupID: groupID})
}

func (g *Group) cancelMuteGroup(ctx context.Context, groupID string) error {
	return api.CancelMuteGroup.Execute(ctx, &group.CancelMuteGroupReq{GroupID: groupID})
}

func (g *Group) getDesignatedGroupMembers(ctx context.Context, groupID string, userIDs []string) ([]*sdkws.GroupMemberFullInfo, error) {
	req := &group.GetGroupMembersInfoReq{GroupID: groupID, UserIDs: userIDs}
	return api.ExtractField(ctx, api.GetGroupMembersInfo.Invoke, req, (*group.GetGroupMembersInfoResp).GetMembers)
}

func (g *Group) getServerSelfGroupApplication(ctx context.Context, groupIDs []string,
	handleResults []int32, pageNumber, showNumber int32) ([]*sdkws.GroupRequest, error) {
	req := &group.GetUserReqApplicationListReq{UserID: g.loginUserID, Pagination: &sdkws.RequestPagination{PageNumber: pageNumber, ShowNumber: showNumber},
		GroupIDs: groupIDs, HandleResults: handleResults}
	if showNumber <= 0 {
		return api.Page(ctx, req, api.GetSendGroupApplicationList.Invoke, (*group.GetUserReqApplicationListResp).GetGroupRequests)
	}
	resp, err := api.GetSendGroupApplicationList.Invoke(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetGroupRequests(), nil
}

func (g *Group) getServerJoinGroup(ctx context.Context) ([]*sdkws.GroupInfo, error) {
	req := &group.GetJoinedGroupListReq{FromUserID: g.loginUserID, Pagination: &sdkws.RequestPagination{}}
	return api.Page(ctx, req, api.GetJoinedGroupList.Invoke, (*group.GetJoinedGroupListResp).GetGroups)
}

func (g *Group) getServerAdminGroupApplicationList(ctx context.Context, groupIDs []string,
	handleResults []int32, pageNumber, showNumber int32) ([]*sdkws.GroupRequest, error) {
	req := &group.GetGroupApplicationListReq{FromUserID: g.loginUserID, Pagination: &sdkws.RequestPagination{PageNumber: pageNumber, ShowNumber: showNumber},
		GroupIDs: groupIDs, HandleResults: handleResults}
	if showNumber <= 0 {
		return api.Page(ctx, req, api.GetRecvGroupApplicationList.Invoke, (*group.GetGroupApplicationListResp).GetGroupRequests)
	}
	resp, err := api.GetRecvGroupApplicationList.Invoke(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.GetGroupRequests(), nil
}

func (g *Group) getGroupsInfoFromServer(ctx context.Context, groupIDs []string) ([]*sdkws.GroupInfo, error) {
	req := &group.GetGroupsInfoReq{GroupIDs: groupIDs}
	return api.ExtractField(ctx, api.GetGroupsInfo.Invoke, req, (*group.GetGroupsInfoResp).GetGroupInfos)
}

func (g *Group) inviteUserToGroup(ctx context.Context, req *group.InviteUserToGroupReq) error {
	return api.InviteUserToGroup.Execute(ctx, req)
}

func (g *Group) handlerGroupApplication(ctx context.Context, req *group.GroupApplicationResponseReq) error {
	return api.AcceptGroupApplication.Execute(ctx, req)
}

func (g *Group) getGroupApplicationUnhandledCount(ctx context.Context, time int64) (int32, error) {
	req := &group.GetGroupApplicationUnhandledCountReq{UserID: g.loginUserID, Time: time}
	resp, err := api.GetGroupApplicationUnhandledCount.Invoke(ctx, req)
	if err != nil {
		return 0, err
	}
	return int32(resp.GetCount()), nil
}

func (g *Group) getCommonGroupsWithFriend(ctx context.Context, friendUserID string) (*group.GetCommonGroupsWithFriendResp, error) {
	return api.GetCommonGroupsWithFriend.Invoke(ctx, &group.GetCommonGroupsWithFriendReq{FriendUserID: friendUserID})
}

func (g *Group) pinGroupMessage(ctx context.Context, req *group.PinGroupMessageReq) (*group.PinGroupMessageResp, error) {
	return api.PinGroupMessage.Invoke(ctx, req)
}

func (g *Group) unpinGroupMessage(ctx context.Context, req *group.UnpinGroupMessageReq) (*group.UnpinGroupMessageResp, error) {
	return api.UnpinGroupMessage.Invoke(ctx, req)
}

func (g *Group) getGroupPinnedMessages(ctx context.Context, groupID string) (*group.GetGroupPinnedMessagesResp, error) {
	return api.GetGroupPinnedMessages.Invoke(ctx, &group.GetGroupPinnedMessagesReq{GroupID: groupID})
}

func (g *Group) setGroupMute(ctx context.Context, req *group.SetGroupMuteReq) error {
	return api.SetGroupMute.Execute(ctx, req)
}

func (g *Group) getGroupMute(ctx context.Context, req *group.GetGroupMuteReq) (*group.GetGroupMuteResp, error) {
	return api.GetGroupMute.Invoke(ctx, req)
}

func (g *Group) pinGroup(ctx context.Context, req *group.PinGroupReq) error {
	return api.PinGroup.Execute(ctx, req)
}

func (g *Group) unpinGroup(ctx context.Context, req *group.UnpinGroupReq) error {
	return api.UnpinGroup.Execute(ctx, req)
}

func (g *Group) setInviteLinkSetting(ctx context.Context, req *api.SetInviteLinkSettingReq) error {
	return api.SetInviteLinkSetting.Execute(ctx, req)
}

func (g *Group) createGroupInviteLink(ctx context.Context, req *group.CreateGroupInviteLinkReq) (*group.CreateGroupInviteLinkResp, error) {
	return api.CreateGroupInviteLink.Invoke(ctx, req)
}

func (g *Group) getGroupInviteLink(ctx context.Context, req *group.GetGroupInviteLinkReq) (*group.GetGroupInviteLinkResp, error) {
	return api.GetGroupInviteLink.Invoke(ctx, req)
}

func (g *Group) joinGroupByInviteLink(ctx context.Context, req *group.JoinGroupByInviteLinkReq) error {
	return api.JoinGroupByInviteLink.Execute(ctx, req)
}

func (g *Group) revokeGroupInviteLink(ctx context.Context, req *group.RevokeGroupInviteLinkReq) error {
	return api.RevokeGroupInviteLink.Execute(ctx, req)
}

func (g *Group) listGroupInviteLinks(ctx context.Context, req *group.ListGroupInviteLinksReq) (*group.ListGroupInviteLinksResp, error) {
	return api.ListGroupInviteLinks.Invoke(ctx, req)
}
