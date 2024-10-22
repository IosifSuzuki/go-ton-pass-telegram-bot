package app

type TelegramCommand uint

const (
	NotTelegramCommand TelegramCommand = iota
	StartTelegramCommand
	HelpTelegramCommand
	UnknownTelegramCommand
)
