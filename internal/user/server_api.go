package user

import (
	"context"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/api"
	"github.com/openimsdk/protocol/auth"
	"github.com/openimsdk/protocol/sdkws"
	"github.com/openimsdk/protocol/user"
)

func (u *User) getUsersInfo(ctx context.Context, userIDs []string) ([]*sdkws.UserInfo, error) {
	req := &user.GetDesignateUsersReq{UserIDs: userIDs}
	return api.ExtractField(ctx, api.GetUsersInfo.Invoke, req, (*user.GetDesignateUsersResp).GetUsersInfo)
}

func (u *User) updateUserInfo(ctx context.Context, userInfo *sdkws.UserInfoWithEx) error {
	userInfo.UserID = u.loginUserID
	return api.UpdateUserInfoEx.Execute(ctx, &user.UpdateUserInfoExReq{UserInfo: userInfo})
}

func (u *User) processUserCommandAdd(ctx context.Context, req *user.ProcessUserCommandAddReq) error {
	return api.ProcessUserCommandAdd.Execute(ctx, req)
}

func (u *User) processUserCommandDelete(ctx context.Context, req *user.ProcessUserCommandDeleteReq) error {
	return api.ProcessUserCommandDelete.Execute(ctx, req)
}

func (u *User) processUserCommandUpdate(ctx context.Context, req *user.ProcessUserCommandUpdateReq) error {
	return api.ProcessUserCommandUpdate.Execute(ctx, req)
}

func (u *User) processUserCommandGetAll(ctx context.Context, req *user.ProcessUserCommandGetAllReq) (*user.ProcessUserCommandGetAllResp, error) {
	return api.ProcessUserCommandGetAll.Invoke(ctx, req)
}

func (u *User) getActiveDevices(ctx context.Context) ([]*auth.DeviceInfo, error) {
	req := &auth.GetActiveDevicesReq{UserID: u.loginUserID}
	return api.ExtractField(ctx, api.GetActiveDevices.Invoke, req, (*auth.GetActiveDevicesResp).GetDevices)
}

func (u *User) kickDevice(ctx context.Context, platformID int32) error {
	req := &auth.KickDeviceReq{
		UserID:     u.loginUserID,
		PlatformID: platformID,
	}
	return api.KickDevice.Execute(ctx, req)
}

// setPhoneVisibility 调服务端设置手机号及可见性（0=所有人 1=仅好友 2=隐藏）
func (u *User) setPhoneVisibility(ctx context.Context, req *user.SetPhoneVisibilityReq) error {
	req.UserID = u.loginUserID
	return api.SetPhoneVisibility.Execute(ctx, req)
}

// setCallAcceptSetting 调服务端设置音视频通话接受权限（0=所有人 1=仅好友 2=不接受任何通话）
func (u *User) setCallAcceptSetting(ctx context.Context, req *user.SetCallAcceptSettingReq) error {
	req.UserID = u.loginUserID
	return api.SetCallAcceptSetting.Execute(ctx, req)
}

// setMsgReceiveSetting 调服务端设置会话消息接收权限（0=所有人 1=仅好友 2=所有人不可发送）
func (u *User) setMsgReceiveSetting(ctx context.Context, req *user.SetMsgReceiveSettingReq) error {
	req.UserID = u.loginUserID
	return api.SetMsgReceiveSetting.Execute(ctx, req)
}

// setGroupInviteSetting 调服务端设置群邀请权限（0=所有人 1=仅好友 2=所有人不可邀请）
func (u *User) setGroupInviteSetting(ctx context.Context, req *user.SetGroupInviteSettingReq) error {
	req.UserID = u.loginUserID
	return api.SetGroupInviteSetting.Execute(ctx, req)
}

// setUserMsgBurnDuration 调服务端设置全局阅后即焚时长（秒），0 表示关闭
func (u *User) setUserMsgBurnDuration(ctx context.Context, req *user.SetUserMsgBurnDurationReq) error {
	req.UserID = u.loginUserID
	return api.SetUserMsgBurnDuration.Execute(ctx, req)
}

// setDeleteAccountInterval 调服务端设置删除账号等待间隔（秒）；0 表示使用系统默认（18 个月）
func (u *User) setDeleteAccountInterval(ctx context.Context, req *user.SetDeleteAccountIntervalReq) error {
	req.UserID = u.loginUserID
	return api.SetDeleteAccountInterval.Execute(ctx, req)
}

// getUserPrivacySettings 调服务端获取当前登录用户隐私与接收相关设置（HTTP /user/get_user_privacy_settings）
func (u *User) getUserPrivacySettings(ctx context.Context) (*user.GetUserPrivacySettingsResp, error) {
	return api.GetUserPrivacySettings.Invoke(ctx, &user.GetUserPrivacySettingsReq{})
}

// getUserByPhone 调服务端按手机号精确查询用户；userInfo 为空表示未找到或无权限
func (u *User) getUserByPhone(ctx context.Context, phone string) (*sdkws.UserInfo, error) {
	req := &user.GetUserByPhoneReq{Phone: phone}
	resp, err := api.GetUserByPhone.Invoke(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.UserInfo == nil {
		return nil, nil
	}
	return resp.UserInfo, nil
}

// getUsersByNickname 按昵称精确查询普通用户列表
func (u *User) getUsersByNickname(ctx context.Context, nickname string) ([]*sdkws.UserInfo, error) {
	req := &user.GetUsersByNicknameReq{Nickname: nickname}
	resp, err := api.GetUsersByNickname.Invoke(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.UsersInfo == nil {
		return []*sdkws.UserInfo{}, nil
	}
	return resp.UsersInfo, nil
}
