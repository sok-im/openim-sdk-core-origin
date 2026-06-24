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

import (
	"github.com/openimsdk/openim-sdk-core/v3/open_im_sdk_callback"
)

func GetUsersInfo(callback open_im_sdk_callback.Base, operationID string, userIDs string) {
	call(callback, operationID, UserForSDK.User().GetUsersInfo, userIDs)
}

// SetSelfInfo sets the user's own information.
func SetSelfInfo(callback open_im_sdk_callback.Base, operationID string, userInfo string) {
	call(callback, operationID, UserForSDK.User().SetSelfInfo, userInfo)
}

//// SetSelfInfo sets the user's own information with Ex field.
//func SetSelfInfo(callback open_im_sdk_callback.Base, operationID string, userInfo string) {
//	call(callback, operationID, UserForSDK.User().SetSelfInfo, userInfo)
//}

// GetSelfUserInfo obtains the user's own information.
func GetSelfUserInfo(callback open_im_sdk_callback.Base, operationID string) {
	call(callback, operationID, UserForSDK.User().GetSelfUserInfo)
}

// AddUserCommand add to user's favorite
func AddUserCommand(callback open_im_sdk_callback.Base, operationID string, Type int32, uuid string, value string) {
	call(callback, operationID, UserForSDK.User().ProcessUserCommandAdd, Type, uuid, value)
}

// DeleteUserCommand delete from user's favorite
func DeleteUserCommand(callback open_im_sdk_callback.Base, operationID string, Type int32, uuid string) {
	call(callback, operationID, UserForSDK.User().ProcessUserCommandDelete, Type, uuid)
}

// GetAllUserCommands get user's favorite
func GetAllUserCommands(callback open_im_sdk_callback.Base, operationID string, Type int32) {
	call(callback, operationID, UserForSDK.User().ProcessUserCommandGetAll, Type)
}

// GetActiveDevices 获取当前登录用户的所有活跃设备列表
func GetActiveDevices(callback open_im_sdk_callback.Base, operationID string) {
	call(callback, operationID, UserForSDK.User().GetActiveDevices)
}

// KickDevice 将当前登录用户的指定平台设备踢下线，platformID 参考平台常量定义
func KickDevice(callback open_im_sdk_callback.Base, operationID string, platformID int32) {
	call(callback, operationID, UserForSDK.User().KickDevice, platformID)
}

// SetPhoneVisibility 设置手机号及其可见性（phone 可为空表示只改策略）：
//   phoneVisibility 0=所有人可见 1=仅好友可见 2=隐藏
func SetPhoneVisibility(callback open_im_sdk_callback.Base, operationID string, phone string, phoneVisibility int32) {
	call(callback, operationID, UserForSDK.User().SetPhoneVisibility, phone, phoneVisibility)
}

// SetCallAcceptSetting 设置音视频通话接受权限：
//   callAcceptSetting 0=所有人可发起 1=仅好友可发起 2=不接受任何通话
func SetCallAcceptSetting(callback open_im_sdk_callback.Base, operationID string, callAcceptSetting int32) {
	call(callback, operationID, UserForSDK.User().SetCallAcceptSetting, callAcceptSetting)
}

// SetMsgReceiveSetting 设置会话消息接收权限：
//   msgReceiveSetting 0=所有人可发 1=仅好友可发 2=所有人不可发
func SetMsgReceiveSetting(callback open_im_sdk_callback.Base, operationID string, msgReceiveSetting int32) {
	call(callback, operationID, UserForSDK.User().SetMsgReceiveSetting, msgReceiveSetting)
}

// SetGroupInviteSetting 设置群邀请权限：
//   groupInviteSetting 0=所有人可邀请 1=仅好友可邀请 2=所有人不可邀请
func SetGroupInviteSetting(callback open_im_sdk_callback.Base, operationID string, groupInviteSetting int32) {
	call(callback, operationID, UserForSDK.User().SetGroupInviteSetting, groupInviteSetting)
}

// SetUserMsgBurnDuration 设置用户全局阅后即焚时长（秒），0 表示关闭。
func SetUserMsgBurnDuration(callback open_im_sdk_callback.Base, operationID string, msgBurnDuration int32) {
	call(callback, operationID, UserForSDK.User().SetUserMsgBurnDuration, msgBurnDuration)
}

