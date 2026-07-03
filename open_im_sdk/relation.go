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

func GetSpecifiedFriendsInfo(callback open_im_sdk_callback.Base, operationID string, userIDList string, filterBlack bool) {
	call(callback, operationID, UserForSDK.Relation().GetSpecifiedFriendsInfo, userIDList, filterBlack)
}

func GetFriendList(callback open_im_sdk_callback.Base, operationID string, filterBlack bool) {
	call(callback, operationID, UserForSDK.Relation().GetFriendList, filterBlack)
}

func GetFriendListPage(callback open_im_sdk_callback.Base, operationID string, offset int32, count int32, filterBlack bool) {
	call(callback, operationID, UserForSDK.Relation().GetFriendListPage, offset, count, filterBlack)
}

func SearchFriends(callback open_im_sdk_callback.Base, operationID string, searchParam string) {
	call(callback, operationID, UserForSDK.Relation().SearchFriends, searchParam)
}

// SearchFriendsByProfile 按备注、firstName+lastName、nickname 模糊搜索好友；searchParam 为 JSON，见 sdk_params_callback.SearchFriendsByProfileParam。
func SearchFriendsByProfile(callback open_im_sdk_callback.Base, operationID string, searchParam string) {
	call(callback, operationID, UserForSDK.Relation().SearchFriendsByProfile, searchParam)
}

func CheckFriend(callback open_im_sdk_callback.Base, operationID string, userIDList string) {
	call(callback, operationID, UserForSDK.Relation().CheckFriend, userIDList)
}

func AddFriend(callback open_im_sdk_callback.Base, operationID string, userIDReqMsg string) {
	call(callback, operationID, UserForSDK.Relation().AddFriend, userIDReqMsg)
}

// AddOnewayFriend adds the specified user to the caller's friend list immediately,
// without sending a friend request or requiring the target user's approval.
// The target user's friend list is NOT updated.
func AddOnewayFriend(callback open_im_sdk_callback.Base, operationID string, userIDReqMsg string) {
	call(callback, operationID, UserForSDK.Relation().AddOnewayFriend, userIDReqMsg)
}

func UpdateFriends(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Relation().UpdateFriends, req)
}

// SetFriendName sets the owner's custom firstName/lastName (and optionally note)
// for a friend; req is JSON (SetFriendNameReq). firstName/lastName/note are all
// optional: omit a field to leave it unchanged, e.g. {"friendUserID":"x","note":"vip"}
// updates only the note. Passing note here is equivalent to calling SetFriendNote.
func SetFriendName(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Relation().SetFriendName, req)
}

// SetFriendNote sets the owner's private note for a friend; req is JSON (relation.SetFriendNoteReq).
// The note is independent of remark and does not affect the friend's display name.
// Implemented via /friend/set_friend_name (the former /friend/set_note route was merged there).
func SetFriendNote(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Relation().SetFriendNote, req)
}

func DeleteFriend(callback open_im_sdk_callback.Base, operationID string, friendUserID string) {
	call(callback, operationID, UserForSDK.Relation().DeleteFriend, friendUserID)
}

// DeleteFriendOneway removes the friend only on the caller's side; the peer still has the caller as friend.
func DeleteFriendOneway(callback open_im_sdk_callback.Base, operationID string, friendUserID string) {
	call(callback, operationID, UserForSDK.Relation().DeleteFriendOneway, friendUserID)
}

func GetFriendApplicationListAsRecipient(callback open_im_sdk_callback.Base, operationID, req string) {
	call(callback, operationID, UserForSDK.Relation().GetFriendApplicationListAsRecipient, req)
}

func GetFriendApplicationListAsApplicant(callback open_im_sdk_callback.Base, operationID, req string) {
	call(callback, operationID, UserForSDK.Relation().GetFriendApplicationListAsApplicant, req)
}

func AcceptFriendApplication(callback open_im_sdk_callback.Base, operationID string, userIDHandleMsg string) {
	call(callback, operationID, UserForSDK.Relation().AcceptFriendApplication, userIDHandleMsg)
}

func RefuseFriendApplication(callback open_im_sdk_callback.Base, operationID string, userIDHandleMsg string) {
	call(callback, operationID, UserForSDK.Relation().RefuseFriendApplication, userIDHandleMsg)
}

func AddBlack(callback open_im_sdk_callback.Base, operationID string, blackUserID string, ex string) {
	call(callback, operationID, UserForSDK.Relation().AddBlack, blackUserID, ex)
}

func GetBlackList(callback open_im_sdk_callback.Base, operationID string) {
	call(callback, operationID, UserForSDK.Relation().GetBlackList)
}

func RemoveBlack(callback open_im_sdk_callback.Base, operationID string, removeUserID string) {
	call(callback, operationID, UserForSDK.Relation().RemoveBlack, removeUserID)
}

func GetFriendApplicationUnhandledCount(callback open_im_sdk_callback.Base, operationID, req string) {
	call(callback, operationID, UserForSDK.Relation().GetFriendApplicationUnhandledCount, req)
}

// SetFriendMute 设置对好友（或单聊对象）的消息免打扰；req 为 JSON，字段见 sdk_params_callback.SetFriendMuteReq。
func SetFriendMute(callback open_im_sdk_callback.Base, operationID, req string) {
	call(callback, operationID, UserForSDK.Relation().SetFriendMute, req)
}

// GetFriendMute 查询免打扰状态；req 为 JSON，字段见 sdk_params_callback.GetFriendMuteReq。
func GetFriendMute(callback open_im_sdk_callback.Base, operationID, req string) {
	call(callback, operationID, UserForSDK.Relation().GetFriendMute, req)
}

// GetFriendPhone 查询好友手机号与区号（受对方可见性限制）；friendUserID 为对方用户 ID。
func GetFriendPhone(callback open_im_sdk_callback.Base, operationID string, friendUserID string) {
	call(callback, operationID, UserForSDK.Relation().GetFriendPhone, friendUserID)
}

// PinFriend 置顶好友会话；req 为 JSON，字段见 sdk_params_callback.PinFriendReq。
func PinFriend(callback open_im_sdk_callback.Base, operationID, req string) {
	call(callback, operationID, UserForSDK.Relation().PinFriend, req)
}

// UnpinFriend 取消置顶好友会话；req 为 JSON，字段见 sdk_params_callback.UnpinFriendReq。
func UnpinFriend(callback open_im_sdk_callback.Base, operationID, req string) {
	call(callback, operationID, UserForSDK.Relation().UnpinFriend, req)
}
