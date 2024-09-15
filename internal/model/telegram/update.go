package telegram

type Update struct {
	ID            int64          `json:"update_id"`
	Message       Message        `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}
