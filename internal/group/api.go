// Copyright © 2023 OpenIM SDK. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package group

import (
	"context"
	"time"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/utils"
	"github.com/openimsdk/tools/errs"

	"github.com/openimsdk/tools/utils/datautil"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/datafetcher"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/api"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/sdk_params_callback"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/sdkerrs"

	"github.com/openimsdk/protocol/group"
	"github.com/openimsdk/protocol/sdkws"
)

func (g *Group) CreateGroup(ctx context.Context, req *group.CreateGroupReq) (*sdkws.GroupInfo, error) {
	if req.OwnerUserID == "" {
		req.OwnerUserID = g.loginUserID
	}
	if req.GroupInfo.GroupType != constant.WorkingGroup {
		return nil, sdkerrs.ErrGroupType
	}
	req.GroupInfo.CreatorUserID = g.loginUserID
	resp, err := g.createGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()

	if err := g.IncrSyncJoinGroup(ctx); err != nil {
		return nil, err
	}
	if err := g.IncrSyncGroupAndMember(ctx, resp.GroupInfo.GroupID); err != nil {
		return nil, err
	}
	return resp.GroupInfo, nil
}

func (g *Group) JoinGroup(ctx context.Context, groupID, reqMsg string, joinSource int32, ex string) error {
	req := &group.JoinGroupReq{GroupID: groupID, ReqMessage: reqMsg, JoinSource: joinSource, InviterUserID: g.loginUserID, Ex: ex}
	if err := g.joinGroup(ctx, req); err != nil {
		return err
	}
	return nil
}

func (g *Group) QuitGroup(ctx context.Context, groupID string) error {
	if err := g.quitGroup(ctx, groupID); err != nil {
		return err
	}
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()

	if err := g.IncrSyncJoinGroup(ctx); err != nil {
		return err
	}
	return nil
}

func (g *Group) DismissGroup(ctx context.Context, groupID string) error {
	if err := g.dismissGroup(ctx, groupID); err != nil {
		return err
	}
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()

	if err := g.IncrSyncJoinGroup(ctx); err != nil {
		return err
	}
	return nil
}

func (g *Group) ChangeGroupMute(ctx context.Context, groupID string, isMute bool) (err error) {
	if isMute {
		err = g.muteGroup(ctx, groupID)
	} else {
		err = g.cancelMuteGroup(ctx, groupID)
	}
	if err != nil {
		return err
	}

	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()

	if err := g.IncrSyncGroupAndMember(ctx, groupID); err != nil {
		return err
	}
	return g.refreshGroupInfoFromServer(ctx, groupID)
}

func (g *Group) ChangeGroupMemberMute(ctx context.Context, groupID, userID string, mutedSeconds int) error {
	if mutedSeconds == 0 {
		return g.cancelMuteGroupMember(ctx, &group.CancelMuteGroupMemberReq{GroupID: groupID, UserID: userID})
	} else {
		return g.muteGroupMember(ctx, &group.MuteGroupMemberReq{GroupID: groupID, UserID: userID, MutedSeconds: uint32(mutedSeconds)})
	}
}

func (g *Group) TransferGroupOwner(ctx context.Context, groupID, newOwnerUserID string) error {
	req := &group.TransferGroupOwnerReq{GroupID: groupID, OldOwnerUserID: g.loginUserID, NewOwnerUserID: newOwnerUserID}
	if err := g.transferGroup(ctx, req); err != nil {
		return err
	}
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()

	if err := g.IncrSyncGroupAndMember(ctx, groupID); err != nil {
		return err
	}
	return nil
}

func (g *Group) KickGroupMember(ctx context.Context, groupID string, reason string, userIDList []string) error {
	req := &group.KickGroupMemberReq{GroupID: groupID, KickedUserIDs: userIDList, Reason: reason}
	if err := g.kickGroupMember(ctx, req); err != nil {
		return err
	}

	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()

	return g.IncrSyncGroupAndMember(ctx, groupID)
}

func (g *Group) SetGroupInfo(ctx context.Context, groupInfo *group.SetGroupInfoExReq) error {
	if err := g.setGroupInfo(ctx, groupInfo); err != nil {
		return err
	}

	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()

	return g.IncrSyncJoinGroup(ctx)
}

