package user

import (
	"context"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/common"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/constant"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/db/model_struct"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/utils"
	"github.com/openimsdk/openim-sdk-core/v3/sdk_struct"
	authPb "github.com/openimsdk/protocol/auth"
	"github.com/openimsdk/protocol/sdkws"
	userPb "github.com/openimsdk/protocol/user"
	"github.com/openimsdk/tools/log"
	"github.com/openimsdk/tools/utils/datautil"
)

// ProcessUserCommandGetAll get user's choice
func (u *User) ProcessUserCommandGetAll(ctx context.Context) ([]*userPb.CommandInfoResp, error) {
	localCommands, err := u.DataBase.ProcessUserCommandGetAll(ctx)
	if err != nil {
		return nil, err // Handle the error appropriately
	}
	var result []*userPb.CommandInfoResp
	for _, localCommand := range localCommands {
		result = append(result, &userPb.CommandInfoResp{
			Type:       localCommand.Type,
			CreateTime: localCommand.CreateTime,
			Uuid:       localCommand.Uuid,
			Value:      localCommand.Value,
		})
	}
	return result, nil
}

func (u *User) UserOnlineStatusChange(users map[string][]int32) {
	for userID, onlinePlatformIDs := range users {
		status := userPb.OnlineStatus{
			UserID:      userID,
			PlatformIDs: onlinePlatformIDs,
		}
		if len(status.PlatformIDs) == 0 {
			status.Status = constant.Offline
		} else {
			status.Status = constant.Online
		}
		u.listener().OnUserStatusChanged(utils.StructToJsonString(&status))
	}
}

func (u *User) GetSelfUserInfo(ctx context.Context) (*model_struct.LocalUser, error) {
	return u.GetUserInfoWithCache(ctx, u.loginUserID)
}

func (u *User) SetSelfInfo(ctx context.Context, userInfo *sdkws.UserInfoWithEx) error {
	// updateSelfUserInfo updates the user's information with Ex field.
	userInfo.UserID = u.loginUserID
	if err := u.updateUserInfo(ctx, userInfo); err != nil {
		return err
	}
	err := u.SyncLoginUserInfo(ctx)
	if err != nil {
		log.ZWarn(ctx, "SyncLoginUserInfo", err)
	}
	return nil
}

// ProcessUserCommandAdd CRUD user command
func (u *User) ProcessUserCommandAdd(ctx context.Context, userCommand *userPb.ProcessUserCommandAddReq) error {
	req := &userPb.ProcessUserCommandAddReq{UserID: u.loginUserID, Type: userCommand.Type, Uuid: userCommand.Uuid, Value: userCommand.Value}
	if err := u.processUserCommandAdd(ctx, req); err != nil {
		return err
	}
	return u.SyncAllCommand(ctx)
}

// ProcessUserCommandDelete delete user's choice
func (u *User) ProcessUserCommandDelete(ctx context.Context, userCommand *userPb.ProcessUserCommandDeleteReq) error {
	req := &userPb.ProcessUserCommandDeleteReq{UserID: u.loginUserID, Type: userCommand.Type, Uuid: userCommand.Uuid}
	if err := u.processUserCommandDelete(ctx, req); err != nil {
		return err
	}
	return u.SyncAllCommand(ctx)
}

// ProcessUserCommandUpdate update user's choice
func (u *User) ProcessUserCommandUpdate(ctx context.Context, userCommand *userPb.ProcessUserCommandUpdateReq) error {
	req := &userPb.ProcessUserCommandUpdateReq{UserID: u.loginUserID, Type: userCommand.Type, Uuid: userCommand.Uuid, Value: userCommand.Value}
	if err := u.processUserCommandUpdate(ctx, req); err != nil {
		return err
	}
	return u.SyncAllCommand(ctx)
}

// GetActiveDevices 获取当前用户所有在线设备信息
func (u *User) GetActiveDevices(ctx context.Context) ([]*authPb.DeviceInfo, error) {
	return u.getActiveDevices(ctx)
}

// KickDevice 将当前用户的指定平台设备踢下线
func (u *User) KickDevice(ctx context.Context, platformID int32) error {
	return u.kickDevice(ctx, platformID)
}

// SetPhoneVisibility 设置手机号及其可见性策略：
//
//	0 = 所有人可见，1 = 仅好友可见，2 = 隐藏
//
// phone 可为空（只修改可见性但不更新手机号）。
func (u *User) SetPhoneVisibility(ctx context.Context, phone string, phoneVisibility int32) error {
	return u.setPhoneVisibility(ctx, &userPb.SetPhoneVisibilityReq{
		Phone:           phone,
		PhoneVisibility: phoneVisibility,
	})
}

