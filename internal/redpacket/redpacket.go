package redpacket

import (
	"context"

	"github.com/openimsdk/openim-sdk-core/v3/pkg/api"
	pbredpacket "github.com/openimsdk/protocol/redpacket"
)

// RedPacket exposes HTTP API calls to the server's /redpacket routes (requires login; user identity comes from token).
type RedPacket struct{}

func NewRedPacket() *RedPacket {
	return &RedPacket{}
}

func (r *RedPacket) CreateOrder(ctx context.Context, req *pbredpacket.CreateOrderReq) (*pbredpacket.CreateOrderResp, error) {
	return api.RedPacketCreateOrder.Invoke(ctx, req)
}

func (r *RedPacket) CreatedCallback(ctx context.Context, req *pbredpacket.CreatedCallbackReq) (*pbredpacket.CreatedCallbackResp, error) {
	return api.RedPacketCreatedCallback.Invoke(ctx, req)
}

func (r *RedPacket) GetDetail(ctx context.Context, req *pbredpacket.GetDetailReq) (*pbredpacket.GetDetailResp, error) {
	return api.RedPacketGetDetail.Invoke(ctx, req)
}

func (r *RedPacket) IssueClaimSign(ctx context.Context, req *pbredpacket.IssueClaimSignReq) (*pbredpacket.IssueClaimSignResp, error) {
	return api.RedPacketIssueClaimSign.Invoke(ctx, req)
}

func (r *RedPacket) ClaimResult(ctx context.Context, req *pbredpacket.ClaimResultReq) (*pbredpacket.ClaimResultResp, error) {
	return api.RedPacketClaimResult.Invoke(ctx, req)
}

func (r *RedPacket) RequestRefund(ctx context.Context, req *pbredpacket.RequestRefundReq) (*pbredpacket.RequestRefundResp, error) {
	return api.RedPacketRequestRefund.Invoke(ctx, req)
}

func (r *RedPacket) GetRefund(ctx context.Context, req *pbredpacket.GetRefundReq) (*pbredpacket.GetRefundResp, error) {
	return api.RedPacketGetRefund.Invoke(ctx, req)
}

func (r *RedPacket) IssueWalletBindChallenge(ctx context.Context, req *pbredpacket.IssueWalletBindChallengeReq) (*pbredpacket.IssueWalletBindChallengeResp, error) {
	return api.RedPacketIssueWalletBindChallenge.Invoke(ctx, req)
}

func (r *RedPacket) ConfirmWalletBind(ctx context.Context, req *pbredpacket.ConfirmWalletBindReq) (*pbredpacket.ConfirmWalletBindResp, error) {
	return api.RedPacketConfirmWalletBind.Invoke(ctx, req)
}

func (r *RedPacket) GetWalletBinding(ctx context.Context, req *pbredpacket.GetWalletBindingReq) (*pbredpacket.GetWalletBindingResp, error) {
	return api.RedPacketGetWalletBinding.Invoke(ctx, req)
}
