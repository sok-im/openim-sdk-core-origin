package api

import "github.com/openimsdk/protocol/group"

var (
	CreateGroupInviteLink        = newApi[group.CreateGroupInviteLinkReq, group.CreateGroupInviteLinkResp]("/group/create_invite_link")
	GetGroupInviteLink           = newApi[group.GetGroupInviteLinkReq, group.GetGroupInviteLinkResp]("/group/get_invite_link")
	JoinGroupByInviteLink        = newApi[group.JoinGroupByInviteLinkReq, group.JoinGroupByInviteLinkResp]("/group/join_by_invite_link")
	RevokeGroupInviteLink        = newApi[group.RevokeGroupInviteLinkReq, group.RevokeGroupInviteLinkResp]("/group/revoke_invite_link")
	GetGroupInviteLinkByGroupID  = newApi[group.GetGroupInviteLinkByGroupIDReq, group.GetGroupInviteLinkByGroupIDResp]("/group/get_group_invite_link_by_group")
)
