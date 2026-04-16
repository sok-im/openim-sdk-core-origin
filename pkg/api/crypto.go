package api

import "github.com/openimsdk/protocol/crypto"

var (
	CryptoRegisterDevice     = newApi[crypto.RegisterDeviceReq, crypto.RegisterDeviceResp]("/crypto/register_device")
	CryptoGetDevices         = newApi[crypto.GetDevicesReq, crypto.GetDevicesResp]("/crypto/get_devices")
	CryptoRevokeDevice       = newApi[crypto.RevokeDeviceReq, crypto.RevokeDeviceResp]("/crypto/revoke_device")
	CryptoGetVirgilJWT       = newApi[crypto.GetVirgilJWTReq, crypto.GetVirgilJWTResp]("/crypto/get_virgil_jwt")
	CryptoGetGroupKeyVersion = newApi[crypto.GetGroupKeyVersionReq, crypto.GetGroupKeyVersionResp]("/crypto/get_group_key_version")
	CryptoGetGroupKeyEvents  = newApi[crypto.GetGroupKeyEventsReq, crypto.GetGroupKeyEventsResp]("/crypto/get_group_key_events")
	CryptoSecurityPrecheck   = newApi[crypto.SecurityPrecheckReq, crypto.SecurityPrecheckResp]("/crypto/security_precheck")
	CryptoIntegrityReport    = newApi[crypto.IntegrityReportReq, crypto.IntegrityReportResp]("/crypto/integrity_report")
)
