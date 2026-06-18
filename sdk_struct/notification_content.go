package sdk_struct

const (
	ServiceNotificationSubTypeSecurity = 1
	ServiceNotificationSubTypeAccount  = 2
	ServiceNotificationSubTypeSystem   = 3
	ServiceNotificationSubTypeUpdate   = 4

	PaymentTransactionTypeTransfer            = "转账"
	PaymentTransactionTypeRedPacket           = "红包"
	PaymentTransactionTypeRedPacketTransfer   = "红包/转账"
	DefaultNotificationDetailText             = "查看详情"
)

// PaymentNotificationAction 钱包通知底部次要操作（如「去赎回」）。
type PaymentNotificationAction struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

// ServiceNotificationContent SOK 服务通知卡片内容。
type ServiceNotificationContent struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	DetailURL  string `json:"detailURL,omitempty"`
	DetailText string `json:"detailText,omitempty"`
	SubType    int32  `json:"subType,omitempty"`
}

// PaymentNotificationContent SOK 钱包通知卡片内容。
type PaymentNotificationContent struct {
	Title           string                     `json:"title"`
	Amount          string                     `json:"amount"`
	TransactionType string                     `json:"transactionType"`
	TransactionTime string                     `json:"transactionTime"`
	Currency        string                     `json:"currency"`
	CurrencyIconURL string                     `json:"currencyIconURL,omitempty"`
	DetailURL       string                     `json:"detailURL,omitempty"`
	DetailText      string                     `json:"detailText,omitempty"`
	SecondaryAction *PaymentNotificationAction `json:"secondaryAction,omitempty"`
	OrderNo         string                     `json:"orderNo,omitempty"`
	BizID           string                     `json:"bizID,omitempty"`
}
