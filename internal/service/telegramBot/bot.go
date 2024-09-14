package telegramBot

import (
	"fmt"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/pkg/logger"
	"strings"
)

type Bot interface {
	Processing(update *telegram.Update) (*telegram.SendResponse, error)
}

type bot struct {
	container container.Container
}

const (
	startCmdText = "/start"
	helpCmdText  = "/help"
)

func NewTelegramBot(container container.Container) Bot {
	return &bot{
		container: container,
	}
}

func (b *bot) Processing(update *telegram.Update) (*telegram.SendResponse, error) {
	localizer := b.container.GetLocalizer(update.Message.From.LanguageCode)
	log := b.container.GetLogger()
	log.Debug("receive telegram message",
		logger.F("message", update.Message.Text),
		logger.F("user_id", update.Message.From.ID),
		logger.F("language_code", update.Message.From.LanguageCode),
	)

	var response telegram.SendResponse
	response.DisableNotification = true
	response.ChatId = update.Message.Chat.ID
	telegramCmd, _ := b.parseUserCommand(update.Message.Text)
	switch telegramCmd {
	case app.StartTelegramCommand:
		response.Text = localizer.LocalizedString("select_preferred_language")
		response.ReplyMarkup = b.prepareLanguagesReplyKeyboardMarkup()
	case app.UnknownTelegramCommand:
		response.Text = localizer.LocalizedString("unknown_telegram_command")
		response.ReplyMarkup = nil
	default:
		response.Text = localizer.LocalizedString("unknown_error")
		response.ReplyMarkup = nil
	}

	return &response, nil
}

func (b *bot) parseUserCommand(text string) (app.TelegramCommand, error) {
	switch text {
	case startCmdText:
		return app.StartTelegramCommand, nil
	default:
		if strings.HasPrefix(text, "/") {
			return app.UnknownTelegramCommand, app.NotSupportedTelegramCommandError
		}
		return app.NotTelegramCommand, app.NotTelegramCommandError
	}
}

func (b *bot) prepareLanguagesReplyKeyboardMarkup() *telegram.ReplyKeyboardMarkup {
	log := b.container.GetLogger()

	languages := b.container.GetConfig().AvailableLanguages()
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