// SetSendMessageSetting 设置群成员发消息权限（HTTP POST /group/set_send_message_setting）：
//
//	allowSendMsg 0 = 全员可发，1 = 仅群主/管理员可发
func (g *Group) SetSendMessageSetting(ctx context.Context, groupID string, allowSendMsg int32) error {
	if err := g.setSendMessageSetting(ctx, &api.SetSendMessageSettingReq{
		GroupID:      groupID,
		AllowSendMsg: allowSendMsg,
	}); err != nil {
		return err
	}
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()
	return g.IncrSyncJoinGroup(ctx)
}

// GetGroupSetting 查询群权限相关设置（HTTP POST /group/get_group_setting）。
func (g *Group) GetGroupSetting(ctx context.Context, groupID string) (*api.GetGroupSettingResp, error) {
	return g.getGroupSetting(ctx, groupID)
}

// GetSendMessageSetting 查询群发消息权限（委托 GetGroupSetting，成功回调仅含 groupID、allowSendMsg）。
func (g *Group) GetSendMessageSetting(ctx context.Context, groupID string) (*api.GetSendMessageSettingResp, error) {
	return g.getSendMessageSetting(ctx, groupID)
}

// SetInviteLinkSetting 开启/关闭群邀请链接（HTTP POST /group/set_invite_link_setting）：
//
//	enableInviteLink 0 = 关闭，1 = 开启；首次开启时服务端默认 needVerification=2（分享链接免审）
func (g *Group) SetInviteLinkSetting(ctx context.Context, groupID string, enableInviteLink int32) error {
	if err := g.setInviteLinkSetting(ctx, &api.SetInviteLinkSettingReq{
		GroupID:          groupID,
		EnableInviteLink: enableInviteLink,
	}); err != nil {
		return err
	}
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()
	return g.IncrSyncJoinGroup(ctx)
}

// CreateGroupInviteLink 生成群邀请链接（HTTP POST /group/create_invite_link）。
func (g *Group) CreateGroupInviteLink(ctx context.Context, groupID string, expireSeconds int64, maxUseCount int32) (*group.CreateGroupInviteLinkResp, error) {
	resp, err := g.createGroupInviteLink(ctx, &group.CreateGroupInviteLinkReq{
		GroupID:       groupID,
		ExpireSeconds: expireSeconds,
		MaxUseCount:   maxUseCount,
	})
	if err != nil {
		return nil, err
	}
	if err := g.updateLocalGroupInviteLink(ctx, groupID, groupInviteLinkInfoToSDKWS(resp.GetLink())); err != nil {
		_ = g.refreshLocalGroupInviteLink(ctx, groupID)
	}
	return resp, nil
}

// GetGroupInviteLink 查询邀请链接详情及群预览（HTTP POST /group/get_invite_link，无需登录）。
func (g *Group) GetGroupInviteLink(ctx context.Context, linkID string) (*group.GetGroupInviteLinkResp, error) {
	return g.getGroupInviteLink(ctx, &group.GetGroupInviteLinkReq{LinkID: linkID})
}

// JoinGroupByInviteLink 通过邀请链接申请入群（HTTP POST /group/join_by_invite_link）。
func (g *Group) JoinGroupByInviteLink(ctx context.Context, linkID, reqMessage string) error {
	if err := g.joinGroupByInviteLink(ctx, &group.JoinGroupByInviteLinkReq{
		LinkID:     linkID,
		ReqMessage: reqMessage,
	}); err != nil {
		return err
	}
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()
	return g.IncrSyncJoinGroup(ctx)
}

// RevokeGroupInviteLink 吊销群邀请链接（HTTP POST /group/revoke_invite_link）。
func (g *Group) RevokeGroupInviteLink(ctx context.Context, linkID, groupID string) error {
	if err := g.revokeGroupInviteLink(ctx, &group.RevokeGroupInviteLinkReq{
		LinkID:  linkID,
		GroupID: groupID,
	}); err != nil {
		return err
	}
	return g.refreshLocalGroupInviteLink(ctx, groupID)
}

