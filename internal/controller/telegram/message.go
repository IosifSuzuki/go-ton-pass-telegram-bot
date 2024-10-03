package telegram

import (
	"context"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
)

func (b *botController) messageToSelectLanguage(ctx context.Context, update *telegram.Update) error {
	telegramID, err := getTelegramID(update)
	if err != nil {
		return err
	}
	langTag, err := b.getLanguageCode(ctx, *update.Message.From)
	if err != nil {
		return err
	}
	textResp := telegram.SendPhoto{
		ChatID:      update.Message.Chat.ID,
		Photo:       "https://i.imghippo.com/files/vi44s1726518102.png",
		Caption:     b.container.GetLocalizer(*langTag).LocalizedString("select_preferred_language"),
		ReplyMarkup: b.telegramBotService.GetLanguagesReplyKeyboardMarkup(),
	}
	if err := b.sessionService.SaveBotStateForUser(ctx, app.SelectLanguageBotState, *telegramID); err != nil {
		return err
	}
	return b.telegramBotService.SendResponse(textResp, app.SendPhotoTelegramMethod)
}

func (b *botController) messageToSelectPreferredCurrency(ctx context.Context, update *telegram.Update) error {
	log := b.container.GetLogger()
	telegramID, err := getTelegramID(update)
	if err != nil {
		return err
	}
	log.Debug("prepare message messageToSelectPreferredCurrency")
	langTag, err := b.getLanguageCode(ctx, *update.Message.From)
	if err != nil {
		return err
	}
	resp := telegram.SendResponse{
		ChatID:      update.Message.Chat.ID,
		Text:        b.container.GetLocalizer(*langTag).LocalizedString("select_preferred_currency"),
		ReplyMarkup: b.telegramBotService.GetCurrenciesReplyKeyboardMarkup(),
	}
	log.Debug("prepared message messageToSelectPreferredCurrency")
	if err := b.sessionService.SaveBotStateForUser(ctx, app.SelectCurrencyBotState, *telegramID); err != nil {
		return err
	}
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}

func (b *botController) messageWelcome(ctx context.Context, update *telegram.Update) error {
	langTag, err := b.getLanguageCode(ctx, *update.Message.From)
	if err != nil {
		return err
	}
	sendPhotoResp := telegram.SendPhoto{
		ChatID:    update.Message.Chat.ID,
		Photo:     "https://i.imghippo.com/files/vi44s1726518102.png",
		Caption:   b.container.GetLocalizer(*langTag).LocalizedString("bot_markdown_description"),
		ParseMode: utils.NewString("MarkdownV2"),
	}
	return b.telegramBotService.SendResponse(sendPhotoResp, app.SendPhotoTelegramMethod)
}

func (b *botController) messageMainMenu(ctx context.Context, update *telegram.Update) error {
	langTag, err := b.getLanguageCode(ctx, *update.Message.From)
	if err != nil {
		return err
	}
	mainMenuInlineKeyboardMarkup, err := b.getMainMenuInlineKeyboardMarkup(ctx, *update.Message.From)
	if err != nil {
		return err
	}
	resp := telegram.SendResponse{
		ChatID:      update.Message.Chat.ID,
		Text:        b.container.GetLocalizer(*langTag).LocalizedString("short_description"),
		ReplyMarkup: mainMenuInlineKeyboardMarkup,
	}
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}

func (b *botController) messageListPayCurrencies(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		return err
	}
	localizer := b.container.GetLocalizer(*langTag)
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		log.Error("fail to send a AnswerCallbackQuery to telegram servers", logger.FError(err))
		return err
	}
	payCurrenciesInlineKeyboardMarkup, err := b.getPayCurrenciesInlineKeyboardMarkup(*langTag)
	if err != nil {
		return err
	}
	resp := telegram.SendResponse{
		ChatID:      callbackQuery.Message.Chat.ID,
		Text:        localizer.LocalizedString("select_currency_to_pay"),
		ReplyMarkup: payCurrenciesInlineKeyboardMarkup,
	}
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}

func (b *botController) messageEnterAmountCurrency(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	telegramID := callbackQuery.From.ID
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		return err
	}
	localizer := b.container.GetLocalizer(*langTag)
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		log.Error("fail to send a AnswerCallbackQuery to telegram servers", logger.FError(err))
		return err
	}
	resp := telegram.SendResponse{
		ChatID:      callbackQuery.Message.Chat.ID,
		Text:        localizer.LocalizedString("enter_amount_for_payment_in_currency"),
		ReplyMarkup: nil,
	}
	if err := b.sessionService.SaveBotStateForUser(ctx, app.EnteringAmountCurrencyBotState, telegramID); err != nil {
		return err
	}
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}

func (b *botController) messageWithPlainText(_ context.Context, text string, update *telegram.Update) error {
	resp := telegram.SendResponse{
		ChatID: update.Message.Chat.ID,
		Text:   text,
	}
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}
