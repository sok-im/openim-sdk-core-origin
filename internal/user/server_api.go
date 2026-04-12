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
