package telegram

import (
	"context"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
)

func (b *botController) messageToSelectLanguage(ctx context.Context, update *telegram.Update) error {
	telegramID := getTelegramID(update)
	langTag := b.getLanguageCode(ctx, update.Message.From)
	textResp := telegram.SendPhoto{
		ChatID:      update.Message.Chat.ID,
		Photo:       "https://i.imghippo.com/files/vi44s1726518102.png",
		Caption:     b.container.GetLocalizer(langTag).LocalizedString("select_preferred_language"),
		ReplyMarkup: b.telegramBotService.GetLanguagesReplyKeyboardMarkup(),
	}
	if err := b.sessionService.SaveBotStateForUser(ctx, app.SelectLanguageBotState, telegramID); err != nil {
		return err
	}
	return b.telegramBotService.SendResponse(textResp, app.SendPhotoTelegramMethod)
}

func (b *botController) messageToSelectCurrency(ctx context.Context, update *telegram.Update) error {
	telegramID := getTelegramID(update)
	langTag := b.getLanguageCode(ctx, update.Message.From)
	resp := telegram.SendResponse{
		ChatID:      update.Message.Chat.ID,
		Text:        b.container.GetLocalizer(langTag).LocalizedString("select_preferred_currency"),
		ReplyMarkup: b.telegramBotService.GetCurrenciesReplyKeyboardMarkup(),
	}
	if err := b.sessionService.SaveBotStateForUser(ctx, app.SelectCurrencyBotState, telegramID); err != nil {
		return err
	}
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}

func (b *botController) messageWelcome(ctx context.Context, update *telegram.Update) error {
	langTag := b.getLanguageCode(ctx, update.Message.From)
	sendPhotoResp := telegram.SendPhoto{
		ChatID:    update.Message.Chat.ID,
		Photo:     "https://i.imghippo.com/files/vi44s1726518102.png",
		Caption:   b.container.GetLocalizer(langTag).LocalizedString("bot_markdown_description"),
		ParseMode: utils.NewString("MarkdownV2"),
	}
	return b.telegramBotService.SendResponse(sendPhotoResp, app.SendPhotoTelegramMethod)
}

func (b *botController) messageMainMenu(ctx context.Context, update *telegram.Update) error {
	langTag := b.getLanguageCode(ctx, update.Message.From)
	resp := telegram.SendResponse{
		ChatID:      update.Message.Chat.ID,
		Text:        b.container.GetLocalizer(langTag).LocalizedString("short_description"),
		ReplyMarkup: b.telegramBotService.GetMenuInlineKeyboardMarkup(langTag),
	}
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}

func (b *botController) messageWithPlainText(ctx context.Context, text string, update *telegram.Update) error {
	resp := telegram.SendResponse{
		ChatID: update.Message.Chat.ID,
		Text:   text,
	}
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}
