package telegram

type SetMyCommands struct {
	Commands []BotCommand `json:"commands"`
}
