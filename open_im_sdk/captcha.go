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
	pbcaptcha "github.com/openimsdk/protocol/captcha"
)

// GenerateCaptcha 请求服务端生成滑动验证码（对应 HTTP POST /captcha/generate）。
func (u *LoginMgr) GenerateCaptcha(ctx context.Context) (*pbcaptcha.GenerateCaptchaResp, error) {
	return api.GenerateCaptcha.Invoke(ctx, &pbcaptcha.GenerateCaptchaReq{})
}

// VerifyCaptcha 提交用户滑动结果（对应 HTTP POST /captcha/verify）。
func (u *LoginMgr) VerifyCaptcha(ctx context.Context, req *pbcaptcha.VerifyCaptchaReq) (*pbcaptcha.VerifyCaptchaResp, error) {
	return api.VerifyCaptcha.Invoke(ctx, req)
}

// GenerateClickCaptcha 请求服务端生成点选验证码（对应 HTTP POST /captcha/click_generate）。
func (u *LoginMgr) GenerateClickCaptcha(ctx context.Context) (*pbcaptcha.GenerateClickCaptchaResp, error) {
	return api.GenerateClickCaptcha.Invoke(ctx, &pbcaptcha.GenerateClickCaptchaReq{})
}

// VerifyClickCaptcha 提交用户点选结果（对应 HTTP POST /captcha/click_verify）。
func (u *LoginMgr) VerifyClickCaptcha(ctx context.Context, req *pbcaptcha.VerifyClickCaptchaReq) (*pbcaptcha.VerifyClickCaptchaResp, error) {
	return api.VerifyClickCaptcha.Invoke(ctx, req)
}

// GenerateCaptcha 在已 InitSDK 后即可调用，无需登录。
func GenerateCaptcha(callback open_im_sdk_callback.Base, operationID string) {
	call(callback, operationID, UserForSDK.GenerateCaptcha)
}

// VerifyCaptcha 参数 req 为 VerifyCaptchaReq 的 JSON，字段 captchaID、x、y。
func VerifyCaptcha(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.VerifyCaptcha, req)
}

// GenerateClickCaptcha 在已 InitSDK 后即可调用，无需登录。
func GenerateClickCaptcha(callback open_im_sdk_callback.Base, operationID string) {
	call(callback, operationID, UserForSDK.GenerateClickCaptcha)
}

// VerifyClickCaptcha 参数 req 为 VerifyClickCaptchaReq 的 JSON，字段 captchaID、dots。
func VerifyClickCaptcha(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.VerifyClickCaptcha, req)
}
