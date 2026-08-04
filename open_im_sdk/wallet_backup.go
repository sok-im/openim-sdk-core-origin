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

// WalletSetBackupInfo 上报钱包备份元数据（POST /wallet/set_backup_info，需登录 token）
func (u *LoginMgr) WalletSetBackupInfo(ctx context.Context, req *api.WalletSetBackupInfoReq) (*api.WalletSetBackupInfoResp, error) {
	return api.WalletSetBackupInfo.Invoke(ctx, req)
}

// WalletGetBackupInfo 查询钱包备份元数据（POST /wallet/get_backup_info，需登录 token）
func (u *LoginMgr) WalletGetBackupInfo(ctx context.Context, req *api.WalletGetBackupInfoReq) (*api.WalletGetBackupInfoResp, error) {
	return api.WalletGetBackupInfo.Invoke(ctx, req)
}

// WalletSetBackupInfo 参数 req 为 JSON：{"uid":"...","backupTime":1720000000,"fileSize":1048576,"name":"wallet-backup.zip"}
func WalletSetBackupInfo(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.WalletSetBackupInfo, req)
}

// WalletGetBackupInfo 参数 req 为 JSON：{"uid":"..."}
func WalletGetBackupInfo(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.WalletGetBackupInfo, req)
}