// SetDeleteAccountInterval 设置删除账号等待间隔（秒）；0 表示使用系统默认（18 个月）。对应 HTTP POST /user/set_delete_account_interval。
func SetDeleteAccountInterval(callback open_im_sdk_callback.Base, operationID string, deleteAccountIntervalSec int32) {
	call(callback, operationID, UserForSDK.User().SetDeleteAccountInterval, deleteAccountIntervalSec)
}

// GetUserPrivacySettings 获取当前登录用户的隐私与接收相关设置（对应 HTTP POST /user/get_user_privacy_settings）。
// 成功回调 JSON 为 openim.user.GetUserPrivacySettingsResp（msgBurnDuration、phoneVisibility、callAcceptSetting、globalRecvMsgOpt、msgReceiveSetting、groupInviteSetting）。
func GetUserPrivacySettings(callback open_im_sdk_callback.Base, operationID string) {
	call(callback, operationID, UserForSDK.User().GetUserPrivacySettings)
}

// SetMsgNotificationSwitch 设置消息通知开关：enable true=打开 false=关闭。
func SetMsgNotificationSwitch(callback open_im_sdk_callback.Base, operationID string, enable bool) {
	call(callback, operationID, UserForSDK.User().SetMsgNotificationSwitch, enable)
}

// SetSokimPaymentNotificationSwitch 设置 sokim 支付通知开关：enable true=打开 false=关闭。
func SetSokimPaymentNotificationSwitch(callback open_im_sdk_callback.Base, operationID string, enable bool) {
	call(callback, operationID, UserForSDK.User().SetSokimPaymentNotificationSwitch, enable)
}

// SetSokimServiceNotificationSwitch 设置 sokim 服务通知开关：enable true=打开 false=关闭。
func SetSokimServiceNotificationSwitch(callback open_im_sdk_callback.Base, operationID string, enable bool) {
	call(callback, operationID, UserForSDK.User().SetSokimServiceNotificationSwitch, enable)
}

// SetAvNotificationSwitch 设置音视频通知开关：enable true=打开 false=关闭。
func SetAvNotificationSwitch(callback open_im_sdk_callback.Base, operationID string, enable bool) {
	call(callback, operationID, UserForSDK.User().SetAvNotificationSwitch, enable)
}

// SetAvCallRingtoneSwitch 设置音视频来电铃声开关：enable true=打开 false=关闭。
func SetAvCallRingtoneSwitch(callback open_im_sdk_callback.Base, operationID string, enable bool) {
	call(callback, operationID, UserForSDK.User().SetAvCallRingtoneSwitch, enable)
}

// SetPlayCalleeRingtoneOnAnswerSwitch 设置接听时播放对方铃声开关：enable true=打开 false=关闭。
func SetPlayCalleeRingtoneOnAnswerSwitch(callback open_im_sdk_callback.Base, operationID string, enable bool) {
	call(callback, operationID, UserForSDK.User().SetPlayCalleeRingtoneOnAnswerSwitch, enable)
}

// GetUserNotificationSettings 获取当前登录用户的通知相关开关（对应 HTTP POST /user/get_user_notification_settings）。
// 成功回调 JSON 为 openim.user.GetUserNotificationSettingsResp。
func GetUserNotificationSettings(callback open_im_sdk_callback.Base, operationID string) {
	call(callback, operationID, UserForSDK.User().GetUserNotificationSettings)
}

// GetUserByPhone 根据手机号精确查询用户，成功回调 JSON 为 sdkws.UserInfo 或 null（未找到）
func GetUserByPhone(callback open_im_sdk_callback.Base, operationID string, phone string) {
	call(callback, operationID, UserForSDK.User().GetUserByPhone, phone)
}

// GetUsersByNickname 根据昵称精确查询用户，成功回调 JSON 为 []*sdkws.UserInfo（可能为空数组）
func GetUsersByNickname(callback open_im_sdk_callback.Base, operationID string, nickname string) {
	call(callback, operationID, UserForSDK.User().GetUsersByNickname, nickname)
}

// CheckNickname 检查昵称是否已被占用，成功回调 JSON 为 bool。
// excludeUserID 修改昵称时传入当前用户 ID 以排除本人；注册场景可传空字符串。
func CheckNickname(callback open_im_sdk_callback.Base, operationID string, nickname, excludeUserID string) {
	call(callback, operationID, UserForSDK.User().CheckNickname, nickname, excludeUserID)
}

// CheckUserExist 检查指定 userID 的用户是否存在，成功回调 JSON 为 bool。
func CheckUserExist(callback open_im_sdk_callback.Base, operationID string, userID string) {
	call(callback, operationID, UserForSDK.User().CheckUserExist, userID)
}
