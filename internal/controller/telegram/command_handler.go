package telegram

import (
	"context"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
)

func (b *botController) startTelegramCommandHandler(ctx context.Context, update *telegram.Update) error {
	userID := update.Message.From.ID
	langTag := b.getLanguageCode(ctx, update.Message.From)

	sendPhotoResp := telegram.SendPhoto{
		ChatID:    update.Message.Chat.ID,
		Photo:     "https://i.imghippo.com/files/vi44s1726518102.png",
		Caption:   b.container.GetLocalizer(langTag).LocalizedString("bot_markdown_description"),
		ParseMode: utils.NewString("MarkdownV2"),
	}
	if err := b.telegramBotService.SendResponse(sendPhotoResp, app.SendPhotoTelegramMethod); err != nil {
		return err
	}

	textResp := telegram.SendResponse{
		ChatID:      update.Message.Chat.ID,
		Text:        b.container.GetLocalizer(langTag).LocalizedString("select_preferred_language"),
		ReplyMarkup: b.telegramBotService.GetLanguagesReplyKeyboardMarkup(),
	}
	if err := b.sessionService.SaveBotStateForUser(ctx, app.SelectLanguageBotState, userID); err != nil {
		return err
	}

	return b.telegramBotService.SendResponse(textResp, app.SendMessageTelegramMethod)
}

func (b *botController) helpTelegramCommandHandler(ctx context.Context, update *telegram.Update) error {
	userID := update.Message.From.ID
	langTag := b.getLanguageCode(ctx, update.Message.From)

	if err := b.sessionService.ClearBotStateForUser(ctx, userID); err != nil {
		return err
	}

	resp := telegram.SendResponse{}
	resp.ChatID = update.Message.Chat.ID
	resp.Text = b.container.GetLocalizer(langTag).LocalizedString("help_cmd_text")
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}

func (b *botController) unknownTelegramCommandHandler(ctx context.Context, update *telegram.Update) error {
	userID := update.Message.From.ID
	langTag := b.getLanguageCode(ctx, update.Message.From)

	if err := b.sessionService.ClearBotStateForUser(ctx, userID); err != nil {
		return err
	}

	resp := telegram.SendResponse{}
	resp.ChatID = update.Message.Chat.ID
	resp.Text = b.container.GetLocalizer(langTag).LocalizedString("unknown_cmd_text")
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}
