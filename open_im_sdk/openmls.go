package open_im_sdk

import "github.com/openimsdk/openim-sdk-core/v3/open_im_sdk_callback"

// OpenMLSUploadKeyPackage calls POST /openmls/v1/key_packages/upload. req: JSON of openmls.UploadKeyPackageReq.
func OpenMLSUploadKeyPackage(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.OpenMLS().UploadKeyPackage, req)
}

// OpenMLSGetKeyPackages calls POST /openmls/v1/key_packages/fetch. req: JSON of openmls.GetKeyPackagesReq.
func OpenMLSGetKeyPackages(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.OpenMLS().GetKeyPackages, req)
}

// OpenMLSGetKeyPackageCount calls POST /openmls/v1/key_packages/count. req: JSON of openmls.GetKeyPackageCountReq.
func OpenMLSGetKeyPackageCount(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.OpenMLS().GetKeyPackageCount, req)
}

// OpenMLSRefreshKeyPackages calls POST /openmls/v1/key_packages/refresh. req: JSON of openmls.RefreshKeyPackagesReq.
func OpenMLSRefreshKeyPackages(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.OpenMLS().RefreshKeyPackages, req)
}

// OpenMLSSubmitCommit calls POST /openmls/v1/groups/commit. req: JSON of openmls.SubmitCommitReq.
func OpenMLSSubmitCommit(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.OpenMLS().SubmitCommit, req)
}

// OpenMLSGetCommits calls POST /openmls/v1/groups/commits. req: JSON of openmls.GetCommitsReq.
func OpenMLSGetCommits(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.OpenMLS().GetCommits, req)
}

// OpenMLSSendWelcome calls POST /openmls/v1/groups/welcome. req: JSON of openmls.SendWelcomeReq.
func OpenMLSSendWelcome(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.OpenMLS().SendWelcome, req)
}

// OpenMLSGetGroupState calls POST /openmls/v1/groups/state. req: JSON of openmls.GetGroupStateReq.
func OpenMLSGetGroupState(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.OpenMLS().GetGroupState, req)
}

// OpenMLSDeleteGroup calls POST /openmls/v1/groups/delete. req: JSON of openmls.DeleteGroupReq.
func OpenMLSDeleteGroup(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.OpenMLS().DeleteGroup, req)
}

// OpenMLSIssueCredential calls POST /crypto/v1/credential/issue. req: JSON of openmls.IssueCredentialReq.
func OpenMLSIssueCredential(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.OpenMLS().IssueCredential, req)
}

// OpenMLSVerifyCredential calls POST /crypto/v1/credential/verify. req: JSON of openmls.VerifyCredentialReq.
func OpenMLSVerifyCredential(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.OpenMLS().VerifyCredential, req)
}

// OpenMLSGetRootPublicKey calls POST /crypto/v1/root_public_key. req: JSON of openmls.GetRootPublicKeyReq.
func OpenMLSGetRootPublicKey(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.OpenMLS().GetRootPublicKey, req)
}
