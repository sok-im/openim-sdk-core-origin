package open_im_sdk

import "github.com/openimsdk/openim-sdk-core/v3/open_im_sdk_callback"

// VirgilIssueJWT calls POST /virgil/v1/jwt. req: JSON of virgilsecurity.IssueVirgilJWTReq.
func VirgilIssueJWT(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.VirgilSecurity().IssueVirgilJWT, req)
}

// VirgilRegisterDevice calls POST /virgil/v1/devices/register. req: JSON of virgilsecurity.RegisterDeviceReq.
func VirgilRegisterDevice(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.VirgilSecurity().RegisterDevice, req)
}

// VirgilGetDevices calls POST /virgil/v1/devices. req: JSON of virgilsecurity.GetDevicesReq.
func VirgilGetDevices(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.VirgilSecurity().GetDevices, req)
}

// VirgilRevokeDevice calls POST /virgil/v1/devices/revoke. req: JSON of virgilsecurity.RevokeDeviceReq.
func VirgilRevokeDevice(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.VirgilSecurity().RevokeDevice, req)
}

// VirgilEnsureConversation calls POST /virgil/v1/conversations/ensure-1v1. req: JSON of virgilsecurity.EnsureConversationReq.
func VirgilEnsureConversation(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.VirgilSecurity().EnsureConversation, req)
}

// VirgilSubscribeEvents calls POST /virgil/v1/events/subscribe. req: JSON of virgilsecurity.SubscribeEventsReq.
func VirgilSubscribeEvents(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.VirgilSecurity().SubscribeEvents, req)
}

// VirgilCreateUploadURL calls POST /virgil/v1/files/upload-url. req: JSON of virgilsecurity.CreateUploadURLReq.
func VirgilCreateUploadURL(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.VirgilSecurity().CreateUploadURL, req)
}
