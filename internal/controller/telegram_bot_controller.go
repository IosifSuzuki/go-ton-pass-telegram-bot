package controller

import (
	"context"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/pkg/logger"
)

type TelegramBotController interface {
	Serve(update *telegram.Update) error
}

type telegramBotController struct {
	container          container.Container
	telegramBotService service.TelegramBotService
	sessionService     service.SessionService
}

func NewTelegramBotController(container container.Container, sessionService service.SessionService) TelegramBotController {
	return &telegramBotController{
		container:          container,
		telegramBotService: service.NewTelegramBot(container),
		sessionService:     sessionService,
	}
}

func (t *telegramBotController) Serve(update *telegram.Update) error {
	log := t.container.GetLogger()
	ctx := context.Background()
	userID := update.Message.From.ID
	log.Debug("receive telegram message",
		logger.F("message", update.Message.Text),
		logger.F("user_id", userID),
		logger.F("language_code", update.Message.From.LanguageCode),
	)

	telegramCmd, err := t.telegramBotService.ParseTelegramCommand(update)
	switch telegramCmd {
	case app.StartTelegramCommand:
		return t.startTelegramCommandHandler(ctx, update)
	case app.HelpTelegramCommand:
		return t.helpTelegramCommandHandler(ctx, update)
	}
	if err != nil && telegramCmd == app.UnknownTelegramCommand {
		if err := t.sessionService.ClearBotStateForUser(ctx, userID); err != nil {
			return err
		}
		return t.unknownTelegramCommandHandler(ctx, update)
	}

	userBotState := t.sessionService.GetBotStateForUser(ctx, userID)
	switch userBotState {
	case app.SelectLanguageBotState:
		return t.userSelectedLanguageHandler(ctx, update)
	case app.SelectCurrencyBotState:
		return t.userSelectedCurrencyHandler(ctx, update)
	}

	return t.helpTelegramCommandHandler(ctx, update)
}

func (t *telegramBotController) startTelegramCommandHandler(ctx context.Context, update *telegram.Update) error {
	userID := update.Message.From.ID
	langTag := update.Message.From.LanguageCode

	if err := t.sessionService.SaveBotStateForUser(ctx, app.SelectLanguageBotState, userID); err != nil {
		return err
	}
	resp := telegram.SendResponse{}
	resp.ChatId = update.Message.Chat.ID
	resp.Text = t.container.GetLocalizer(langTag).LocalizedString("select_preferred_language")
	resp.ReplyMarkup = t.telegramBotService.GetLanguagesReplyKeyboardMarkup()
	return t.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}

func (t *telegramBotController) helpTelegramCommandHandler(ctx context.Context, update *telegram.Update) error {
	userID := update.Message.From.ID
	langTag := update.Message.From.LanguageCode

	if err := t.sessionService.ClearBotStateForUser(ctx, userID); err != nil {
		return err
	}

	resp := telegram.SendResponse{}
	resp.ChatId = update.Message.Chat.ID
	resp.Text = t.container.GetLocalizer(langTag).LocalizedString("help_cmd_text")
	return t.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}

func (t *telegramBotController) unknownTelegramCommandHandler(ctx context.Context, update *telegram.Update) error {
	userID := update.Message.From.ID
	langTag := update.Message.From.LanguageCode

	if err := t.sessionService.ClearBotStateForUser(ctx, userID); err != nil {
		return err
	}

	resp := telegram.SendResponse{}
	resp.ChatId = update.Message.Chat.ID
	resp.Text = t.container.GetLocalizer(langTag).LocalizedString("unknown_cmd_text")
	return t.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}

func (t *telegramBotController) userSelectedLanguageHandler(ctx context.Context, update *telegram.Update) error {
	userID := update.Message.From.ID
	langTag := update.Message.From.LanguageCode

	if err := t.sessionService.SaveBotStateForUser(ctx, app.SelectCurrencyBotState, userID); err != nil {
		return err
	}
	resp := telegram.SendResponse{}
	resp.ChatId = update.Message.Chat.ID
	resp.Text = t.container.GetLocalizer(langTag).LocalizedString("select_preferred_currency")
	resp.ReplyMarkup = t.telegramBotService.GetCurrenciesReplyKeyboardMarkup()
	return t.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}

func (t *telegramBotController) userSelectedCurrencyHandler(ctx context.Context, update *telegram.Update) error {
	userID := update.Message.From.ID
	langTag := update.Message.From.LanguageCode

	if err := t.sessionService.ClearBotStateForUser(ctx, userID); err != nil {
		return err
	}

	resp := telegram.SendResponse{}
	resp.ChatId = update.Message.Chat.ID
	resp.Text = t.container.GetLocalizer(langTag).LocalizedString("short_description")
	return t.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}
