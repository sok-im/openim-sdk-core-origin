// Copyright © 2024 OpenIM SDK. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.

package api

// WalletSetBackupInfoReq 对应服务端 POST /wallet/set_backup_info
type WalletSetBackupInfoReq struct {
	UID        string `json:"uid"`
	BackupTime int64  `json:"backupTime"` // 秒
	FileSize   int64  `json:"fileSize"`   // 字节
	Name       string `json:"name"`
}

// WalletSetBackupInfoResp 成功时服务端 data 为 null，可无字段
type WalletSetBackupInfoResp struct{}

// WalletGetBackupInfoReq 对应服务端 POST /wallet/get_backup_info
type WalletGetBackupInfoReq struct {
	UID string `json:"uid"`
}

// WalletGetBackupInfoResp 与服务端返回的 data 一致；无记录时字段为零值
type WalletGetBackupInfoResp struct {
	UID        string `json:"uid"`
	BackupTime int64  `json:"backupTime"`
	FileSize   int64  `json:"fileSize"`
	Name       string `json:"name"`
}

var (
	WalletSetBackupInfo = newApi[WalletSetBackupInfoReq, WalletSetBackupInfoResp]("/wallet/set_backup_info")
	WalletGetBackupInfo = newApi[WalletGetBackupInfoReq, WalletGetBackupInfoResp]("/wallet/get_backup_info")
)
