package app

import "errors"

var (
	NotTelegramCommandError          = errors.New("not telegramBot command error")
	NotSupportedTelegramCommandError = errors.New("not supported telegramBot command error")
	TelegramResponseBotError         = errors.New("telegram response bot error")
	UnknownLanguageError             = errors.New("unknown language error")
	EmptyUpdateError                 = errors.New("receive empty update from telegram servers")
	NilError                         = errors.New("nil pointer")
	UnknownValueError                = errors.New("unknown value error")
	EmptyValueError                  = errors.New("empty value")
	UnknownError                     = errors.New("unknown error")
	IndexOutOfRangeError             = errors.New("index out of range error")
	RequiredFieldError               = errors.New("required filed is missing error")
	DeleteInvoiceError               = errors.New("delete invoice error")
	UnknownPhoneNumberFormatError    = errors.New("unknown phone number format")
)
