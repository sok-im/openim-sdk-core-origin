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

var (
	SendPaymentNotification = newApi[SendPaymentNotificationReq, msg.SendMsgResp]("/msg/send_payment_notification")
)
