package api

import "github.com/openimsdk/protocol/virgilsecurity"

var (
	VirgilIssueJWT            = newApi[virgilsecurity.IssueVirgilJWTReq, virgilsecurity.IssueVirgilJWTResp]("/virgil/v1/jwt")
	VirgilRegisterDevice      = newApi[virgilsecurity.RegisterDeviceReq, virgilsecurity.RegisterDeviceResp]("/virgil/v1/devices/register")
	VirgilGetDevices          = newApi[virgilsecurity.GetDevicesReq, virgilsecurity.GetDevicesResp]("/virgil/v1/devices")
	VirgilRevokeDevice        = newApi[virgilsecurity.RevokeDeviceReq, virgilsecurity.RevokeDeviceResp]("/virgil/v1/devices/revoke")
	VirgilEnsureConversation  = newApi[virgilsecurity.EnsureConversationReq, virgilsecurity.EnsureConversationResp]("/virgil/v1/conversations/ensure-1v1")
	VirgilSubscribeEvents     = newApi[virgilsecurity.SubscribeEventsReq, virgilsecurity.SubscribeEventsResp]("/virgil/v1/events/subscribe")
	VirgilCreateUploadURL     = newApi[virgilsecurity.CreateUploadURLReq, virgilsecurity.CreateUploadURLResp]("/virgil/v1/files/upload-url")
)
