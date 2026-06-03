package api

import "github.com/openimsdk/protocol/totp"

var (
	TotpGetSecret  = newApi[totp.GetSecretReq, totp.GetSecretResp]("/totp/secret")
	TotpBind       = newApi[totp.BindTotpReq, totp.BindTotpResp]("/totp/bind")
	TotpVerify     = newApi[totp.VerifyTotpReq, totp.VerifyTotpResp]("/totp/verify")
	TotpGetStatus  = newApi[totp.GetStatusReq, totp.GetStatusResp]("/totp/status")
	TotpUnbind     = newApi[totp.UnbindTotpReq, totp.UnbindTotpResp]("/totp/unbind")
)
