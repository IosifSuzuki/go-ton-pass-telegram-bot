package app

import "errors"

var (
	NotTelegramCommandError          = errors.New("not telegramBot command error")
	NotSupportedTelegramCommandError = errors.New("not supported telegramBot command error")
	TelegramResponseBotError         = errors.New("telegram response bot error")
	UnknownLanguageError             = errors.New("unknown language error")
	EmptyUpdateError                 = errors.New("receive empty update from telegram servers")
	NilError                         = errors.New("nil pointer")
)
