// Copyright © 2023 OpenIM SDK. All rights reserved.

package sdk_struct

import "github.com/openimsdk/protocol/rtc"

// SearchLocalSignalCallRecordsParams 本地通话记录查询条件（数据仅本机写入，不依赖服务端历史接口）。
type SearchLocalSignalCallRecordsParams struct {
	Offset      int   `json:"offset"`
	Count       int   `json:"count"`
	SessionType int32 `json:"sessionType"`
	// DialStatus 见 constant.SignalCallDialStatus*：0=全部 1=未拨通 2=已拨通
	DialStatus int32 `json:"dialStatus"`
	// Direction 见 constant.SignalCallDirection*：0=全部 1=主叫 2=被叫-已接 3=被叫-未接
	Direction int32 `json:"direction"`
	StartTime int64 `json:"startTime"`
	EndTime   int64 `json:"endTime"`
	Keyword   string `json:"keyword"`
	// UserName 仅按被叫模糊查询：匹配 callee_match_text（被叫 userID、信令 participant 昵称、本地好友昵称/备注）
	UserName string `json:"userName"`
}

// SignalCallRecordWithDialStatus 本地查询结果：协议消息 + 拨通状态 + 通话方向 + 时长信息。
// record.sID 可作为 GetLocalSignalCallRecordDetail 入参。
type SignalCallRecordWithDialStatus struct {
	Record     *rtc.SignalRecord `json:"record"`
	// DialStatus 见 constant.SignalCallDialStatus*：1=未拨通 2=已拨通
	DialStatus int32 `json:"dialStatus"`
	// Direction 见 constant.SignalCallDirection*：1=主叫(outgoing) 2=被叫-已接(incoming) 3=被叫-未接(missed)
	Direction int32 `json:"direction"`
	// ConnectTime 接通时间（毫秒时间戳），未接通时为 0
	ConnectTime int64 `json:"connectTime"`
	// DialDuration 拨打/振铃时长（毫秒）：从本端发起/收到邀请到接通/拒接/取消所经历的时间
	DialDuration int64 `json:"dialDuration"`
	// CallDuration 通话时长（毫秒），未接通时为 0
	CallDuration int64 `json:"callDuration"`
	// InviteeUserNickname 被叫昵称（单聊时首位被叫，用于 UI 展示；群组通话时可能为空）
	InviteeUserNickname string `json:"inviteeUserNickname"`
}
