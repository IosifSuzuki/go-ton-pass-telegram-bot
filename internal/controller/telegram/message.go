package telegram

import (
	"context"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
)

func (b *botController) messageToSelectInitialLanguage(ctx context.Context, update *telegram.Update) error {
	log := b.container.GetLogger()
	langTag, _ := b.getLanguageCode(ctx, *update.Message.From)
	localizer := b.container.GetLocalizer(langTag)
	replyMarkup, err := b.keyboardManager.InitialLanguagesKeyboardMarkup()
	if err != nil {
		log.Error("fail to get initial languages keyboard markup", logger.FError(err))
		return err
	}
	sendPhoto := telegram.SendPhoto{
		ChatID:      update.Message.Chat.ID,
		Photo:       selectPreferredLanguageImageURL,
		Caption:     localizer.LocalizedString("select_preferred_language_markdown"),
		ReplyMarkup: replyMarkup,
	}
	return b.telegramBotService.SendResponse(sendPhoto, app.SendPhotoTelegramMethod)
}

func (b *botController) messageToSelectInitialPreferredCurrency(ctx context.Context, chatID int64, user *telegram.User) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, *user)
	if err != nil {
		log.Debug("fail to retrieve language code", logger.FError(err))
		return err
	}
	localizer := b.container.GetLocalizer(langTag)
	replyMarkup, err := b.keyboardManager.InitialPreferredCurrenciesKeyboardMarkup()
	if err != nil {
		log.Error("fail to get initial languages keyboard markup", logger.FError(err))
		return err
	}
	resp := telegram.SendPhoto{
		ChatID:      chatID,
		Photo:       selectPreferredCurrencyImageURL,
		Caption:     localizer.LocalizedString("select_preferred_currency_markdown"),
		ReplyMarkup: replyMarkup,
	}
	log.Debug("prepared message messageToSelectPreferredCurrency")
	return b.telegramBotService.SendResponse(resp, app.SendPhotoTelegramMethod)
}

func (b *botController) messageWelcome(ctx context.Context, chatID int64, user *telegram.User) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, *user)
	if err != nil {
		log.Debug("fail to retrieve language code", logger.FError(err))
		return err
	}
	sendPhotoResp := telegram.SendPhoto{
		ChatID:    chatID,
		Photo:     welcomeImageURL,
		Caption:   b.container.GetLocalizer(langTag).LocalizedString("bot_markdown_description"),
		ParseMode: utils.NewString("MarkdownV2"),
	}
	return b.telegramBotService.SendResponse(sendPhotoResp, app.SendPhotoTelegramMethod)
}

func (b *botController) messageMainMenu(ctx context.Context, update *telegram.Update) error {
	langTag, _ := b.getLanguageCode(ctx, *update.Message.From)
	mainMenuInlineKeyboardMarkup, err := b.keyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		return err
	}
	resp := telegram.SendPhoto{
		ChatID:      update.Message.Chat.ID,
		Caption:     b.container.GetLocalizer(langTag).LocalizedString("short_description_markdown"),
		Photo:       avatarImageURL,
		ParseMode:   utils.NewString("MarkdownV2"),
		ReplyMarkup: mainMenuInlineKeyboardMarkup,
	}
	return b.telegramBotService.SendResponse(resp, app.SendPhotoTelegramMethod)
}

func (b *botController) editMessageMainMenu(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	langTag, _ := b.getLanguageCode(ctx, callbackQuery.From)
	localizer := b.container.GetLocalizer(langTag)
	mainMenuInlineKeyboardMarkup, err := b.keyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		return err
	}
	return b.AnswerCallbackQueryWithEditMessageMedia(
		callbackQuery,
		localizer.LocalizedString("short_description_markdown"),
		avatarImageURL,
		mainMenuInlineKeyboardMarkup,
	)
}

func (b *botController) messageListPayCurrencies(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Debug("fail to retrieve language code", logger.F("langTag", langTag), logger.FError(err))
	}
	localizer := b.container.GetLocalizer(langTag)
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		log.Debug("fail to send a AnswerCallbackQuery to telegram servers", logger.FError(err))
		return err
	}
	payCurrenciesInlineKeyboardMarkup, err := b.keyboardManager.PayCurrenciesKeyboardMarkup()
	if err != nil {
		log.Debug("fail to get pay inline keyboard", logger.FError(err))
		return err
	}
	resp := telegram.SendPhoto{
		ChatID:      callbackQuery.Message.Chat.ID,
		Caption:     localizer.LocalizedString("select_currency_to_pay_markdown"),
		Photo:       topUpImageURL,
		ParseMode:   utils.NewString("MarkdownV2"),
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
		Caption: localizer.LocalizedStringWithTemplateData("enter_amount_for_payment_in_currency_markdown", map[string]any{
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
