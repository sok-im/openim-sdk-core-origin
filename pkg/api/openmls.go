package api

import "github.com/openimsdk/protocol/openmls"

var (
	OpenMLSUploadKeyPackage    = newApi[openmls.UploadKeyPackageReq, openmls.UploadKeyPackageResp]("/openmls/v1/key_packages/upload")
	OpenMLSGetKeyPackages      = newApi[openmls.GetKeyPackagesReq, openmls.GetKeyPackagesResp]("/openmls/v1/key_packages/fetch")
	OpenMLSGetKeyPackageCount  = newApi[openmls.GetKeyPackageCountReq, openmls.GetKeyPackageCountResp]("/openmls/v1/key_packages/count")
	OpenMLSRefreshKeyPackages  = newApi[openmls.RefreshKeyPackagesReq, openmls.RefreshKeyPackagesResp]("/openmls/v1/key_packages/refresh")
	OpenMLSSubmitCommit        = newApi[openmls.SubmitCommitReq, openmls.SubmitCommitResp]("/openmls/v1/groups/commit")
	OpenMLSGetCommits          = newApi[openmls.GetCommitsReq, openmls.GetCommitsResp]("/openmls/v1/groups/commits")
	OpenMLSSendWelcome         = newApi[openmls.SendWelcomeReq, openmls.SendWelcomeResp]("/openmls/v1/groups/welcome")
	OpenMLSGetGroupState       = newApi[openmls.GetGroupStateReq, openmls.GetGroupStateResp]("/openmls/v1/groups/state")
	OpenMLSDeleteGroup         = newApi[openmls.DeleteGroupReq, openmls.DeleteGroupResp]("/openmls/v1/groups/delete")
	OpenMLSIssueCredential     = newApi[openmls.IssueCredentialReq, openmls.IssueCredentialResp]("/crypto/v1/credential/issue")
	OpenMLSVerifyCredential    = newApi[openmls.VerifyCredentialReq, openmls.VerifyCredentialResp]("/crypto/v1/credential/verify")
	OpenMLSGetRootPublicKey    = newApi[openmls.GetRootPublicKeyReq, openmls.GetRootPublicKeyResp]("/crypto/v1/root_public_key")
)
