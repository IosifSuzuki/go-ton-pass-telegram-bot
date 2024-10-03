package telegram

import (
	"context"
	"fmt"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
)

func (b *botController) balanceCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	telegramID := callbackQuery.From.ID
	telegramProfile, err := b.profileRepository.FetchByTelegramID(ctx, telegramID)
	if err != nil {
		log.Error("fail to fetchByTelegramID", logger.FError(err))
		return err
	}
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	localizer := b.container.GetLocalizer(*langTag)
	if err != nil {
		log.Error("fail to getLanguageCode", logger.FError(err))
		return err
	}
	currency := b.container.GetConfig().CurrencyByAbbr(*telegramProfile.PreferredCurrency)
	log.Debug("execute balanceCallbackQueryCommandHandler", logger.F("callbackQuery", callbackQuery))
	balanceText := b.container.GetLocalizer(*langTag).LocalizedStringWithTemplateData("your_balance_is", map[string]any{
		"Balance":  telegramProfile.Balance,
		"Currency": currency.Symbol,
	})
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		log.Error("fail to send a AnswerCallbackQuery to telegram servers", logger.FError(err))
		return err
	}
	listPayCurrenciesTelegramCallbackData := app.TelegramCallbackData{
		Name:       app.ListPayCurrenciesCallbackQueryCmdText,
		Parameters: nil,
	}
	listPayCurrenciesData, err := utils.EncodeTelegramCallbackData(listPayCurrenciesTelegramCallbackData)
	if err != nil {
		return err
	}
	listPayCurrenciesKeyboardButton := telegram.InlineKeyboardButton{
		Text: localizer.LocalizedString("top_up_balance"),
		Data: listPayCurrenciesData,
	}
	replyKeyboardMarkup, err := b.getInlineKeyboardMarkupWithMainMenuButton(
		*langTag,
		[]telegram.InlineKeyboardButton{listPayCurrenciesKeyboardButton},
		1,
	)
	editMessage := telegram.EditMessage{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Text:        balanceText,
		ReplyMarkup: replyKeyboardMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMessage, app.EditMessageTextTelegramMethod); err != nil {
		log.Error("fail to send a EditMessage to telegram servers", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) listPayCurrenciesCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		return err
	}
	if err := b.messageListPayCurrencies(ctx, callbackQuery); err != nil {
		log.Error("fail to send message select pay currency", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) selectedPayCurrenciesCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		return err
	}
	telegramID := callbackQuery.From.ID
	if err := b.sessionService.SaveBotStateForUser(ctx, app.EnterAmountCurrencyBotState, telegramID); err != nil {
		log.Error("fail to save bot state", logger.FError(err))
		return err
	}
	return b.enterAmountCurrencyBotStageHandler(ctx, callbackQuery)
}

func (b *botController) languagesCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	telegramID := callbackQuery.From.ID
	telegramProfile, err := b.profileRepository.FetchByTelegramID(ctx, telegramID)
	languageCode := telegramProfile.PreferredLanguage
	selectedLanguage := b.container.GetConfig().LanguageByCode(*languageCode)
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		return err
	}
	keyboardMarkup, err := b.getLanguagesInlineKeyboardMarkup(ctx, callbackQuery.From)
	if err != nil {
		log.Error("fail to get a keyboardMarkup", logger.FError(err))
		return err
	}
	editMessage := telegram.EditMessage{
		ChatID:    &callbackQuery.Message.Chat.ID,
		MessageID: &callbackQuery.Message.ID,
		Text: b.container.GetLocalizer(*languageCode).LocalizedStringWithTemplateData("choose_language", map[string]any{
			"Language": fmt.Sprintf("%s %s", selectedLanguage.FlagEmoji, selectedLanguage.NativeName),
		}),
		ReplyMarkup: keyboardMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMessage, app.EditMessageTextTelegramMethod); err != nil {
		log.Error("fail to send a EditMessage to telegram servers", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) selectLanguageCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	telegramID := callbackQuery.From.ID
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		return err
	}
	telegramCallbackData, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if err != nil {
		return err
	}
	parameters := *telegramCallbackData.Parameters
	selectedLanguageCode := parameters[0].(string)
	if err := b.profileRepository.SetPreferredLanguage(ctx, telegramID, selectedLanguageCode); err != nil {
		log.Error("fail to SetPreferredLanguage", logger.F("preferredLanguage", selectedLanguageCode))
		return err
	}
	mainMenuInlineKeyboardMarkup, err := b.getMainMenuInlineKeyboardMarkup(ctx, callbackQuery.From)
	if err != nil {
		return err
	}
	editMessage := telegram.EditMessage{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Text:        b.container.GetLocalizer(selectedLanguageCode).LocalizedString("short_description"),
		ReplyMarkup: mainMenuInlineKeyboardMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMessage, app.EditMessageTextTelegramMethod); err != nil {
		log.Error("fail to send a EditMessage to telegram servers", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) historyCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Error("fail to getLanguageCode", logger.FError(err))
		return err
	}
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		return err
	}
	backToMenuInlineKeyboardMarkup, err := b.getMenuInlineKeyboardMarkup(*langTag)
	if err != nil {
		return err
	}
	editMessage := telegram.EditMessage{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Text:        b.container.GetLocalizer(*langTag).LocalizedString("empty_history"),
		ReplyMarkup: backToMenuInlineKeyboardMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMessage, app.EditMessageTextTelegramMethod); err != nil {
		log.Error("fail to send a EditMessage to telegram servers", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) helpCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Error("fail to getLanguageCode", logger.FError(err))
		return err
	}
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		return err
	}
	backToMenuInlineKeyboardMarkup, err := b.getMenuInlineKeyboardMarkup(*langTag)
	if err != nil {
		return err
	}
	editMessage := telegram.EditMessage{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Text:        b.container.GetLocalizer(*langTag).LocalizedString("help_cmd_text"),
		ReplyMarkup: backToMenuInlineKeyboardMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMessage, app.EditMessageTextTelegramMethod); err != nil {
		log.Error("fail to send a EditMessage to telegram servers", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) unsupportedCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Error("fail to getLanguageCode", logger.FError(err))
		return err
	}
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      utils.NewString(b.container.GetLocalizer(*langTag).LocalizedString("development_process")),
		ShowAlert: true,
	}
	return b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod)
}

