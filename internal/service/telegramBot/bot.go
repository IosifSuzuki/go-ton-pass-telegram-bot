package telegramBot

import (
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
)

func NewTelegramBot(container container.Container) Bot {
	return &bot{
		container: container,
	}
}

func (b *bot) Processing(update *telegram.Update) (*telegram.SendResponse, error) {
	log := b.container.GetLogger()
	log.Debug("receive telegram message",
		logger.F("message", update.Message.Text),
		logger.F("user_id", update.Message.From.ID),
	)
	var response telegram.SendResponse
	response.ChatId = update.Message.Chat.ID
	telegramCmd, _ := b.parseUserCommand(update.Message.Text)
	switch telegramCmd {
	case app.StartTelegramCommand:
		response.Text = "Hello"
	case app.UnknownTelegramCommand:
		response.Text = b.container.LocalizedString("unknown_telegram_command")
	default:
		response.Text = b.container.LocalizedString("unknown_error")
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
