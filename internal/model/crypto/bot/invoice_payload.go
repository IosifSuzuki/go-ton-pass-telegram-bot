package bot

type InvoicePayload struct {
	ChatID     int64 `json:"chat_id"`
	TelegramID int64 `json:"telegram_id"`
}
