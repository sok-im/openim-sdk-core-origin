package open_im_sdk

import "github.com/openimsdk/openim-sdk-core/v3/open_im_sdk_callback"

// CryptoRegisterDevice registers the current device so it can obtain Virgil JWTs.
// req: JSON of crypto.RegisterDeviceReq (deviceID, platform, deviceModel, appVersion)
func CryptoRegisterDevice(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Crypto().RegisterDevice, req)
}

// CryptoGetDevices returns all devices registered for the current user.
// req: JSON of crypto.GetDevicesReq (userID is filled automatically)
func CryptoGetDevices(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Crypto().GetDevices, req)
}

// CryptoRevokeDevice revokes a device so it can no longer obtain Virgil JWTs.
// req: JSON of crypto.RevokeDeviceReq (deviceID)
func CryptoRevokeDevice(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Crypto().RevokeDevice, req)
}

// CryptoGetVirgilJWT obtains a short-lived Virgil JWT for the given device.
// Pass the returned virgilJWT to Virgil E3Kit on the client to authenticate with Virgil Cloud.
// req: JSON of crypto.GetVirgilJWTReq (deviceID)
func CryptoGetVirgilJWT(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Crypto().GetVirgilJWT, req)
}

// CryptoGetGroupKeyVersion returns the current group key version.
// Clients use this to detect whether the group session key needs to be refreshed.
// req: JSON of crypto.GetGroupKeyVersionReq (groupID)
func CryptoGetGroupKeyVersion(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Crypto().GetGroupKeyVersion, req)
}

// CryptoGetGroupKeyEvents fetches group key rotation events since a given version.
// Use this to replay missed key rotations after reconnection.
// req: JSON of crypto.GetGroupKeyEventsReq (groupID, sinceVersion)
func CryptoGetGroupKeyEvents(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Crypto().GetGroupKeyEvents, req)
}

// CryptoSecurityPrecheck verifies the device is active and belongs to the current
// user before performing sensitive crypto operations.
// req: JSON of crypto.SecurityPrecheckReq (deviceID, action)
func CryptoSecurityPrecheck(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Crypto().SecurityPrecheck, req)
}

// CryptoIntegrityReport submits a device integrity attestation report to the server
// (e.g. SafetyNet / Play Integrity / Apple App Attest data).
// req: JSON of crypto.IntegrityReportReq (deviceID, reportData, timestamp)
func CryptoIntegrityReport(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.Crypto().IntegrityReport, req)
}
