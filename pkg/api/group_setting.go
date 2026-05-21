package api

import "github.com/openimsdk/protocol/group"

// Types below match open-im-server internal/api/group.go JSON for
// /group/set_*_setting, /group/get_group_setting, /group/set_msg_burn_duration, and /group/get_msg_burn_duration routes.

type SetSendMessageSettingReq struct {
	GroupID      string `json:"groupID"`
	AllowSendMsg int32  `json:"allowSendMsg"`
}

type GetSendMessageSettingReq struct {
	GroupID string `json:"groupID"`
}

type GetSendMessageSettingResp struct {
	GroupID      string `json:"groupID"`
	AllowSendMsg int32  `json:"allowSendMsg"`
}

type SetInviteSettingReq struct {
	GroupID        string `json:"groupID"`
	AllowAddMember int32  `json:"allowAddMember"`
}

type GetInviteSettingReq struct {
	GroupID string `json:"groupID"`
}

type GetInviteSettingResp struct {
	GroupID        string `json:"groupID"`
	AllowAddMember int32  `json:"allowAddMember"`
}

type SetPinSettingReq struct {
	GroupID     string `json:"groupID"`
	AllowPinMsg int32  `json:"allowPinMsg"`
}

type GetPinSettingReq struct {
	GroupID string `json:"groupID"`
}

type GetPinSettingResp struct {
	GroupID     string `json:"groupID"`
	AllowPinMsg int32  `json:"allowPinMsg"`
}

type SetEditSettingReq struct {
	GroupID            string `json:"groupID"`
	AllowEditGroupInfo int32  `json:"allowEditGroupInfo"`
}

type GetEditSettingReq struct {
	GroupID string `json:"groupID"`
}

type GetEditSettingResp struct {
	GroupID            string `json:"groupID"`
	AllowEditGroupInfo int32  `json:"allowEditGroupInfo"`
}

type GetGroupSettingReq struct {
	GroupID string `json:"groupID"`
}

type GetGroupSettingResp struct {
	GroupID            string `json:"groupID"`
	AllowSendMsg       int32  `json:"allowSendMsg"`
	AllowAddMember     int32  `json:"allowAddMember"`
	AllowPinMsg        int32  `json:"allowPinMsg"`
	AllowEditGroupInfo int32  `json:"allowEditGroupInfo"`
}

// SetMsgBurnDurationReq matches HTTP POST /group/set_msg_burn_duration (burnDuration in seconds; 0 = off).
type SetMsgBurnDurationReq struct {
	GroupID      string `json:"groupID"`
	BurnDuration int32  `json:"burnDuration"`
}

type GetMsgBurnDurationReq struct {
	GroupID string `json:"groupID"`
}

type GetMsgBurnDurationResp struct {
	GroupID      string `json:"groupID"`
	BurnDuration int32  `json:"burnDuration"`
}

var (
	SetSendMessageSetting = newApi[SetSendMessageSettingReq, group.SetGroupInfoExResp]("/group/set_send_message_setting")
	SetInviteSetting      = newApi[SetInviteSettingReq, group.SetGroupInfoExResp]("/group/set_invite_setting")
	SetPinSetting         = newApi[SetPinSettingReq, group.SetGroupInfoExResp]("/group/set_pin_setting")
	SetEditSetting        = newApi[SetEditSettingReq, group.SetGroupInfoExResp]("/group/set_edit_setting")
	GetGroupSetting       = newApi[GetGroupSettingReq, GetGroupSettingResp]("/group/get_group_setting")
	SetMsgBurnDuration    = newApi[SetMsgBurnDurationReq, group.SetGroupInfoExResp]("/group/set_msg_burn_duration")
	GetMsgBurnDuration    = newApi[GetMsgBurnDurationReq, GetMsgBurnDurationResp]("/group/get_msg_burn_duration")
)
