// Copyright © 2024 OpenIM SDK. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.

package api

// PhoneGetSNInfoReq 对应服务端 POST /phone/get_sn_info
type PhoneGetSNInfoReq struct {
	Phone string `json:"phone"`
}

// PhoneGetSNInfoResp 与 open-im-server internal/api/phone_sn 返回的 data 一致
type PhoneGetSNInfoResp struct {
	IsSnd  bool  `json:"is_snd"`
	UserID int64 `json:"userID"`
}

// PhoneSetSNInfoReq 对应服务端 POST /phone/set_sn_info
type PhoneSetSNInfoReq struct {
	Phone  string `json:"phone"`
	UserID int64  `json:"userID"`
	IsSnd  bool   `json:"is_snd"`
}

// PhoneSetSNInfoResp 成功时服务端 data 为 null，可无字段
type PhoneSetSNInfoResp struct{}

var (
	PhoneGetSNInfo = newApi[PhoneGetSNInfoReq, PhoneGetSNInfoResp]("/phone/get_sn_info")
	PhoneSetSNInfo = newApi[PhoneSetSNInfoReq, PhoneSetSNInfoResp]("/phone/set_sn_info")
)
