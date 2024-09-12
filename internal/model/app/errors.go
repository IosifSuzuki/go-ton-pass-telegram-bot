package app

import "errors"

var (
	NotTelegramCommandError          = errors.New("not telegramBot command error")
	NotSupportedTelegramCommandError = errors.New("not supported telegramBot command error")
	TelegramResponseBotError         = errors.New("telegram response bot error")
)
