package telegram

type SendResponse struct {
	ChatID              int64   `json:"chat_id"`
	Text                string  `json:"text"`
	ParseMode           *string `json:"parse_mode,omitempty"`
	ReplyMarkup         any     `json:"reply_markup,omitempty"`
	DisableNotification bool    `json:"disable_notification"`
}
