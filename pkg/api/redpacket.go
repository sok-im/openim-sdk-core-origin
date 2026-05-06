package api

import "github.com/openimsdk/protocol/redpacket"

var (
	RedPacketCreateOrder               = newApi[redpacket.CreateOrderReq, redpacket.CreateOrderResp]("/redpacket/create_order")
	RedPacketCreatedCallback           = newApi[redpacket.CreatedCallbackReq, redpacket.CreatedCallbackResp]("/redpacket/created_callback")
	RedPacketGetDetail                 = newApi[redpacket.GetDetailReq, redpacket.GetDetailResp]("/redpacket/detail")
	RedPacketIssueClaimSign            = newApi[redpacket.IssueClaimSignReq, redpacket.IssueClaimSignResp]("/redpacket/issue_claim_sign")
	RedPacketClaimResult               = newApi[redpacket.ClaimResultReq, redpacket.ClaimResultResp]("/redpacket/claim_result")
	RedPacketRequestRefund             = newApi[redpacket.RequestRefundReq, redpacket.RequestRefundResp]("/redpacket/request_refund")
	RedPacketGetRefund                 = newApi[redpacket.GetRefundReq, redpacket.GetRefundResp]("/redpacket/get_refund")
	RedPacketIssueWalletBindChallenge  = newApi[redpacket.IssueWalletBindChallengeReq, redpacket.IssueWalletBindChallengeResp]("/redpacket/wallet_bind/challenge")
	RedPacketConfirmWalletBind         = newApi[redpacket.ConfirmWalletBindReq, redpacket.ConfirmWalletBindResp]("/redpacket/wallet_bind/confirm")
	RedPacketGetWalletBinding          = newApi[redpacket.GetWalletBindingReq, redpacket.GetWalletBindingResp]("/redpacket/wallet_bind/detail")
)
