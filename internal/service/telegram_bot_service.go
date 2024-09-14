package service

import (
	"fmt"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/pkg/logger"
	"strings"
)

type TelegramBotService interface {
	ParseTelegramCommand(update *telegram.Update) (app.TelegramCommand, error)
	GetLanguagesReplyKeyboardMarkup() *telegram.ReplyKeyboardMarkup
	GetCurrenciesReplyKeyboardMarkup() *telegram.ReplyKeyboardMarkup
}

type telegramBotService struct {
	container container.Container
}

const (
	startCmdText = "/start"
	helpCmdText  = "/help"
)

func NewTelegramBot(container container.Container) TelegramBotService {
	return &telegramBotService{
		container: container,
	}
}

func (t *telegramBotService) ParseTelegramCommand(update *telegram.Update) (app.TelegramCommand, error) {
	text := update.Message.Text
	return parseTelegramCommand(text)
}

func (t *telegramBotService) GetLanguagesReplyKeyboardMarkup() *telegram.ReplyKeyboardMarkup {
	log := t.container.GetLogger()
	languages := t.container.GetConfig().AvailableLanguages()

	log.Debug("AvailableLanguages from configuration", logger.F("AvailableLanguages", languages))

	keyboardButtons := make([][]telegram.KeyboardButton, 0, len(languages))
	for _, language := range languages {
		buttonText := fmt.Sprintf("%s %s", language.FlagEmoji, language.NativeName)
		keyboardButtons = append(keyboardButtons, []telegram.KeyboardButton{
			{
				Text: buttonText,
			},
		})
	}
	return &telegram.ReplyKeyboardMarkup{
		Keyboard:                  keyboardButtons,
		PersistentDisplayKeyboard: false,
		ResizeKeyboard:            true,
		OneTimeKeyboard:           true,
		Placeholder:               nil,
	}
}

func (t *telegramBotService) GetCurrenciesReplyKeyboardMarkup() *telegram.ReplyKeyboardMarkup {
	log := t.container.GetLogger()
	currencies := t.container.GetConfig().AvailableCurrencies()

	log.Debug("AvailableCurrencies from configuration", logger.F("AvailableCurrencies", currencies))
	keyboardButtons := make([][]telegram.KeyboardButton, 0, len(currencies))
	for _, currency := range currencies {
		buttonText := fmt.Sprintf("%s", currency.ABBR)
		keyboardButtons = append(keyboardButtons, []telegram.KeyboardButton{
			{
				Text: buttonText,
			},
		})
	}
	return &telegram.ReplyKeyboardMarkup{
		Keyboard:                  keyboardButtons,
		PersistentDisplayKeyboard: false,
		ResizeKeyboard:            true,
		OneTimeKeyboard:           true,
		Placeholder:               nil,
	}
}

func parseTelegramCommand(text string) (app.TelegramCommand, error) {
	switch text {
	case startCmdText:
		return app.StartTelegramCommand, nil
	case helpCmdText:
		return app.HelpTelegramCommand, nil
	default:
		break
	}
	if strings.HasPrefix(text, "/") {
		return app.UnknownTelegramCommand, app.NotSupportedTelegramCommandError
	}
	return app.NotTelegramCommand, app.NotTelegramCommandError
}
