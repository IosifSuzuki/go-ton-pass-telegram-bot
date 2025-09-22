package telegram

type Message struct {
	ID                int64              `json:"message_id"`
	From              *User              `json:"from"`
	Text              *string            `json:"text"`
	Chat              *Chat              `json:"chat"`
	Date              int64              `json:"date"`
	SuccessfulPayment *SuccessfulPayment `json:"successful_payment"`
	RefundedPayment   *RefundedPayment   `json:"refunded_payment"`
}
