package telegram

type InlineKeyboardButton struct {
	Text string  `json:"text"`
	URL  *string `json:"url,omitempty"`
	Data *string `json:"callback_data,omitempty"`
	Pay  bool    `json:"pay,omitempty"`
}
