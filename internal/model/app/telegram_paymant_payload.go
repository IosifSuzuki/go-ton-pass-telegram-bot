package app

type TelegramPaymentPayload struct {
	CreditBalance float64 `json:"credit_balance"`
	ProfileID     int64   `json:"profile_id"`
}
