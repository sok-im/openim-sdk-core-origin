// Copyright © 2023 OpenIM SDK. All rights reserved.

package sdk_struct

// SearchLocalSignalCallRecordsParams 本地通话记录查询条件（数据仅本机写入，不依赖服务端历史接口）。
type SearchLocalSignalCallRecordsParams struct {
	Offset      int   `json:"offset"`
	Count       int   `json:"count"`
	SessionType int32 `json:"sessionType"`
	// Status 见 constant.SignalCallStatus*：0=全部 1=已接听 2=未接通
	Status int32 `json:"status"`
	// Direction 见 constant.SignalCallDirection*：0=全部 1=主叫 2=被叫-已接 3=被叫-未接
	Direction int32 `json:"direction"`
	StartTime int64 `json:"startTime"`
	EndTime   int64 `json:"endTime"`
	Keyword   string `json:"keyword"`
	// UserName 用户名模糊查询：匹配 inviter_user_nickname、invitee_user_nickname、callee_match_text
	UserName string `json:"userName"`
	// InviteeNickname 被叫昵称模糊查询：匹配 invitee_user_nickname
	InviteeNickname string `json:"inviteeNickname"`
}

// GetLocalCallRecordsParams 查询本地通话记录（可按用户筛选）。
type GetLocalCallRecordsParams struct {
	// UserID 对方用户 ID；为空则不过滤用户
	UserID string `json:"userID"`
	// Status 见 constant.SignalCallStatus*：0=全部（已接听+未接通） 1=已接听 2=未接通
	Status    int32 `json:"status"`
	Offset    int   `json:"offset"`
	Count     int   `json:"count"`
	StartTime int64 `json:"startTime"`
	EndTime   int64 `json:"endTime"`
}

// GetLocalCallRecordsResp 本地通话记录查询结果。
type GetLocalCallRecordsResp struct {
	Total   int64                             `json:"total"`
	Records []*SignalCallRecordWithDialStatus `json:"records"`
}

// GetLocalMissedCallRecordsParams 查询本地未接来电（被叫未接，direction=3）。
type GetLocalMissedCallRecordsParams struct {
	// UserID 主叫用户 ID（inviter_user_id）；为空则不过滤
	UserID      string `json:"userID"`
	Offset      int    `json:"offset"`
	Count       int    `json:"count"`
	SessionType int32  `json:"sessionType"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
	Keyword     string `json:"keyword"`
}

// GetLocalAnsweredCallRecordsParams 查询本地已接通通话（status=已接听）。
type GetLocalAnsweredCallRecordsParams struct {
	// UserID 对方用户 ID；为空则不过滤（匹配主叫或被叫列表）
	UserID      string `json:"userID"`
	Offset      int    `json:"offset"`
	Count       int    `json:"count"`
	SessionType int32  `json:"sessionType"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
	Keyword     string `json:"keyword"`
}

// GetLocalCallRecordsByUserNameParams 按用户名模糊查询本地通话记录。
type GetLocalCallRecordsByUserNameParams struct {
	// UserName 用户名关键字（模糊匹配主叫/被叫昵称及 callee_match_text）
	UserName    string `json:"userName"`
	Offset      int    `json:"offset"`
	Count       int    `json:"count"`
	SessionType int32  `json:"sessionType"`
	// Status 见 constant.SignalCallStatus*：0=全部 1=已接听 2=未接通
	Status int32 `json:"status"`
	// Direction 见 constant.SignalCallDirection*：0=全部 1=主叫 2=被叫-已接 3=被叫-未接
	Direction int32 `json:"direction"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
}

// GetLocalAllCallRecordsParams 查询本地全部通话记录（不筛选 status/direction）。
type GetLocalAllCallRecordsParams struct {
	// UserID 对方用户 ID；为空则不过滤（匹配主叫或被叫列表）
	UserID      string `json:"userID"`
	Offset      int    `json:"offset"`
	Count       int    `json:"count"`
	SessionType int32  `json:"sessionType"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
	Keyword     string `json:"keyword"`
}

// SignalCallRecordWithDialStatus 本地查询结果。
// sID 可作为 GetLocalSignalCallRecordDetail 入参。
type SignalCallRecordWithDialStatus struct {
	SID                 string `json:"sID"`
	RoomID              string `json:"roomID"`
	Status       int32 `json:"status"`
	CreateTime   int64 `json:"createTime"`
	// DialDuration 拨打时长（毫秒）：发起/收到邀请 → 接通；未接通时为发起 → 结束
	DialDuration int64 `json:"dialDuration"`
	// CallDuration 通话时长（毫秒）：接通 → 挂断；未接通时为 0
	CallDuration int64 `json:"callDuration"`
	MediaType           string `json:"mediaType"`
	SessionType         int32  `json:"sessionType"`
	InviterUserID       string `json:"inviterUserID"`
	InviterUserNickname string `json:"inviterUserNickname"`
	InviterUserFaceURL  string `json:"inviterUserFaceURL"`
	GroupID             string `json:"groupID"`
	GroupName           string `json:"groupName"`
	Direction           int32  `json:"direction"`
	// Role 见 constant.SignalCallRole*：1=主叫 2=被叫
	Role        int32 `json:"role"`
	ConnectTime int64 `json:"connectTime"`
	InviteeUserNickname string `json:"inviteeUserNickname"`
	InviteeUID          string `json:"inviteeUID"`
	InviteeUserFaceURL  string `json:"inviteeUserFaceURL"`
	// Action 信令动作：accept / reject / cancel / hungup / timeout
	Action string `json:"action"`
}
