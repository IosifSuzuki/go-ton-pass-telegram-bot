package telegram

import (
	"context"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
)

func (b *botController) startTelegramCommandHandler(ctx context.Context, update *telegram.Update) error {
	userID := update.Message.From.ID
	langTag := update.Message.From.LanguageCode

	if err := b.sessionService.SaveBotStateForUser(ctx, app.SelectLanguageBotState, userID); err != nil {
		return err
	}
	resp := telegram.SendResponse{}
	resp.ChatId = update.Message.Chat.ID
	resp.Text = b.container.GetLocalizer(langTag).LocalizedString("select_preferred_language")
	resp.ReplyMarkup = b.telegramBotService.GetLanguagesReplyKeyboardMarkup()
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}

func (b *botController) helpTelegramCommandHandler(ctx context.Context, update *telegram.Update) error {
	userID := update.Message.From.ID
	langTag := update.Message.From.LanguageCode

	if err := b.sessionService.ClearBotStateForUser(ctx, userID); err != nil {
		return err
	}

	resp := telegram.SendResponse{}
	resp.ChatId = update.Message.Chat.ID
	resp.Text = b.container.GetLocalizer(langTag).LocalizedString("help_cmd_text")
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}

func (b *botController) unknownTelegramCommandHandler(ctx context.Context, update *telegram.Update) error {
	userID := update.Message.From.ID
	langTag := update.Message.From.LanguageCode

	if err := b.sessionService.ClearBotStateForUser(ctx, userID); err != nil {
		return err
	}

	resp := telegram.SendResponse{}
	resp.ChatId = update.Message.Chat.ID
	resp.Text = b.container.GetLocalizer(langTag).LocalizedString("unknown_cmd_text")
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}

func (b *botController) userSelectedLanguageCommandHandler(ctx context.Context, update *telegram.Update) error {
	userID := update.Message.From.ID
	langTag := update.Message.From.LanguageCode

	if err := b.sessionService.SaveBotStateForUser(ctx, app.SelectCurrencyBotState, userID); err != nil {
		return err
	}
	resp := telegram.SendResponse{}
	resp.ChatId = update.Message.Chat.ID
	resp.Text = b.container.GetLocalizer(langTag).LocalizedString("select_preferred_currency")
	resp.ReplyMarkup = b.telegramBotService.GetCurrenciesReplyKeyboardMarkup()
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}

func (b *botController) userSelectedCurrencyCommandHandler(ctx context.Context, update *telegram.Update) error {
	userID := update.Message.From.ID
	langTag := update.Message.From.LanguageCode

	if err := b.sessionService.ClearBotStateForUser(ctx, userID); err != nil {
		return err
	}

	resp := telegram.SendResponse{}
	resp.ChatId = update.Message.Chat.ID
	resp.Text = b.container.GetLocalizer(langTag).LocalizedString("short_description")
	resp.ReplyMarkup = b.telegramBotService.GetMenuInlineKeyboardMarkup(langTag)
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}
