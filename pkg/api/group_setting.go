package api

import "github.com/openimsdk/protocol/group"

// Types below match open-im-server internal/api/group.go JSON for
// /group/set_*_setting and /group/get_*_setting routes.

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

var (
	SetSendMessageSetting = newApi[SetSendMessageSettingReq, group.SetGroupInfoExResp]("/group/set_send_message_setting")
	GetSendMessageSetting = newApi[GetSendMessageSettingReq, GetSendMessageSettingResp]("/group/get_send_message_setting")
	SetInviteSetting      = newApi[SetInviteSettingReq, group.SetGroupInfoExResp]("/group/set_invite_setting")
	GetInviteSetting      = newApi[GetInviteSettingReq, GetInviteSettingResp]("/group/get_invite_setting")
	SetPinSetting         = newApi[SetPinSettingReq, group.SetGroupInfoExResp]("/group/set_pin_setting")
	GetPinSetting         = newApi[GetPinSettingReq, GetPinSettingResp]("/group/get_pin_setting")
	SetEditSetting        = newApi[SetEditSettingReq, group.SetGroupInfoExResp]("/group/set_edit_setting")
	GetEditSetting        = newApi[GetEditSettingReq, GetEditSettingResp]("/group/get_edit_setting")
)
