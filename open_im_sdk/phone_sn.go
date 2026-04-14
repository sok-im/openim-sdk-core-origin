// Copyright © 2024 OpenIM SDK. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.

package open_im_sdk

import (
	"context"

	"github.com/openimsdk/openim-sdk-core/v3/open_im_sdk_callback"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/api"
)

// PhoneGetSNInfo 查询手机号 is_snd（POST /phone/get_sn_info，需登录 token）
func (u *LoginMgr) PhoneGetSNInfo(ctx context.Context, req *api.PhoneGetSNInfoReq) (*api.PhoneGetSNInfoResp, error) {
	return api.PhoneGetSNInfo.Invoke(ctx, req)
}

// PhoneSetSNInfo 设置手机号 is_snd 与 userID（POST /phone/set_sn_info，需登录 token）
func (u *LoginMgr) PhoneSetSNInfo(ctx context.Context, req *api.PhoneSetSNInfoReq) (*api.PhoneSetSNInfoResp, error) {
	return api.PhoneSetSNInfo.Invoke(ctx, req)
}

// PhoneGetSNInfo 参数 req 为 JSON：{"phone":"..."}
func PhoneGetSNInfo(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.PhoneGetSNInfo, req)
}

// PhoneSetSNInfo 参数 req 为 JSON：{"phone":"...","userID":123,"is_snd":true}
func PhoneSetSNInfo(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.PhoneSetSNInfo, req)
}
