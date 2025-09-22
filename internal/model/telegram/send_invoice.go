package telegram

type SendInvoice struct {
	ChatID         int64          `json:"chat_id"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	Payload        string         `json:"payload"`
	ProviderToken  string         `json:"provider_token"`
	Currency       string         `json:"currency"`
	Prices         []LabeledPrice `json:"prices"`
	PhotoURL       string         `json:"photo_url"`
	ProtectContent bool           `json:"protect_content"`
	ReplyMarkup    any            `json:"reply_markup"`
}
