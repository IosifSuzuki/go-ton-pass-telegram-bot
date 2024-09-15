package telegram

type InlineKeyboardButton struct {
	Text string  `json:"text"`
	Data *string `json:"callback_data,omitempty"`
}