// SetCallAcceptSetting 设置音视频通话接受权限：
//
//	0 = 所有人可发起，1 = 仅好友可发起，2 = 不接受任何通话
func (u *User) SetCallAcceptSetting(ctx context.Context, callAcceptSetting int32) error {
	return u.setCallAcceptSetting(ctx, &userPb.SetCallAcceptSettingReq{
		CallAcceptSetting: callAcceptSetting,
	})
}

// SetMsgReceiveSetting 设置会话消息接收权限：
//
//	0 = 所有人可发，1 = 仅好友可发，2 = 所有人不可发
func (u *User) SetMsgReceiveSetting(ctx context.Context, msgReceiveSetting int32) error {
	return u.setMsgReceiveSetting(ctx, &userPb.SetMsgReceiveSettingReq{
		MsgReceiveSetting: msgReceiveSetting,
	})
}

// SetGroupInviteSetting 设置群邀请权限：
//
//	0 = 所有人可邀请，1 = 仅好友可邀请，2 = 所有人不可邀请
func (u *User) SetGroupInviteSetting(ctx context.Context, groupInviteSetting int32) error {
	return u.setGroupInviteSetting(ctx, &userPb.SetGroupInviteSettingReq{
		GroupInviteSetting: groupInviteSetting,
	})
}

// SetUserMsgBurnDuration 设置用户全局阅后即焚时长（秒），0 表示关闭。
func (u *User) SetUserMsgBurnDuration(ctx context.Context, msgBurnDuration int32) error {
	if err := u.setUserMsgBurnDuration(ctx, &userPb.SetUserMsgBurnDurationReq{
		MsgBurnDuration: msgBurnDuration,
	}); err != nil {
		return err
	}
	return u.SyncLoginUserInfo(ctx)
}

// SetDeleteAccountInterval 设置删除账号等待间隔（秒）；0 表示使用系统默认（18 个月）。对应 HTTP POST /user/set_delete_account_interval。
func (u *User) SetDeleteAccountInterval(ctx context.Context, deleteAccountIntervalSec int32) error {
	return u.setDeleteAccountInterval(ctx, &userPb.SetDeleteAccountIntervalReq{
		DeleteAccountInterval: deleteAccountIntervalSec,
	})
}

// GetUserPrivacySettings 获取当前登录用户的隐私与接收相关设置（HTTP /user/get_user_privacy_settings；userID 由服务端从 token 解析）。
func (u *User) GetUserPrivacySettings(ctx context.Context) (*userPb.GetUserPrivacySettingsResp, error) {
	return u.getUserPrivacySettings(ctx)
}

// SetMsgNotificationSwitch 设置消息通知开关：enable true=打开 false=关闭。
func (u *User) SetMsgNotificationSwitch(ctx context.Context, enable bool) error {
	return u.setMsgNotificationSwitch(ctx, &userPb.SetMsgNotificationSwitchReq{
		Enable: enable,
	})
}

// SetSokimPaymentNotificationSwitch 设置 sokim 支付通知开关：enable true=打开 false=关闭。
func (u *User) SetSokimPaymentNotificationSwitch(ctx context.Context, enable bool) error {
	return u.setSokimPaymentNotificationSwitch(ctx, &userPb.SetSokimPaymentNotificationSwitchReq{
		Enable: enable,
	})
}

// SetSokimServiceNotificationSwitch 设置 sokim 服务通知开关：enable true=打开 false=关闭。
func (u *User) SetSokimServiceNotificationSwitch(ctx context.Context, enable bool) error {
	return u.setSokimServiceNotificationSwitch(ctx, &userPb.SetSokimServiceNotificationSwitchReq{
		Enable: enable,
	})
}

// SetAvNotificationSwitch 设置音视频通知开关：enable true=打开 false=关闭。
func (u *User) SetAvNotificationSwitch(ctx context.Context, enable bool) error {
	return u.setAvNotificationSwitch(ctx, &userPb.SetAvNotificationSwitchReq{
		Enable: enable,
	})
}

// SetAvCallRingtoneSwitch 设置音视频来电铃声开关：enable true=打开 false=关闭。
func (u *User) SetAvCallRingtoneSwitch(ctx context.Context, enable bool) error {
	return u.setAvCallRingtoneSwitch(ctx, &userPb.SetAvCallRingtoneSwitchReq{
		Enable: enable,
	})
}

