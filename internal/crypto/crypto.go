package crypto

import (
	"context"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/api"
	"github.com/openimsdk/protocol/crypto"
)

// Crypto provides client-side access to the server's E2EE crypto service.
// The server is responsible only for:
//   - Issuing Virgil JWTs (so the client can authenticate with Virgil Cloud)
//   - Managing device registration / revocation records
//   - Tracking group key rotation versions
//
// All actual encryption/decryption is performed on the client via Virgil E3Kit.
type Crypto struct {
	loginUserID string
}

func NewCrypto(loginUserID string) *Crypto {
	return &Crypto{loginUserID: loginUserID}
}

// RegisterDevice registers the current device with the crypto service so it
// can subsequently obtain Virgil JWTs.
func (c *Crypto) RegisterDevice(ctx context.Context, req *crypto.RegisterDeviceReq) (*crypto.RegisterDeviceResp, error) {
	req.UserID = c.loginUserID
	return api.CryptoRegisterDevice.Invoke(ctx, req)
}

// GetDevices returns all devices registered for the current user.
func (c *Crypto) GetDevices(ctx context.Context, req *crypto.GetDevicesReq) (*crypto.GetDevicesResp, error) {
	req.UserID = c.loginUserID
	return api.CryptoGetDevices.Invoke(ctx, req)
}

// RevokeDevice revokes a registered device so it can no longer obtain JWTs.
func (c *Crypto) RevokeDevice(ctx context.Context, req *crypto.RevokeDeviceReq) (*crypto.RevokeDeviceResp, error) {
	req.UserID = c.loginUserID
	return api.CryptoRevokeDevice.Invoke(ctx, req)
}

// GetVirgilJWT obtains a short-lived Virgil JWT for the given device.
// The client passes this token to Virgil E3Kit to authenticate with Virgil Cloud.
func (c *Crypto) GetVirgilJWT(ctx context.Context, req *crypto.GetVirgilJWTReq) (*crypto.GetVirgilJWTResp, error) {
	req.UserID = c.loginUserID
	return api.CryptoGetVirgilJWT.Invoke(ctx, req)
}

// GetGroupKeyVersion returns the current group key version.
// Clients poll this to detect whether the group session key needs to be refreshed.
func (c *Crypto) GetGroupKeyVersion(ctx context.Context, req *crypto.GetGroupKeyVersionReq) (*crypto.GetGroupKeyVersionResp, error) {
	return api.CryptoGetGroupKeyVersion.Invoke(ctx, req)
}

// GetGroupKeyEvents fetches group key rotation events since a given version.
// Clients use this to replay missed key rotations after reconnection.
func (c *Crypto) GetGroupKeyEvents(ctx context.Context, req *crypto.GetGroupKeyEventsReq) (*crypto.GetGroupKeyEventsResp, error) {
	return api.CryptoGetGroupKeyEvents.Invoke(ctx, req)
}

// SecurityPrecheck performs a server-side pre-flight check to verify the device
// is active and belongs to the current user before sensitive crypto operations.
func (c *Crypto) SecurityPrecheck(ctx context.Context, req *crypto.SecurityPrecheckReq) (*crypto.SecurityPrecheckResp, error) {
	req.UserID = c.loginUserID
	return api.CryptoSecurityPrecheck.Invoke(ctx, req)
}

// IntegrityReport submits a device integrity attestation report to the server.
// The server may revoke the device if the report indicates a compromised environment.
func (c *Crypto) IntegrityReport(ctx context.Context, req *crypto.IntegrityReportReq) (*crypto.IntegrityReportResp, error) {
	req.UserID = c.loginUserID
	return api.CryptoIntegrityReport.Invoke(ctx, req)
}
