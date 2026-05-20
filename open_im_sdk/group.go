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

package open_im_sdk

import "github.com/openimsdk/openim-sdk-core/v3/open_im_sdk_callback"

func CreateGroup(callback open_im_sdk_callback.Base, operationID string, groupReqInfo string) {
	call(callback, operationID, UserForSDK.Group().CreateGroup, groupReqInfo)
}

func JoinGroup(callback open_im_sdk_callback.Base, operationID string, groupID string, reqMsg string, joinSource int32, ex string) {
	call(callback, operationID, UserForSDK.Group().JoinGroup, groupID, reqMsg, joinSource, ex)
}

func QuitGroup(callback open_im_sdk_callback.Base, operationID string, groupID string) {
	call(callback, operationID, UserForSDK.Group().QuitGroup, groupID)
}

func DismissGroup(callback open_im_sdk_callback.Base, operationID string, groupID string) {
	call(callback, operationID, UserForSDK.Group().DismissGroup, groupID)
}

func ChangeGroupMute(callback open_im_sdk_callback.Base, operationID string, groupID string, isMute bool) {
	call(callback, operationID, UserForSDK.Group().ChangeGroupMute, groupID, isMute)
}

func ChangeGroupMemberMute(callback open_im_sdk_callback.Base, operationID string, groupID string, userID string, mutedSeconds int) {
	call(callback, operationID, UserForSDK.Group().ChangeGroupMemberMute, groupID, userID, mutedSeconds)
}

func TransferGroupOwner(callback open_im_sdk_callback.Base, operationID string, groupID string, newOwnerUserID string) {
	call(callback, operationID, UserForSDK.Group().TransferGroupOwner, groupID, newOwnerUserID)
}

func KickGroupMember(callback open_im_sdk_callback.Base, operationID string, groupID string, reason string, userIDList string) {
	call(callback, operationID, UserForSDK.Group().KickGroupMember, groupID, reason, userIDList)
}

func SetGroupInfo(callback open_im_sdk_callback.Base, operationID string, groupInfo string) {
	call(callback, operationID, UserForSDK.Group().SetGroupInfo, groupInfo)
}

// SetSendMessageSetting 设置群成员发消息权限（allowSendMsg：0=全员可发 1=仅群主/管理员）。对应 HTTP POST /group/set_send_message_setting。
func SetSendMessageSetting(callback open_im_sdk_callback.Base, operationID string, groupID string, allowSendMsg int32) {
	call(callback, operationID, UserForSDK.Group().SetSendMessageSetting, groupID, allowSendMsg)
}

// GetSendMessageSetting 查询群发消息权限。成功回调 JSON：groupID、allowSendMsg。对应 HTTP POST /group/get_send_message_setting。
func GetSendMessageSetting(callback open_im_sdk_callback.Base, operationID string, groupID string) {
	call(callback, operationID, UserForSDK.Group().GetSendMessageSetting, groupID)
}

// SetInviteSetting 设置群成员邀请他人入群权限（allowAddMember：0=全员 1=仅群主/管理员）。对应 HTTP POST /group/set_invite_setting。
func SetInviteSetting(callback open_im_sdk_callback.Base, operationID string, groupID string, allowAddMember int32) {
	call(callback, operationID, UserForSDK.Group().SetInviteSetting, groupID, allowAddMember)
}

// GetInviteSetting 查询群邀请权限。成功回调 JSON：groupID、allowAddMember。
func GetInviteSetting(callback open_im_sdk_callback.Base, operationID string, groupID string) {
	call(callback, operationID, UserForSDK.Group().GetInviteSetting, groupID)
}

// SetPinSetting 设置群成员置顶消息权限（allowPinMsg：0=全员 1=仅群主/管理员）。对应 HTTP POST /group/set_pin_setting。
func SetPinSetting(callback open_im_sdk_callback.Base, operationID string, groupID string, allowPinMsg int32) {
	call(callback, operationID, UserForSDK.Group().SetPinSetting, groupID, allowPinMsg)
}

// GetPinSetting 查询群置顶消息权限。成功回调 JSON：groupID、allowPinMsg。
func GetPinSetting(callback open_im_sdk_callback.Base, operationID string, groupID string) {
	call(callback, operationID, UserForSDK.Group().GetPinSetting, groupID)
}

// SetEditSetting 设置群成员编辑群资料权限（allowEditGroupInfo：0=全员 1=仅群主/管理员）。对应 HTTP POST /group/set_edit_setting。
func SetEditSetting(callback open_im_sdk_callback.Base, operationID string, groupID string, allowEditGroupInfo int32) {
	call(callback, operationID, UserForSDK.Group().SetEditSetting, groupID, allowEditGroupInfo)
}

// GetEditSetting 查询群编辑资料权限。成功回调 JSON：groupID、allowEditGroupInfo。
func GetEditSetting(callback open_im_sdk_callback.Base, operationID string, groupID string) {
	call(callback, operationID, UserForSDK.Group().GetEditSetting, groupID)
}

func SetGroupMemberInfo(callback open_im_sdk_callback.Base, operationID string, groupMemberInfo string) {
	call(callback, operationID, UserForSDK.Group().SetGroupMemberInfo, groupMemberInfo)
}

