package api

import "github.com/openimsdk/protocol/group"

// Types below match open-im-server internal/api/group.go JSON for
// /group/set_*_setting, /group/get_group_setting, /group/set_msg_burn_duration, /group/get_msg_burn_duration,
// /group/set_group_announcement, and /group/get_group_announcement routes.

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

// SetInviteLinkSettingReq matches HTTP POST /group/set_invite_link_setting.
type SetInviteLinkSettingReq struct {
	GroupID          string `json:"groupID"`
	EnableInviteLink int32  `json:"enableInviteLink"`
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

type SetBurnSettingReq struct {
	GroupID   string `json:"groupID"`
	AllowBurn int32  `json:"allowBurn"`
}

type GetBurnSettingReq struct {
	GroupID string `json:"groupID"`
}

type GetBurnSettingResp struct {
	GroupID         string `json:"groupID"`
	AllowMemberBurn int32  `json:"allowMemberBurn"`
}

type GetGroupSettingReq struct {
	GroupID string `json:"groupID"`
}

type GetGroupSettingResp struct {
	GroupID            string                      `json:"groupID"`
	AllowSendMsg       int32                       `json:"allowSendMsg"`
	AllowAddMember     int32                       `json:"allowAddMember"`
	AllowPinMsg        int32                       `json:"allowPinMsg"`
	AllowEditGroupInfo int32                       `json:"allowEditGroupInfo"`
	AllowMemberBurn    int32                       `json:"allowMemberBurn"`
	EnableInviteLink   int32                       `json:"enableInviteLink"`
	InviteLink         []*group.GroupInviteLinkInfo `json:"inviteLink"`
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

// SetGroupAnnouncementReq matches HTTP POST /group/set_group_announcement.
type SetGroupAnnouncementReq struct {
	GroupID      string `json:"groupID"`
	Notification string `json:"notification"`
}

type GetGroupAnnouncementReq struct {
	GroupID string `json:"groupID"`
}

type GetGroupAnnouncementResp struct {
	GroupID                string `json:"groupID"`
	Notification           string `json:"notification"`
	NotificationUpdateTime int64  `json:"notificationUpdateTime"`
	NotificationUserID     string `json:"notificationUserID"`
}

var (
	SetSendMessageSetting = newApi[SetSendMessageSettingReq, group.SetGroupInfoExResp]("/group/set_send_message_setting")
	SetInviteSetting      = newApi[SetInviteSettingReq, group.SetGroupInfoExResp]("/group/set_invite_setting")
	SetInviteLinkSetting  = newApi[SetInviteLinkSettingReq, group.SetGroupInfoExResp]("/group/set_invite_link_setting")
	SetPinSetting         = newApi[SetPinSettingReq, group.SetGroupInfoExResp]("/group/set_pin_setting")
	SetEditSetting        = newApi[SetEditSettingReq, group.SetGroupInfoExResp]("/group/set_edit_setting")
	SetBurnSetting        = newApi[SetBurnSettingReq, group.SetGroupInfoExResp]("/group/set_burn_setting")
	GetGroupSetting       = newApi[GetGroupSettingReq, GetGroupSettingResp]("/group/get_group_setting")
	SetMsgBurnDuration    = newApi[SetMsgBurnDurationReq, group.SetGroupInfoExResp]("/group/set_msg_burn_duration")
	GetMsgBurnDuration    = newApi[GetMsgBurnDurationReq, GetMsgBurnDurationResp]("/group/get_msg_burn_duration")
	SetGroupAnnouncement  = newApi[SetGroupAnnouncementReq, group.SetGroupInfoExResp]("/group/set_group_announcement")
	GetGroupAnnouncement  = newApi[GetGroupAnnouncementReq, GetGroupAnnouncementResp]("/group/get_group_announcement")
)
