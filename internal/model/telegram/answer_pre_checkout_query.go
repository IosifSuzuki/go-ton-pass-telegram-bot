package telegram

type AnswerPreCheckoutQuery struct {
	PreCheckoutQueryID string `json:"pre_checkout_query_id"`
	OK                 bool   `json:"ok"`
}
