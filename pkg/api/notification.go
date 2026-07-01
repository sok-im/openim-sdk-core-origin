package api

import (
	"github.com/openimsdk/openim-sdk-core/v3/sdk_struct"
	"github.com/openimsdk/protocol/msg"
)

// SendPaymentNotificationReq matches HTTP POST /msg/send_payment_notification.
// sendUserID must be a registered notification account.
type SendPaymentNotificationReq struct {
	SendUserID string                              `json:"sendUserID"`
	RecvUserID string                              `json:"recvUserID"`
	Content    sdk_struct.PaymentNotificationContent `json:"content"`
}

type WalletDualPartyNotifyReq struct {
	SenderUserID   string `json:"senderUserID"`
	ReceiverUserID string `json:"receiverUserID"`
	BizID          string `json:"bizID"`
	ReceiverText   string `json:"receiverText"`
	SenderText     string `json:"senderText"`
	DetailURL      string `json:"detailURL,omitempty"`
}

type WalletExpiredNotifyReq struct {
	SenderUserID string `json:"senderUserID"`
	BizID        string `json:"bizID"`
	Text         string `json:"text"`
	DetailURL    string `json:"detailURL,omitempty"`
}

type WalletNotifyResp struct{}

var (
	SendPaymentNotification    = newApi[SendPaymentNotificationReq, msg.SendMsgResp]("/msg/send_payment_notification")
	NotifyRedPacketClaimed     = newApi[WalletDualPartyNotifyReq, WalletNotifyResp]("/msg/notify_red_packet_claimed")
	NotifyTransferReceived     = newApi[WalletDualPartyNotifyReq, WalletNotifyResp]("/msg/notify_transfer_received")
	NotifyRedPacketExpired     = newApi[WalletExpiredNotifyReq, WalletNotifyResp]("/msg/notify_red_packet_expired")
	NotifyTransferExpired      = newApi[WalletExpiredNotifyReq, WalletNotifyResp]("/msg/notify_transfer_expired")
)
