package manager

import (
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
)

type TelegramInlineButtonBuilder interface {
	SetText(text string) TelegramInlineButtonBuilder
	SetCommandName(commandName string) TelegramInlineButtonBuilder
	SetParameters(parameters []any) TelegramInlineButtonBuilder
	SetLink(url string) TelegramInlineButtonBuilder
	Build() (*telegram.InlineKeyboardButton, error)
	SetPay(pay bool) TelegramInlineButtonBuilder
}

type telegramInlineButtonBuilder struct {
	text        *string
	url         *string
	commandName *string
	pay         bool
	parameters  *[]any
}

func NewTelegramInlineButtonBuilder() TelegramInlineButtonBuilder {
	return &telegramInlineButtonBuilder{
		text:        nil,
		commandName: nil,
		parameters:  nil,
	}
}

func (t *telegramInlineButtonBuilder) SetText(text string) TelegramInlineButtonBuilder {
	t.text = utils.NewString(text)
	return t
}

func (t *telegramInlineButtonBuilder) SetLink(url string) TelegramInlineButtonBuilder {
	t.url = utils.NewString(url)
	return t
}

func (t *telegramInlineButtonBuilder) SetCommandName(commandName string) TelegramInlineButtonBuilder {
	t.commandName = utils.NewString(commandName)
	return t
}

func (t *telegramInlineButtonBuilder) SetParameters(parameters []any) TelegramInlineButtonBuilder {
	t.parameters = &parameters
	return t
}

func (t *telegramInlineButtonBuilder) SetPay(pay bool) TelegramInlineButtonBuilder {
	t.pay = pay
	return t
}

func (t *telegramInlineButtonBuilder) Build() (*telegram.InlineKeyboardButton, error) {
	if t.text == nil {
		return nil, app.RequiredFieldError
	} else if t.text == nil && t.url == nil {
		return nil, app.RequiredFieldError
	}
	var (
		data *string
		text = *t.text
		url  = t.url
	)
	if t.commandName != nil {
		callbackData := app.TelegramCallbackData{
			Name:       *t.commandName,
			Parameters: t.parameters,
		}
		encodedCallbackData, err := utils.EncodeTelegramCallbackData(callbackData)
		if err != nil {
			return nil, err
		}
		data = encodedCallbackData
	}
	return &telegram.InlineKeyboardButton{
		Text: text,
		URL:  url,
		Data: data,
		Pay:  t.pay,
	}, nil
}