func (b *botController) mainMenuCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	telegramID := callbackQuery.From.ID
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Error("fail to getLanguageCode", logger.FError(err))
		return err
	}
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		log.Error("fail to send a AnswerCallbackQuery to telegram servers", logger.FError(err))
		return err
	}
	if err := b.sessionService.ClearBotStateForUser(ctx, telegramID); err != nil {
		log.Error("fail to clear bot state for user", logger.FError(err), logger.F("telegramID", telegramID))
		return err
	}
	mainMenuInlineKeyboardMarkup, err := b.getMainMenuInlineKeyboardMarkup(ctx, callbackQuery.From)
	if err != nil {
		return err
	}
	editMessage := telegram.EditMessage{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Text:        b.container.GetLocalizer(*langTag).LocalizedString("short_description"),
		ReplyMarkup: mainMenuInlineKeyboardMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMessage, app.EditMessageTextTelegramMethod); err != nil {
		log.Error("fail to send a EditMessage to telegram servers", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) servicesCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Error("fail to getLanguageCode", logger.FError(err))
		return err
	}
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		log.Error("fail to send a AnswerCallbackQuery to telegram servers", logger.FError(err))
		return err
	}
	smsServices, err := b.smsService.GetServices()
	if err != nil {
		return err
	}
	replyMarkup, err := b.getServicesInlineKeyboardMarkup(ctx, callbackQuery, smsServices)
	if err != nil {
		return err
	}
	editMessage := telegram.EditMessage{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Text:        b.container.GetLocalizer(*langTag).LocalizedString("select_sms_service_with_country"),
		ReplyMarkup: replyMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMessage, app.EditMessageTextTelegramMethod); err != nil {
		log.Error("fail to send a EditMessage to telegram servers", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) selectServiceCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	telegramCallbackData, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if telegramCallbackData.Parameters == nil {
		return err
	}
	parameters := *telegramCallbackData.Parameters
	selectedService := parameters[0].(string)
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		log.Error("fail to send a AnswerCallbackQuery to telegram servers", logger.FError(err))
		return err
	}
	servicePrices, err := b.smsService.GetPriceForService(selectedService)
	if err != nil {
		log.Error("fail to GetPriceForService", logger.FError(err))
		return err
	}
	countries, err := b.smsService.GetCountries()
	if err != nil {
		log.Error("fail to GetCountries", logger.FError(err))
		return err
	}
	replyMarkup, err := b.getServicePricesInlineKeyboardMarkup(*langTag, servicePrices, countries)
	if err != nil {
		log.Error("fail to getServicePricesInlineKeyboardMarkup", logger.FError(err))
		return err
	}
	editMessage := telegram.EditMessage{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Text:        b.container.GetLocalizer(*langTag).LocalizedString("select_sms_service_with_country"),
		ReplyMarkup: replyMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMessage, app.EditMessageTextTelegramMethod); err != nil {
		log.Error("fail to send a EditMessage to telegram servers", logger.FError(err))
		return err
	}
	return nil
}
