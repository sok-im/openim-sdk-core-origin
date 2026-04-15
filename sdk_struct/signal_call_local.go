// Copyright © 2023 OpenIM SDK. All rights reserved.

package sdk_struct

import "github.com/openimsdk/protocol/rtc"

// SearchLocalSignalCallRecordsParams 本地通话记录查询条件（数据仅本机写入，不依赖服务端历史接口）。
type SearchLocalSignalCallRecordsParams struct {
	Offset      int    `json:"offset"`
	Count       int    `json:"count"`
	SessionType int32  `json:"sessionType"`
	// DialStatus 见 constant.SignalCallDialStatus*：0=全部 1=未拨通 2=已拨通
	DialStatus int32 `json:"dialStatus"`
	StartTime  int64 `json:"startTime"`
	EndTime    int64 `json:"endTime"`
	Keyword    string `json:"keyword"`
	// UserName 仅按被叫模糊查询：匹配 callee_match_text（被叫 userID、信令 participant 昵称、本地好友昵称/备注）
	UserName string `json:"userName"`
}

// SignalCallRecordWithDialStatus 本地查询结果：协议消息 + 拨通状态（dialStatus：1=未拨通 2=已拨通）。record.sID 可作为 GetLocalSignalCallRecordDetail 入参。
type SignalCallRecordWithDialStatus struct {
	Record     *rtc.SignalRecord `json:"record"`
	DialStatus int32             `json:"dialStatus"`
}
