package open_im_sdk

import "github.com/openimsdk/openim-sdk-core/v3/open_im_sdk_callback"

// RedPacketCreateOrder calls POST /redpacket/create_order. req: JSON of redpacket.CreateOrderReq.
func RedPacketCreateOrder(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.RedPacket().CreateOrder, req)
}

// RedPacketCreatedCallback calls POST /redpacket/created_callback. req: JSON of redpacket.CreatedCallbackReq.
func RedPacketCreatedCallback(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.RedPacket().CreatedCallback, req)
}

// RedPacketGetDetail calls POST /redpacket/detail. req: JSON of redpacket.GetDetailReq.
func RedPacketGetDetail(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.RedPacket().GetDetail, req)
}

// RedPacketIssueClaimSign calls POST /redpacket/issue_claim_sign. req: JSON of redpacket.IssueClaimSignReq.
func RedPacketIssueClaimSign(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.RedPacket().IssueClaimSign, req)
}

// RedPacketClaimResult calls POST /redpacket/claim_result. req: JSON of redpacket.ClaimResultReq.
func RedPacketClaimResult(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.RedPacket().ClaimResult, req)
}

// RedPacketRequestRefund calls POST /redpacket/request_refund. req: JSON of redpacket.RequestRefundReq.
func RedPacketRequestRefund(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.RedPacket().RequestRefund, req)
}

// RedPacketGetRefund calls POST /redpacket/get_refund. req: JSON of redpacket.GetRefundReq.
func RedPacketGetRefund(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.RedPacket().GetRefund, req)
}

// RedPacketIssueWalletBindChallenge calls POST /redpacket/wallet_bind/challenge. req: JSON of redpacket.IssueWalletBindChallengeReq.
func RedPacketIssueWalletBindChallenge(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.RedPacket().IssueWalletBindChallenge, req)
}

// RedPacketConfirmWalletBind calls POST /redpacket/wallet_bind/confirm. req: JSON of redpacket.ConfirmWalletBindReq.
func RedPacketConfirmWalletBind(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.RedPacket().ConfirmWalletBind, req)
}

// RedPacketGetWalletBinding calls POST /redpacket/wallet_bind/detail. req: JSON of redpacket.GetWalletBindingReq.
func RedPacketGetWalletBinding(callback open_im_sdk_callback.Base, operationID string, req string) {
	call(callback, operationID, UserForSDK.RedPacket().GetWalletBinding, req)
}