// ListGroupInviteLinks 分页查询群内邀请链接列表（HTTP POST /group/list_invite_links）。
func (g *Group) ListGroupInviteLinks(ctx context.Context, groupID string, pageNumber, showNumber int32) (*group.ListGroupInviteLinksResp, error) {
	return g.listGroupInviteLinks(ctx, &group.ListGroupInviteLinksReq{
		GroupID: groupID,
		Pagination: &sdkws.RequestPagination{
			PageNumber: pageNumber,
			ShowNumber: showNumber,
		},
	})
}

// SetInviteSetting 设置群成员邀请他人入群权限（HTTP POST /group/set_invite_setting）：
//
//	allowAddMember 0 = 全员可邀请，1 = 仅群主/管理员可邀请
func (g *Group) SetInviteSetting(ctx context.Context, groupID string, allowAddMember int32) error {
	if err := g.setInviteSetting(ctx, &api.SetInviteSettingReq{
		GroupID:        groupID,
		AllowAddMember: allowAddMember,
	}); err != nil {
		return err
	}
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()
	return g.IncrSyncJoinGroup(ctx)
}

// GetInviteSetting 查询群邀请权限（委托 GetGroupSetting，成功回调仅含 groupID、allowAddMember）。
func (g *Group) GetInviteSetting(ctx context.Context, groupID string) (*api.GetInviteSettingResp, error) {
	return g.getInviteSetting(ctx, groupID)
}

// SetPinSetting 设置群成员置顶消息权限（HTTP POST /group/set_pin_setting）：
//
//	allowPinMsg 0 = 全员可置顶，1 = 仅群主/管理员可置顶
func (g *Group) SetPinSetting(ctx context.Context, groupID string, allowPinMsg int32) error {
	if err := g.setPinSetting(ctx, &api.SetPinSettingReq{
		GroupID:     groupID,
		AllowPinMsg: allowPinMsg,
	}); err != nil {
		return err
	}
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()
	return g.IncrSyncJoinGroup(ctx)
}

// GetPinSetting 查询群置顶消息权限（委托 GetGroupSetting，成功回调仅含 groupID、allowPinMsg）。
func (g *Group) GetPinSetting(ctx context.Context, groupID string) (*api.GetPinSettingResp, error) {
	return g.getPinSetting(ctx, groupID)
}

// SetEditSetting 设置群成员编辑群资料权限（HTTP POST /group/set_edit_setting）：
//
//	allowEditGroupInfo 0 = 全员可编辑，1 = 仅群主/管理员可编辑
func (g *Group) SetEditSetting(ctx context.Context, groupID string, allowEditGroupInfo int32) error {
	if err := g.setEditSetting(ctx, &api.SetEditSettingReq{
		GroupID:            groupID,
		AllowEditGroupInfo: allowEditGroupInfo,
	}); err != nil {
		return err
	}
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()
	return g.IncrSyncJoinGroup(ctx)
}

// GetEditSetting 查询群编辑资料权限（委托 GetGroupSetting，成功回调仅含 groupID、allowEditGroupInfo）。
func (g *Group) GetEditSetting(ctx context.Context, groupID string) (*api.GetEditSettingResp, error) {
	return g.getEditSetting(ctx, groupID)
}

// SetBurnSetting 设置群成员阅后即焚权限（HTTP POST /group/set_burn_setting）：
//
//	allowBurn 0 = 仅群主可设置（默认），1 = 全员可设置
func (g *Group) SetBurnSetting(ctx context.Context, groupID string, allowBurn int32) error {
	if err := g.setBurnSetting(ctx, &api.SetBurnSettingReq{
		GroupID:   groupID,
		AllowBurn: allowBurn,
	}); err != nil {
		return err
	}
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()
	return g.IncrSyncJoinGroup(ctx)
}

// GetBurnSetting 查询群成员阅后即焚权限（委托 GetGroupSetting，成功回调仅含 groupID、allowMemberBurn）。
func (g *Group) GetBurnSetting(ctx context.Context, groupID string) (*api.GetBurnSettingResp, error) {
	return g.getBurnSetting(ctx, groupID)
}

