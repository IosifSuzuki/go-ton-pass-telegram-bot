package sms

type WebhookUpdates struct {
	ActivationID int64  `json:"activationId"`
	Service      string `json:"service"`
	Text         string `json:"text"`
	Code         string `json:"code"`
	Country      int64  `json:"country"`
	ReceivedAt   string `json:"receivedAt"`
}
