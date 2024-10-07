package telegram

type EditCaptionMessage struct {
	ChatID                *int64  `json:"chat_id,omitempty"`
	MessageID             *int64  `json:"message_id,omitempty"`
	Caption               *string `json:"caption,omitempty"`
	ParseMode             *string `json:"parse_mode,omitempty"`
	ShowCaptionAboveMedia bool    `json:"show_caption_above_media"`
	ReplyMarkup           any     `json:"reply_markup,omitempty"`
}