// SetMsgBurnDuration 设置群消息阅后即焚时长（HTTP POST /group/set_msg_burn_duration）：
//
//	burnDuration 单位为秒，0 = 关闭
func (g *Group) SetMsgBurnDuration(ctx context.Context, groupID string, burnDuration int32) error {
	if err := g.setMsgBurnDuration(ctx, &api.SetMsgBurnDurationReq{
		GroupID:      groupID,
		BurnDuration: burnDuration,
	}); err != nil {
		return err
	}
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()
	return g.IncrSyncJoinGroup(ctx)
}

// GetMsgBurnDuration 查询群消息阅后即焚时长（HTTP POST /group/get_msg_burn_duration）。
func (g *Group) GetMsgBurnDuration(ctx context.Context, groupID string) (*api.GetMsgBurnDurationResp, error) {
	return g.getMsgBurnDuration(ctx, groupID)
}

// SetGroupAnnouncement 设置群公告（HTTP POST /group/set_group_announcement）。
func (g *Group) SetGroupAnnouncement(ctx context.Context, groupID string, notification string) error {
	if err := g.setGroupAnnouncement(ctx, &api.SetGroupAnnouncementReq{
		GroupID:      groupID,
		Notification: notification,
	}); err != nil {
		return err
	}
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()
	return g.IncrSyncJoinGroup(ctx)
}

// GetGroupAnnouncement 查询群公告（HTTP POST /group/get_group_announcement）。
func (g *Group) GetGroupAnnouncement(ctx context.Context, groupID string) (*api.GetGroupAnnouncementResp, error) {
	return g.getGroupAnnouncement(ctx, groupID)
}

func (g *Group) SetGroupMemberInfo(ctx context.Context, groupMemberInfo *group.SetGroupMemberInfo) error {
	req := &group.SetGroupMemberInfoReq{Members: []*group.SetGroupMemberInfo{groupMemberInfo}}
	if err := g.setGroupMemberInfo(ctx, req); err != nil {
		return err
	}

	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()

	return g.IncrSyncGroupAndMember(ctx, groupMemberInfo.GroupID)
}

func (g *Group) GetJoinedGroupList(ctx context.Context) ([]*model_struct.LocalGroup, error) {
	return g.db.GetJoinedGroupListDB(ctx)
}

func (g *Group) GetJoinedGroupListPage(ctx context.Context, offset, count int32) ([]*model_struct.LocalGroup, error) {
	dataFetcher := datafetcher.NewDataFetcher(
		g.db,
		g.groupTableName(),
		g.loginUserID,
		func(localGroup *model_struct.LocalGroup) string {
			return localGroup.GroupID
		},
		func(ctx context.Context, values []*model_struct.LocalGroup) error {
			return g.db.BatchInsertGroup(ctx, values)
		},
		func(ctx context.Context, groupIDs []string) ([]*model_struct.LocalGroup, bool, error) {
			localGroups, err := g.db.GetGroups(ctx, groupIDs)
			return localGroups, true, err
		},
		func(ctx context.Context, groupIDs []string) ([]*model_struct.LocalGroup, error) {
			serverGroupInfo, err := g.getGroupsInfoFromServer(ctx, groupIDs)
			if err != nil {
				return nil, err
			}
			return datautil.Batch(ServerGroupToLocalGroup, serverGroupInfo), nil
		},
	)
	return dataFetcher.FetchWithPagination(ctx, int(offset), int(count))
}

func (g *Group) GetSpecifiedGroupsInfo(ctx context.Context, groupIDs []string) ([]*model_struct.LocalGroup, error) {
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()

	_, err := g.db.GetVersionSync(ctx, g.groupTableName(), g.loginUserID)
	if err != nil {
		if !errs.ErrRecordNotFound.Is(err) {
			return nil, err
		}

		err := g.IncrSyncJoinGroup(ctx)
		if err != nil {
			return nil, err
		}

	}

	dataFetcher := datafetcher.NewDataFetcher(
		g.db,
		g.groupTableName(),
		g.loginUserID,
		func(localGroup *model_struct.LocalGroup) string {
			return localGroup.GroupID
		},
		func(ctx context.Context, values []*model_struct.LocalGroup) error {
			return g.db.BatchInsertGroup(ctx, values)
		},
		func(ctx context.Context, groupIDs []string) ([]*model_struct.LocalGroup, bool, error) {
			localGroups, err := g.db.GetGroups(ctx, groupIDs)
			return localGroups, true, err
		},
		func(ctx context.Context, groupIDs []string) ([]*model_struct.LocalGroup, error) {
			serverGroupInfo, err := g.getGroupsInfoFromServer(ctx, groupIDs)
			if err != nil {
				return nil, err
			}
			return datautil.Batch(ServerGroupToLocalGroup, serverGroupInfo), nil
		},
	)
	return dataFetcher.FetchMissingAndFillLocal(ctx, groupIDs)
}

