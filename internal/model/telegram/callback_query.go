package telegram

type CallbackQuery struct {
	ID   string `json:"id"`
	From User   `json:"from"`
	Data string `json:"data"`
}
