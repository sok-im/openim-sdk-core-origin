package virgilsecurity

import (
	"context"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/api"
	pbvirgil "github.com/openimsdk/protocol/virgilsecurity"
)

// VirgilSecurity exposes HTTP API calls to the server's /virgil/v1 routes (requires login).
type VirgilSecurity struct{}

func NewVirgilSecurity() *VirgilSecurity {
	return &VirgilSecurity{}
}

func (v *VirgilSecurity) IssueVirgilJWT(ctx context.Context, req *pbvirgil.IssueVirgilJWTReq) (*pbvirgil.IssueVirgilJWTResp, error) {
	return api.VirgilIssueJWT.Invoke(ctx, req)
}

func (v *VirgilSecurity) RegisterDevice(ctx context.Context, req *pbvirgil.RegisterDeviceReq) (*pbvirgil.RegisterDeviceResp, error) {
	return api.VirgilRegisterDevice.Invoke(ctx, req)
}

func (v *VirgilSecurity) GetDevices(ctx context.Context, req *pbvirgil.GetDevicesReq) (*pbvirgil.GetDevicesResp, error) {
	return api.VirgilGetDevices.Invoke(ctx, req)
}

func (v *VirgilSecurity) RevokeDevice(ctx context.Context, req *pbvirgil.RevokeDeviceReq) (*pbvirgil.RevokeDeviceResp, error) {
	return api.VirgilRevokeDevice.Invoke(ctx, req)
}

func (v *VirgilSecurity) EnsureConversation(ctx context.Context, req *pbvirgil.EnsureConversationReq) (*pbvirgil.EnsureConversationResp, error) {
	return api.VirgilEnsureConversation.Invoke(ctx, req)
}

func (v *VirgilSecurity) SubscribeEvents(ctx context.Context, req *pbvirgil.SubscribeEventsReq) (*pbvirgil.SubscribeEventsResp, error) {
	return api.VirgilSubscribeEvents.Invoke(ctx, req)
}

func (v *VirgilSecurity) CreateUploadURL(ctx context.Context, req *pbvirgil.CreateUploadURLReq) (*pbvirgil.CreateUploadURLResp, error) {
	return api.VirgilCreateUploadURL.Invoke(ctx, req)
}