func (g *Group) SearchGroups(ctx context.Context, param sdk_params_callback.SearchGroupsParam) ([]*model_struct.LocalGroup, error) {
	if len(param.KeywordList) == 0 || (!param.IsSearchGroupName && !param.IsSearchGroupID) {
		return nil, sdkerrs.ErrArgs.WrapMsg("keyword is null or search field all false")
	}
	groups, err := g.db.GetAllGroupInfoByGroupIDOrGroupName(ctx, param.KeywordList[0], param.IsSearchGroupID, param.IsSearchGroupName) // todo	param.KeywordList[0]
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (g *Group) GetGroupMemberOwnerAndAdmin(ctx context.Context, groupID string) ([]*model_struct.LocalGroupMember, error) {
	return g.db.GetGroupMemberOwnerAndAdminDB(ctx, groupID)
}

func (g *Group) GetGroupMemberListByJoinTimeFilter(ctx context.Context, groupID string, offset, count int32, joinTimeBegin, joinTimeEnd int64, userIDs []string) ([]*model_struct.LocalGroupMember, error) {
	if joinTimeEnd == 0 {
		joinTimeEnd = time.Now().UnixMilli()
	}

	dataFetcher := datafetcher.NewDataFetcher(
		g.db,
		g.groupAndMemberVersionTableName(),
		groupID,
		func(localGroupMember *model_struct.LocalGroupMember) string {
			return localGroupMember.UserID
		},
		func(ctx context.Context, values []*model_struct.LocalGroupMember) error {
			return g.db.BatchInsertGroupMember(ctx, values)
		},
		func(ctx context.Context, userIDs []string) ([]*model_struct.LocalGroupMember, bool, error) {
			localGroupMembers, err := g.db.GetGroupMemberListSplitByJoinTimeFilter(ctx, groupID, int(offset), int(count), joinTimeBegin, joinTimeEnd, userIDs)
			return localGroupMembers, true, err
		},
		func(ctx context.Context, userIDs []string) ([]*model_struct.LocalGroupMember, error) {
			serverGroupMember, err := g.getDesignatedGroupMembers(ctx, groupID, userIDs)
			if err != nil {
				return nil, err
			}
			return datautil.Batch(ServerGroupMemberToLocalGroupMember, serverGroupMember), nil
		},
	)

	return dataFetcher.FetchWithPagination(ctx, int(offset), int(count))
}

func (g *Group) GetSpecifiedGroupMembersInfo(ctx context.Context, groupID string, userIDList []string) ([]*model_struct.LocalGroupMember, error) {
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()

	lvs, err := g.db.GetVersionSync(ctx, g.groupTableName(), g.loginUserID)
	if err != nil {
		return nil, err
	}
	if datautil.Contain(groupID, lvs.UIDList...) {

		_, err := g.db.GetVersionSync(ctx, g.groupAndMemberVersionTableName(), groupID)
		if err != nil {
			if !errs.ErrRecordNotFound.Is(err) {
				return nil, err
			}
			err := g.IncrSyncGroupAndMember(ctx, groupID)
			if err != nil {
				return nil, err
			}
		}
	} else { // If the user is no longer in the group, return nil immediately
		return nil, nil
	}
	dataFetcher := datafetcher.NewDataFetcher(
		g.db,
		g.groupAndMemberVersionTableName(),
		groupID,
		func(localGroupMember *model_struct.LocalGroupMember) string {
			return localGroupMember.UserID
		},
		func(ctx context.Context, values []*model_struct.LocalGroupMember) error {
			return g.db.BatchInsertGroupMember(ctx, values)
		},
		func(ctx context.Context, userIDs []string) ([]*model_struct.LocalGroupMember, bool, error) {
			localGroupMembers, err := g.db.GetGroupSomeMemberInfo(ctx, groupID, userIDList)
			if err != nil {
				return nil, false, err
			}
			localGroup, err := g.db.GetGroupInfoByGroupID(ctx, groupID)
			if err != nil {
				return nil, false, err
			}
			if localGroup.MemberCount < groupMemberSyncLimit {
				return localGroupMembers, false, nil
			}
			return localGroupMembers, true, nil
		},
		func(ctx context.Context, userIDs []string) ([]*model_struct.LocalGroupMember, error) {
			serverGroupMember, err := g.getDesignatedGroupMembers(ctx, groupID, userIDs)
			if err != nil {
				return nil, err
			}
			if len(serverGroupMember) == 0 {
				return nil, nil
			}
			return datautil.Batch(ServerGroupMemberToLocalGroupMember, serverGroupMember), nil
		},
	)
	return dataFetcher.FetchMissingAndFillLocal(ctx, userIDList)
}

func (g *Group) GetGroupMemberList(ctx context.Context, groupID string, filter, offset, count int32) ([]*model_struct.LocalGroupMember, error) {
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()

	lvs, err := g.db.GetVersionSync(ctx, g.groupTableName(), g.loginUserID)
	if err != nil {
		return nil, err
	}
	if datautil.Contain(groupID, lvs.UIDList...) {

		_, err := g.db.GetVersionSync(ctx, g.groupAndMemberVersionTableName(), groupID)
		if err != nil {
			if !errs.ErrRecordNotFound.Is(err) {
				return nil, err
			}
			err := g.IncrSyncGroupAndMember(ctx, groupID)
			if err != nil {
				return nil, err
			}
		}
	} else { // If the user is no longer in the group, return nil immediately
		return nil, nil
	}

	dataFetcher := datafetcher.NewDataFetcher(
		g.db,
		g.groupAndMemberVersionTableName(),
		groupID,
		func(localGroupMember *model_struct.LocalGroupMember) string {
			return localGroupMember.UserID
		},
		func(ctx context.Context, values []*model_struct.LocalGroupMember) error {
			return g.db.BatchInsertGroupMember(ctx, values)
		},
		func(ctx context.Context, userIDs []string) ([]*model_struct.LocalGroupMember, bool, error) {
			localGroupMembers, err := g.db.GetGroupMemberListByUserIDs(ctx, groupID, filter, userIDs)
			if err != nil {
				return nil, false, err
			}
			switch filter {
			case constant.GroupFilterOwner:
				fallthrough
			case constant.GroupFilterAdmin:
				fallthrough
			case constant.GroupFilterOwnerAndAdmin:
				return localGroupMembers, false, nil
			case constant.GroupFilterAll:
				fallthrough
			case constant.GroupFilterOrdinaryUsers:
				fallthrough
			case constant.GroupFilterAdminAndOrdinaryUsers:
				return localGroupMembers, true, nil
			}
			return nil, false, sdkerrs.ErrArgs
		},
		func(ctx context.Context, userIDs []string) ([]*model_struct.LocalGroupMember, error) {
			serverGroupMember, err := g.getDesignatedGroupMembers(ctx, groupID, userIDs)
			if err != nil {
				return nil, err
			}
			return datautil.Batch(ServerGroupMemberToLocalGroupMember, serverGroupMember), nil
		},
	)
	switch filter {
	case constant.GroupFilterOrdinaryUsers:
		groupOwnerAndGroupMember, err := g.db.GetGroupMemberListSplit(ctx, groupID, constant.GroupFilterOwnerAndAdmin, 0, 100)
		if err != nil {
			return nil, err
		}
		offset = offset + int32(len(groupOwnerAndGroupMember))
	case constant.GroupFilterAdminAndOrdinaryUsers:
		groupOwnerAndGroupMember, err := g.db.GetGroupMemberListSplit(ctx, groupID, constant.GroupFilterOwner, 0, 100)
		if err != nil {
			return nil, err
		}
		offset = offset + int32(len(groupOwnerAndGroupMember))
	}
	return dataFetcher.FetchWithPagination(ctx, int(offset), int(count))
}

func (g *Group) GetGroupApplicationListAsRecipient(ctx context.Context, req *sdk_params_callback.GetGroupApplicationListAsRecipientReq) ([]*model_struct.LocalGroupRequest, error) {
	groupRequests, err := g.getServerAdminGroupApplicationList(ctx, req.GroupIDs, req.HandleResults, utils.GetPageNumber(req.Offset, req.Count), req.Count)
	if err != nil {
		return nil, err
	}
	return datautil.Batch(ServerGroupRequestToLocalGroupRequest, groupRequests), nil
}

func (g *Group) GetGroupApplicationListAsApplicant(ctx context.Context, req *sdk_params_callback.GetGroupApplicationListAsApplicantReq) ([]*model_struct.LocalGroupRequest, error) {
	groupRequests, err := g.getServerSelfGroupApplication(ctx, req.GroupIDs, req.HandleResults, utils.GetPageNumber(req.Offset, req.Count), req.Count)
	if err != nil {
		return nil, err
	}
	return datautil.Batch(ServerGroupRequestToLocalGroupRequest, groupRequests), nil
}

func (g *Group) SearchGroupMembers(ctx context.Context, searchParam *sdk_params_callback.SearchGroupMembersParam) ([]*model_struct.LocalGroupMember, error) {
	return g.db.SearchGroupMembersDB(ctx, searchParam.KeywordList[0], searchParam.GroupID, searchParam.IsSearchMemberNickname, searchParam.IsSearchUserID, searchParam.Offset, searchParam.Count)
}

func (g *Group) IsJoinGroup(ctx context.Context, groupID string) (bool, error) {
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()

	lvs, err := g.db.GetVersionSync(ctx, g.groupTableName(), g.loginUserID)
	if err != nil {
		return false, err
	}
	if datautil.Contain(groupID, lvs.UIDList...) {
		return true, nil
	}
	return false, nil
}

func (g *Group) GetUsersInGroup(ctx context.Context, groupID string, userIDList []string) ([]string, error) {
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()

	lvs, err := g.db.GetVersionSync(ctx, g.groupTableName(), g.loginUserID)
	if err != nil {
		return nil, err
	}
	if !datautil.Contain(groupID, lvs.UIDList...) {
		return nil, nil
	}
	lvs, err = g.db.GetVersionSync(ctx, g.groupAndMemberVersionTableName(), groupID)
	if err != nil {
		return nil, err
	}

	groupMembersMap := datautil.SliceSetAny(lvs.UIDList, func(e string) string {
		return e
	})

	var usersInGroup []string
	for _, userID := range userIDList {
		if _, exists := groupMembersMap[userID]; exists {
			usersInGroup = append(usersInGroup, userID)
		}
	}

	return usersInGroup, nil
}

func (g *Group) InviteUserToGroup(ctx context.Context, groupID, reason string, userIDList []string) error {
	req := &group.InviteUserToGroupReq{GroupID: groupID, Reason: reason, InvitedUserIDs: userIDList}
	if err := g.inviteUserToGroup(ctx, req); err != nil {
		return err
	}

	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()

	if err := g.IncrSyncGroupAndMember(ctx, groupID); err != nil {
		return err
	}
	return nil
}

func (g *Group) AcceptGroupApplication(ctx context.Context, groupID, fromUserID, handleMsg string) error {
	return g.HandlerGroupApplication(ctx, &group.GroupApplicationResponseReq{GroupID: groupID, FromUserID: fromUserID, HandledMsg: handleMsg, HandleResult: constant.GroupResponseAgree})
}

func (g *Group) RefuseGroupApplication(ctx context.Context, groupID, fromUserID, handleMsg string) error {
	return g.HandlerGroupApplication(ctx, &group.GroupApplicationResponseReq{GroupID: groupID, FromUserID: fromUserID, HandledMsg: handleMsg, HandleResult: constant.GroupResponseRefuse})
}

func (g *Group) HandlerGroupApplication(ctx context.Context, req *group.GroupApplicationResponseReq) error {
	if err := g.handlerGroupApplication(ctx, req); err != nil {
		return err
	}
	return nil
}

func (g *Group) GetGroupMemberNameAndFaceURL(ctx context.Context, groupID string, userIDs []string) (map[string]*model_struct.LocalGroupMember, error) {
	return g.GetGroupMembersInfo(ctx, groupID, userIDs)
}

func (g *Group) GetGroupApplicationUnhandledCount(ctx context.Context, req *sdk_params_callback.GetGroupApplicationUnhandledCountReq) (int32, error) {
	return g.getGroupApplicationUnhandledCount(ctx, req.Time)
}

func (g *Group) GetCommonGroupsWithFriend(ctx context.Context, req *sdk_params_callback.GetCommonGroupsWithFriendReq) (*group.GetCommonGroupsWithFriendResp, error) {
	return g.getCommonGroupsWithFriend(ctx, req.FriendUserID)
}

func (g *Group) PinGroupMessage(ctx context.Context, req *sdk_params_callback.PinGroupMessageReq) (*group.PinGroupMessageResp, error) {
	return g.pinGroupMessage(ctx, &group.PinGroupMessageReq{GroupID: req.GroupID, Seq: req.Seq})
}

func (g *Group) UnpinGroupMessage(ctx context.Context, req *sdk_params_callback.UnpinGroupMessageReq) (*group.UnpinGroupMessageResp, error) {
	return g.unpinGroupMessage(ctx, &group.UnpinGroupMessageReq{GroupID: req.GroupID, Seq: req.Seq, PinID: req.PinID})
}

func (g *Group) GetGroupPinnedMessages(ctx context.Context, req *sdk_params_callback.GetGroupPinnedMessagesReq) (*group.GetGroupPinnedMessagesResp, error) {
	return g.getGroupPinnedMessages(ctx, req.GroupID)
}

func (g *Group) SetGroupMute(ctx context.Context, req *sdk_params_callback.SetGroupMuteReq) error {
	if err := g.setGroupMute(ctx, &group.SetGroupMuteReq{GroupID: req.GroupID, Duration: req.Duration}); err != nil {
		return err
	}
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()
	return g.IncrSyncJoinGroup(ctx)
}

func (g *Group) GetGroupMute(ctx context.Context, req *sdk_params_callback.GetGroupMuteReq) (*group.GetGroupMuteResp, error) {
	return g.getGroupMute(ctx, &group.GetGroupMuteReq{GroupID: req.GroupID})
}

func (g *Group) SetGroupBlock(ctx context.Context, req *sdk_params_callback.SetGroupBlockReq) error {
	return g.setGroupBlock(ctx, &group.SetGroupBlockReq{GroupID: req.GroupID, Block: req.Block})
}

func (g *Group) GetGroupBlock(ctx context.Context, req *sdk_params_callback.GetGroupBlockReq) (*group.GetGroupBlockResp, error) {
	return g.getGroupBlock(ctx, &group.GetGroupBlockReq{GroupID: req.GroupID})
}

func (g *Group) GetBlockGroup(ctx context.Context, _ *sdk_params_callback.GetBlockGroupReq) (*group.GetBlockGroupResp, error) {
	return g.getBlockGroup(ctx, &group.GetBlockGroupReq{})
}

func (g *Group) PinGroup(ctx context.Context, req *sdk_params_callback.PinGroupReq) error {
	if err := g.pinGroup(ctx, &group.PinGroupReq{GroupID: req.GroupID}); err != nil {
		return err
	}
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()
	if err := g.IncrSyncJoinGroup(ctx); err != nil {
		return err
	}
	if g.incrSyncConversations != nil {
		return g.incrSyncConversations(ctx)
	}
	return nil
}

func (g *Group) UnpinGroup(ctx context.Context, req *sdk_params_callback.UnpinGroupReq) error {
	if err := g.unpinGroup(ctx, &group.UnpinGroupReq{GroupID: req.GroupID}); err != nil {
		return err
	}
	g.groupSyncMutex.Lock()
	defer g.groupSyncMutex.Unlock()
	if err := g.IncrSyncJoinGroup(ctx); err != nil {
		return err
	}
	if g.incrSyncConversations != nil {
		return g.incrSyncConversations(ctx)
	}
	return nil
}