// SetPlayCalleeRingtoneOnAnswerSwitch 设置接听时播放对方铃声开关：enable true=打开 false=关闭。
func (u *User) SetPlayCalleeRingtoneOnAnswerSwitch(ctx context.Context, enable bool) error {
	return u.setPlayCalleeRingtoneOnAnswerSwitch(ctx, &userPb.SetPlayCalleeRingtoneOnAnswerSwitchReq{
		Enable: enable,
	})
}

// GetUserNotificationSettings 获取当前登录用户的通知相关开关（HTTP /user/get_user_notification_settings）。
func (u *User) GetUserNotificationSettings(ctx context.Context) (*userPb.GetUserNotificationSettingsResp, error) {
	return u.getUserNotificationSettings(ctx)
}

// GetUserByPhone 根据手机号精确查询用户（HTTP /user/get_user_by_phone）。
// 返回 nil 表示未找到、无权限或对方隐藏；具体语义以服务端为准。
func (u *User) GetUserByPhone(ctx context.Context, phone string) (*sdkws.UserInfo, error) {
	return u.getUserByPhone(ctx, phone)
}

// GetUsersByNickname 根据用户昵称精确查询（HTTP /user/get_users_by_nickname），可多结果；无匹配时返回空切片。
func (u *User) GetUsersByNickname(ctx context.Context, nickname string) ([]*sdkws.UserInfo, error) {
	return u.getUsersByNickname(ctx, nickname)
}

// CheckNickname 检查昵称是否已被占用（HTTP /user/check_nickname）。
// excludeUserID 修改昵称时传入当前用户 ID 以排除本人；注册场景可传空字符串。
func (u *User) CheckNickname(ctx context.Context, nickname, excludeUserID string) (bool, error) {
	return u.checkNickname(ctx, nickname, excludeUserID)
}

// CheckUserExist 检查指定 userID 的用户是否存在（HTTP /user/check_user_exist）。
func (u *User) CheckUserExist(ctx context.Context, userID string) (bool, error) {
	return u.checkUserExist(ctx, userID)
}

func (u *User) GetUsersInfo(ctx context.Context, userIDs []string) ([]*sdk_struct.PublicUser, error) {
	usersInfo, err := u.GetUsersInfoWithCache(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	res := datautil.Batch(LocalUserToPublicUser, usersInfo)

	friendList, err := u.GetFriendInfoList(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	friendMap := datautil.SliceToMap(friendList, func(friend *model_struct.LocalFriend) string {
		return friend.FriendUserID
	})

	for _, userInfo := range res {

		// update single conversation

		conversation, err := u.GetConversationByUserID(ctx, userInfo.UserID)
		if err != nil {
			log.ZWarn(ctx, "GetConversationByUserID failed", err, "userInfo", usersInfo)
			continue
		}
		if conversation.ConversationID == "" {
			continue
		}
		log.ZDebug(ctx, "GetConversationByUserID", "conversation", conversation)

		// Use ConversationShowName for friends so remark / friendFirstName+friendLastName
		// take priority over the profile name. Without this, a concurrent GetUsersInfo
		// call can overwrite the show_name that syncConversationShowNameForFriend just
		// wrote with the profile firstName+lastName.
		var showname string
		if friend, ok := friendMap[userInfo.UserID]; ok {
			showname = friend.ConversationShowName()
		} else if userInfo.FirstName != "" || userInfo.LastName != "" {
			showname = userInfo.FirstName + " " + userInfo.LastName
		} else {
			showname = userInfo.Nickname
		}

		if conversation.ShowName != showname || conversation.FaceURL != userInfo.FaceURL {
			log.ZInfo(ctx, " GetUsersInfo", "conversation", conversation, "userInfo", userInfo, "showname", showname)
			_ = common.TriggerCmdUpdateConversation(ctx, common.UpdateConNode{Action: constant.UpdateConFaceUrlAndNickName,
				Args: common.SourceIDAndSessionType{SourceID: userInfo.UserID, SessionType: constant.SingleChatType, FaceURL: userInfo.FaceURL, Nickname: showname}}, u.conversationCh)
			_ = common.TriggerCmdUpdateMessage(ctx, common.UpdateMessageNode{Action: constant.UpdateMsgFaceUrlAndNickName,
				Args: common.UpdateMessageInfo{SessionType: constant.SingleChatType, UserID: userInfo.UserID, FaceURL: userInfo.FaceURL, Nickname: userInfo.Nickname}}, u.conversationCh)
		}
	}
	return res, nil
}
