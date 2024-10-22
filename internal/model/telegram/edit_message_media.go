package telegram

type EditMessageMedia struct {
	ChatID      *int64 `json:"chat_id"`
	MessageID   *int64 `json:"message_id"`
	Media       any    `json:"media"`
	ReplyMarkup any    `json:"reply_markup,omitempty"`
}
