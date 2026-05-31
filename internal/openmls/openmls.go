package openmls

import (
	"context"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/api"
	pbopenmls "github.com/openimsdk/protocol/openmls"
)

// OpenMLS exposes HTTP API calls to the server's /openmls/v1 and /crypto/v1 routes (requires login).
type OpenMLS struct{}

func NewOpenMLS() *OpenMLS {
	return &OpenMLS{}
}

func (o *OpenMLS) UploadKeyPackage(ctx context.Context, req *pbopenmls.UploadKeyPackageReq) (*pbopenmls.UploadKeyPackageResp, error) {
	return api.OpenMLSUploadKeyPackage.Invoke(ctx, req)
}

func (o *OpenMLS) GetKeyPackages(ctx context.Context, req *pbopenmls.GetKeyPackagesReq) (*pbopenmls.GetKeyPackagesResp, error) {
	return api.OpenMLSGetKeyPackages.Invoke(ctx, req)
}

func (o *OpenMLS) GetKeyPackageCount(ctx context.Context, req *pbopenmls.GetKeyPackageCountReq) (*pbopenmls.GetKeyPackageCountResp, error) {
	return api.OpenMLSGetKeyPackageCount.Invoke(ctx, req)
}

func (o *OpenMLS) RefreshKeyPackages(ctx context.Context, req *pbopenmls.RefreshKeyPackagesReq) (*pbopenmls.RefreshKeyPackagesResp, error) {
	return api.OpenMLSRefreshKeyPackages.Invoke(ctx, req)
}

func (o *OpenMLS) SubmitCommit(ctx context.Context, req *pbopenmls.SubmitCommitReq) (*pbopenmls.SubmitCommitResp, error) {
	return api.OpenMLSSubmitCommit.Invoke(ctx, req)
}

func (o *OpenMLS) GetCommits(ctx context.Context, req *pbopenmls.GetCommitsReq) (*pbopenmls.GetCommitsResp, error) {
	return api.OpenMLSGetCommits.Invoke(ctx, req)
}

func (o *OpenMLS) SendWelcome(ctx context.Context, req *pbopenmls.SendWelcomeReq) (*pbopenmls.SendWelcomeResp, error) {
	return api.OpenMLSSendWelcome.Invoke(ctx, req)
}

func (o *OpenMLS) GetGroupState(ctx context.Context, req *pbopenmls.GetGroupStateReq) (*pbopenmls.GetGroupStateResp, error) {
	return api.OpenMLSGetGroupState.Invoke(ctx, req)
}

func (o *OpenMLS) DeleteGroup(ctx context.Context, req *pbopenmls.DeleteGroupReq) (*pbopenmls.DeleteGroupResp, error) {
	return api.OpenMLSDeleteGroup.Invoke(ctx, req)
}

func (o *OpenMLS) IssueCredential(ctx context.Context, req *pbopenmls.IssueCredentialReq) (*pbopenmls.IssueCredentialResp, error) {
	return api.OpenMLSIssueCredential.Invoke(ctx, req)
}

func (o *OpenMLS) VerifyCredential(ctx context.Context, req *pbopenmls.VerifyCredentialReq) (*pbopenmls.VerifyCredentialResp, error) {
	return api.OpenMLSVerifyCredential.Invoke(ctx, req)
}

func (o *OpenMLS) GetRootPublicKey(ctx context.Context, req *pbopenmls.GetRootPublicKeyReq) (*pbopenmls.GetRootPublicKeyResp, error) {
	return api.OpenMLSGetRootPublicKey.Invoke(ctx, req)
}
