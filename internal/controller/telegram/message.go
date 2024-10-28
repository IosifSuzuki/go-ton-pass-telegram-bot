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
	localizer := b.container.GetLocalizer(langTag)
	textResp := telegram.SendPhoto{
		ChatID:      update.Message.Chat.ID,
		Photo:       selectPreferredLanguageImageURL,
		Caption:     localizer.LocalizedString("select_preferred_language"),
		ReplyMarkup: b.GetLanguagesReplyKeyboardMarkup(),
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
	localizer := b.container.GetLocalizer(langTag)
	resp := telegram.SendPhoto{
		ChatID:      update.Message.Chat.ID,
		Photo:       selectPreferredCurrencyImageURL,
		Caption:     localizer.LocalizedString("select_preferred_currency"),
		ReplyMarkup: b.GetCurrenciesReplyKeyboardMarkup(),
	}
	log.Debug("prepared message messageToSelectPreferredCurrency")
	if err := b.sessionService.SaveBotStateForUser(ctx, app.SelectCurrencyBotState, *telegramID); err != nil {
		return err
	}
	return b.telegramBotService.SendResponse(resp, app.SendPhotoTelegramMethod)
}

func (b *botController) messageWelcome(ctx context.Context, update *telegram.Update) error {
	langTag, err := b.getLanguageCode(ctx, *update.Message.From)
	if err != nil {
		return err
	}
	sendPhotoResp := telegram.SendPhoto{
		ChatID:    update.Message.Chat.ID,
		Photo:     welcomeImageURL,
		Caption:   b.container.GetLocalizer(langTag).LocalizedString("bot_markdown_description"),
		ParseMode: utils.NewString("MarkdownV2"),
	}
	return b.telegramBotService.SendResponse(sendPhotoResp, app.SendPhotoTelegramMethod)
}

func (b *botController) messageMainMenu(ctx context.Context, update *telegram.Update) error {
	langTag, err := b.getLanguageCode(ctx, *update.Message.From)
	if err != nil {
		return err
	}
	mainMenuInlineKeyboardMarkup, err := b.keyboardManager.MainMenuInlineKeyboardMarkup()
	if err != nil {
		return err
	}
	resp := telegram.SendPhoto{
		ChatID:      update.Message.Chat.ID,
		Caption:     b.container.GetLocalizer(langTag).LocalizedString("short_description"),
		Photo:       avatarImageURL,
		ReplyMarkup: mainMenuInlineKeyboardMarkup,
	}
	return b.telegramBotService.SendResponse(resp, app.SendPhotoTelegramMethod)
}

func (b *botController) editMessageAndBackToMainMenu(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		return err
	}
	localizer := b.container.GetLocalizer(langTag)
	mainMenuInlineKeyboardMarkup, err := b.keyboardManager.MainMenuInlineKeyboardMarkup()
	if err != nil {
		return err
	}
	photoMedia := telegram.InputPhotoMedia{
		Type:    "photo",
		Media:   avatarImageURL,
		Caption: utils.NewString(localizer.LocalizedString("short_description")),
	}
	editMessageMedia := telegram.EditMessageMedia{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Media:       photoMedia,
		ReplyMarkup: mainMenuInlineKeyboardMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMessageMedia, app.EditMessageMediaTelegramMethod); err != nil {
		return err
	}
	return nil
}

func (b *botController) messageListPayCurrencies(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		return err
	}
	localizer := b.container.GetLocalizer(langTag)
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		log.Error("fail to send a AnswerCallbackQuery to telegram servers", logger.FError(err))
		return err
	}
	payCurrenciesInlineKeyboardMarkup, err := b.keyboardManager.PayCurrenciesInlineKeyboardMarkup()
	if err != nil {
		return err
	}
	resp := telegram.SendPhoto{
		ChatID:      callbackQuery.Message.Chat.ID,
		Caption:     localizer.LocalizedString("select_currency_to_pay"),
		Photo:       topUpImageURL,
		ReplyMarkup: payCurrenciesInlineKeyboardMarkup,
	}
	return b.telegramBotService.SendResponse(resp, app.SendPhotoTelegramMethod)
}

func (b *botController) messageEnterAmountCurrency(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	telegramID := callbackQuery.From.ID
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		return err
	}
	localizer := b.container.GetLocalizer(langTag)
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		log.Error("fail to send a AnswerCallbackQuery to telegram servers", logger.FError(err))
		return err
	}
	profile, err := b.profileRepository.FetchByTelegramID(ctx, telegramID)
	if err != nil {
		log.Error("fail to get profile by telegram ID")
		return err
	}
	currency := b.container.GetConfig().CurrencyByAbbr(*profile.PreferredCurrency)
	if currency == nil {
		log.Error("fail to get currency")
		return app.NilError
	}
	resp := telegram.SendPhoto{
		ChatID: callbackQuery.Message.Chat.ID,
		Caption: localizer.LocalizedStringWithTemplateData("enter_amount_for_payment_in_currency", map[string]any{
			"Currency": currency.Symbol,
		}),
		Photo:       enterAmountImageURL,
		ReplyMarkup: nil,
	}
	if err := b.sessionService.SaveBotStateForUser(ctx, app.EnteringAmountCurrencyBotState, telegramID); err != nil {
		return err
	}
	return b.telegramBotService.SendResponse(resp, app.SendPhotoTelegramMethod)
}

func (b *botController) messageWithPlainText(_ context.Context, text string, update *telegram.Update) error {
	resp := telegram.SendResponse{
		ChatID: update.Message.Chat.ID,
		Text:   text,
	}
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}
