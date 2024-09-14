package telegram

type SendResponse struct {
	ChatId              int64  `json:"chat_id"`
	Text                string `json:"text"`
	ReplyMarkup         any    `json:"reply_markup,omitempty"`
	DisableNotification bool   `json:"disable_notification"`
}
