package api

import "github.com/openimsdk/protocol/group"

var (
	CreateGroupInviteLink  = newApi[group.CreateGroupInviteLinkReq, group.CreateGroupInviteLinkResp]("/group/create_invite_link")
	GetGroupInviteLink     = newApi[group.GetGroupInviteLinkReq, group.GetGroupInviteLinkResp]("/group/get_invite_link")
	JoinGroupByInviteLink  = newApi[group.JoinGroupByInviteLinkReq, group.JoinGroupByInviteLinkResp]("/group/join_by_invite_link")
	RevokeGroupInviteLink  = newApi[group.RevokeGroupInviteLinkReq, group.RevokeGroupInviteLinkResp]("/group/revoke_invite_link")
	ListGroupInviteLinks   = newApi[group.ListGroupInviteLinksReq, group.ListGroupInviteLinksResp]("/group/list_invite_links")
)
