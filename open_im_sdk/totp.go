// Copyright © 2023 OpenIM SDK. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package open_im_sdk

import (
	"context"

	"github.com/openimsdk/openim-sdk-core/v3/open_im_sdk_callback"
	"github.com/openimsdk/openim-sdk-core/v3/pkg/api"
	pbtotp "github.com/openimsdk/protocol/totp"
)

// TotpGetSecret generates a temporary TOTP secret and otpauth URI (HTTP POST /totp/secret).
// Requires login; userID is injected by the server from the token.
func (u *LoginMgr) TotpGetSecret(ctx context.Context, req *pbtotp.GetSecretReq) (*pbtotp.GetSecretResp, error) {
	return api.TotpGetSecret.Invoke(ctx, req)
}

// TotpBind verifies the first dynamic code and activates TOTP (HTTP POST /totp/bind).
func (u *LoginMgr) TotpBind(ctx context.Context, req *pbtotp.BindTotpReq) (*pbtotp.BindTotpResp, error) {
	return api.TotpBind.Invoke(ctx, req)
}

// TotpVerify is the second step of MFA login (HTTP POST /totp/verify).
// Does not require a login token; authenticates via mfaToken.
func (u *LoginMgr) TotpVerify(ctx context.Context, req *pbtotp.VerifyTotpReq) (*pbtotp.VerifyTotpResp, error) {
	return api.TotpVerify.Invoke(ctx, req)
}

// TotpGetStatus returns whether the caller has TOTP bound (HTTP POST /totp/status).
func (u *LoginMgr) TotpGetStatus(ctx context.Context, req *pbtotp.GetStatusReq) (*pbtotp.GetStatusResp, error) {
	return api.TotpGetStatus.Invoke(ctx, req)
}

// TotpUnbind removes TOTP binding after verifying the current code (HTTP POST /totp/unbind).
func (u *LoginMgr) TotpUnbind(ctx context.Context, req *pbtotp.UnbindTotpReq) (*pbtotp.UnbindTotpResp, error) {
	return api.TotpUnbind.Invoke(ctx, req)
}

// TotpGetSecret req: JSON of totp.GetSecretReq (issuer, accountName; userID is filled by server).
func TotpGetSecret(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.TotpGetSecret, req)
}

// TotpBind req: JSON of totp.BindTotpReq (totpCode).
func TotpBind(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.TotpBind, req)
}

// TotpVerify can be called before login. req: JSON of totp.VerifyTotpReq (mfaToken, totpCode, platform).
func TotpVerify(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.TotpVerify, req)
}

// TotpGetStatus req: JSON of totp.GetStatusReq (empty object is sufficient).
func TotpGetStatus(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.TotpGetStatus, req)
}

// TotpUnbind req: JSON of totp.UnbindTotpReq (totpCode).
func TotpUnbind(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.TotpUnbind, req)
}
