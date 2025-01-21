package telegram

import (
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
)

func (b *botController) AnswerCallbackQueryWithEditMessageMedia(
	callbackQuery *telegram.CallbackQuery,
	text string,
	photoURL string,
	replyMarkup any,
) error {
	log := b.container.GetLogger()
	if err := b.AnswerCallbackQuery(callbackQuery, nil, false); err != nil {
		log.Debug("fail to answer callback query", logger.FError(err))
		return err
	}
	if err := b.EditMessageMedia(callbackQuery, text, photoURL, replyMarkup); err != nil {
		log.Debug("fail to perform EditMessageMedia", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) AnswerCallbackQuery(callbackQuery *telegram.CallbackQuery, text *string, showAlert bool) error {
	log := b.container.GetLogger()
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      text,
		ShowAlert: showAlert,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		log.Debug("fail to send a AnswerCallbackQuery to telegram servers", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) EditMessageMedia(callbackQuery *telegram.CallbackQuery, text string, photoURL string, replyMarkup any) error {
	log := b.container.GetLogger()
	photoMedia := telegram.InputPhotoMedia{
		Type:      "photo",
		Media:     photoURL,
		Caption:   utils.NewString(text),
		ParseMode: utils.NewString("MarkdownV2"),
	}
	editMessageMedia := telegram.EditMessageMedia{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Media:       photoMedia,
		ReplyMarkup: replyMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMessageMedia, app.EditMessageMediaTelegramMethod); err != nil {
		log.Error("fail to edit message media", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) SendTextWithPhotoMedia(chatID int64, text string, photoURL string, replyMarkup any) error {
	log := b.container.GetLogger()
	resp := telegram.SendPhoto{
		ChatID:      chatID,
		Caption:     text,
		Photo:       photoURL,
		ParseMode:   utils.NewString("MarkdownV2"),
		ReplyMarkup: replyMarkup,
	}
	if err := b.telegramBotService.SendResponse(resp, app.SendPhotoTelegramMethod); err != nil {
		log.Debug("fail to send message with photo media", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) getPreferredLanguage(options *ContextOptions) string {
	defaultPreferredLanguage := "en"
	if options == nil {
		return defaultPreferredLanguage
	}
	preferredLanguage := options.Profile.PreferredLanguage
	if preferredLanguage == nil {
		return defaultPreferredLanguage
	}
	return *preferredLanguage
}

func (b *botController) deleteMessage(deleteMessage *telegram.DeleteMessage) error {
	log := b.container.GetLogger()
	if err := b.telegramBotService.SendResponse(deleteMessage, app.DeleteMessageTelegramMethod); err != nil {
		log.Error("fail to delete message in bot chat", logger.FError(err))
		return err
	}
	return nil
}