func GetJoinedGroupList(callback open_im_sdk_callback.Base, operationID string) {
	call(callback, operationID, UserForSDK.Group().GetJoinedGroupList)
}

func GetJoinedGroupListPage(callback open_im_sdk_callback.Base, operationID string, offset, count int32) {
	call(callback, operationID, UserForSDK.Group().GetJoinedGroupListPage, offset, count)
}

func GetSpecifiedGroupsInfo(callback open_im_sdk_callback.Base, operationID string, groupIDList string) {
	call(callback, operationID, UserForSDK.Group().GetSpecifiedGroupsInfo, groupIDList)
}

func SearchGroups(callback open_im_sdk_callback.Base, operationID string, searchParam string) {
	call(callback, operationID, UserForSDK.Group().SearchGroups, searchParam)
}

func GetGroupMemberOwnerAndAdmin(callback open_im_sdk_callback.Base, operationID string, groupID string) {
	call(callback, operationID, UserForSDK.Group().GetGroupMemberOwnerAndAdmin, groupID)
}

func GetGroupMemberListByJoinTimeFilter(callback open_im_sdk_callback.Base, operationID string, groupID string, offset int32, count int32, joinTimeBegin int64, joinTimeEnd int64, filterUserIDList string) {
	call(callback, operationID, UserForSDK.Group().GetGroupMemberListByJoinTimeFilter, groupID, offset, count, joinTimeBegin, joinTimeEnd, filterUserIDList)
}

func GetSpecifiedGroupMembersInfo(callback open_im_sdk_callback.Base, operationID string, groupID string, userIDList string) {
	call(callback, operationID, UserForSDK.Group().GetSpecifiedGroupMembersInfo, groupID, userIDList)
}

func GetGroupMemberList(callback open_im_sdk_callback.Base, operationID string, groupID string, filter int32, offset int32, count int32) {
	call(callback, operationID, UserForSDK.Group().GetGroupMemberList, groupID, filter, offset, count)
}

func GetGroupApplicationListAsRecipient(callback open_im_sdk_callback.Base, operationID, req string) {
	call(callback, operationID, UserForSDK.Group().GetGroupApplicationListAsRecipient, req)
}

func GetGroupApplicationListAsApplicant(callback open_im_sdk_callback.Base, operationID, req string) {
	call(callback, operationID, UserForSDK.Group().GetGroupApplicationListAsApplicant, req)
}

func SearchGroupMembers(callback open_im_sdk_callback.Base, operationID string, searchParam string) {
	call(callback, operationID, UserForSDK.Group().SearchGroupMembers, searchParam)
}

func IsJoinGroup(callback open_im_sdk_callback.Base, operationID string, groupID string) {
	call(callback, operationID, UserForSDK.Group().IsJoinGroup, groupID)
}

func GetUsersInGroup(callback open_im_sdk_callback.Base, operationID string, groupID, userIDList string) {
	call(callback, operationID, UserForSDK.Group().GetUsersInGroup, groupID, userIDList)
}

func InviteUserToGroup(callback open_im_sdk_callback.Base, operationID string, groupID string, reason string, userIDList string) {
	call(callback, operationID, UserForSDK.Group().InviteUserToGroup, groupID, reason, userIDList)
}

func AcceptGroupApplication(callback open_im_sdk_callback.Base, operationID string, groupID string, fromUserID string, handleMsg string) {
	call(callback, operationID, UserForSDK.Group().AcceptGroupApplication, groupID, fromUserID, handleMsg)
}

func RefuseGroupApplication(callback open_im_sdk_callback.Base, operationID string, groupID string, fromUserID string, handleMsg string) {
	call(callback, operationID, UserForSDK.Group().RefuseGroupApplication, groupID, fromUserID, handleMsg)
}

func GetGroupApplicationUnhandledCount(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Group().GetGroupApplicationUnhandledCount, req)
}

func GetCommonGroupsWithFriend(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Group().GetCommonGroupsWithFriend, req)
}

func PinGroupMessage(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Group().PinGroupMessage, req)
}

func UnpinGroupMessage(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Group().UnpinGroupMessage, req)
}

func GetGroupPinnedMessages(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Group().GetGroupPinnedMessages, req)
}

// SetGroupMute 设置当前用户对群的会话静音时长；req 为 JSON，字段见 sdk_params_callback.SetGroupMuteReq。
func SetGroupMute(callback open_im_sdk_callback.Base, operationID, req string) {
	call(callback, operationID, UserForSDK.Group().SetGroupMute, req)
}

// GetGroupMute 查询当前用户对群的会话静音状态；req 为 JSON，字段见 sdk_params_callback.GetGroupMuteReq。
func GetGroupMute(callback open_im_sdk_callback.Base, operationID, req string) {
	call(callback, operationID, UserForSDK.Group().GetGroupMute, req)
}

// PinGroup 将群会话置顶；req 为 JSON，字段见 sdk_params_callback.PinGroupReq。
func PinGroup(callback open_im_sdk_callback.Base, operationID, req string) {
	call(callback, operationID, UserForSDK.Group().PinGroup, req)
}

// UnpinGroup 取消群会话置顶；req 为 JSON，字段见 sdk_params_callback.UnpinGroupReq。
func UnpinGroup(callback open_im_sdk_callback.Base, operationID, req string) {
	call(callback, operationID, UserForSDK.Group().UnpinGroup, req)
}
