package telegram

type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}
