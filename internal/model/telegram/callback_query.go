package telegram

type CallbackQuery struct {
	ID              string   `json:"id"`
	InlineMessageID *string  `json:"inline_message_id,omitempty"`
	From            User     `json:"from"`
	Message         *Message `json:"message,omitempty"`
	Data            string   `json:"data"`
}
