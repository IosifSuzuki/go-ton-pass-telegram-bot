package telegram

import (
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
)

func (b *botController) GetCurrenciesReplyKeyboardMarkup() *telegram.ReplyKeyboardMarkup {
	log := b.container.GetLogger()
	currencies := b.container.GetConfig().AvailablePreferredCurrencies()

	log.Debug("AvailablePreferredCurrencies from configuration", logger.F("AvailablePreferredCurrencies", currencies))
	keyboardButtons := make([][]telegram.KeyboardButton, 0, len(currencies))
	for _, currency := range currencies {
		keyboardButtons = append(keyboardButtons, []telegram.KeyboardButton{
			{
				Text: utils.ShortCurrencyTextFormat(currency),
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

func (b *botController) GetLanguagesReplyKeyboardMarkup() *telegram.ReplyKeyboardMarkup {
	log := b.container.GetLogger()
	languages := b.container.GetConfig().AvailableLanguages()

	log.Debug("AvailableLanguages from configuration", logger.F("AvailableLanguages", languages))

	keyboardButtons := make([][]telegram.KeyboardButton, 0, len(languages))
	for _, language := range languages {
		buttonText := utils.LanguageTextFormat(language)
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
