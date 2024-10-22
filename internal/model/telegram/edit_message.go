package telegram

type EditMessage struct {
	ChatID          *int64  `json:"chat_id,omitempty"`
	MessageID       *int64  `json:"message_id,omitempty"`
	InlineMessageID *string `json:"inline_message_id,omitempty"`
	Text            string  `json:"text"`
	ParseMode       *string `json:"parse_mode,omitempty"`
	ReplyMarkup     any     `json:"reply_markup,omitempty"`
}
