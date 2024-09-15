package telegram

import (
	"context"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/pkg/logger"
)

type BotController interface {
	Serve(update *telegram.Update) error
}

type botController struct {
	container          container.Container
	telegramBotService service.TelegramBotService
	sessionService     service.SessionService
}

func NewBotController(container container.Container, sessionService service.SessionService) BotController {
	return &botController{
		container:          container,
		telegramBotService: service.NewTelegramBot(container),
		sessionService:     sessionService,
	}
}

func (b *botController) Serve(update *telegram.Update) error {
	log := b.container.GetLogger()
	ctx := context.Background()
	userID := update.Message.From.ID
	log.Debug("receive telegram message", logger.F("update", update))

	telegramCmd, err := b.telegramBotService.ParseTelegramCommand(update)
	switch telegramCmd {
	case app.StartTelegramCommand:
		return b.startTelegramCommandHandler(ctx, update)
	case app.HelpTelegramCommand:
		return b.helpTelegramCommandHandler(ctx, update)
	}
	if err != nil && telegramCmd == app.UnknownTelegramCommand {
		if err := b.sessionService.ClearBotStateForUser(ctx, userID); err != nil {
			return err
		}
		return b.unknownTelegramCommandHandler(ctx, update)
	}

	userBotState := b.sessionService.GetBotStateForUser(ctx, userID)
	switch userBotState {
	case app.SelectLanguageBotState:
		return b.userSelectedLanguageCommandHandler(ctx, update)
	case app.SelectCurrencyBotState:
		return b.userSelectedCurrencyCommandHandler(ctx, update)
	}

	callbackQueryCommand := b.telegramBotService.ParseCallbackQueryCommand(update)
	switch callbackQueryCommand {
	case app.BalanceCallbackQueryCommand:
		return b.balanceCallbackQueryCommandHandler(ctx, update.CallbackQuery)
	case app.HelpCallbackQueryCommand, app.HistoryCallbackQueryCommand, app.BuyNumberCallbackQueryCommand, app.LanguageCallbackQueryCommand:
		return b.unsupportedCallbackQueryCommandHandle(ctx, update.CallbackQuery)
	}

	return b.helpTelegramCommandHandler(ctx, update)
}
