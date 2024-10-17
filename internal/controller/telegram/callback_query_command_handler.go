package telegram

import (
	"context"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/domain"
	"go-ton-pass-telegram-bot/internal/model/sms"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
	"strconv"
	"strings"
)

func (b *botController) balanceCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	telegramID := callbackQuery.From.ID
	log.Debug("execute balanceCallbackQueryCommandHandler", logger.F("callbackQuery", callbackQuery))
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
	preferredCurrency := telegramProfile.PreferredCurrency
	if preferredCurrency == nil {
		log.Error("preferredCurrency is missing", logger.FError(err))
		return app.NilError
	}
	convertedBalance, err := b.exchangeRateWorker.ConvertFromUSD(telegramProfile.Balance, *preferredCurrency)
	if err != nil {
		log.Error("convert currency failed", logger.F("from", "usd"), logger.F("to", *preferredCurrency))
		return err
	}
	balanceText := b.container.GetLocalizer(*langTag).LocalizedStringWithTemplateData("your_balance_is", map[string]any{
		"Balance": utils.CurrencyAmountTextFormat(*convertedBalance, *currency),
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
	photoMedia := telegram.InputPhotoMedia{
		Type:    "photo",
		Media:   avatarImageURL,
		Caption: &balanceText,
	}
	editMessageMedia := telegram.EditMessageMedia{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Media:       photoMedia,
		ReplyMarkup: replyKeyboardMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMessageMedia, app.EditMessageMediaTelegramMethod); err != nil {
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
	photoMedia := telegram.InputPhotoMedia{
		Type:  "photo",
		Media: selectPreferredLanguageImageURL,
		Caption: utils.NewString(b.container.GetLocalizer(*languageCode).LocalizedStringWithTemplateData("choose_language", map[string]any{
			"Language": utils.LanguageTextFormat(*selectedLanguage),
		})),
	}
	editMessage := telegram.EditMessageMedia{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Media:       photoMedia,
		ReplyMarkup: keyboardMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMessage, app.EditMessageMediaTelegramMethod); err != nil {
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
	if err := b.editMessageAndBackToMainMenu(ctx, callbackQuery); err != nil {
		log.Error("fail to send message with main menu to telegram servers")
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
	localizer := b.container.GetLocalizer(*langTag)
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		return err
	}
	photoMedia := telegram.InputPhotoMedia{
		Type:    "photo",
		Media:   historyImageURL,
		Caption: utils.NewString(localizer.LocalizedString("empty_history")),
	}
	backToMenuInlineKeyboardMarkup, err := b.getMenuInlineKeyboardMarkup(*langTag)
	if err != nil {
		return err
	}
	editMessage := telegram.EditMessageMedia{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Media:       photoMedia,
		ReplyMarkup: backToMenuInlineKeyboardMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMessage, app.EditMessageMediaTelegramMethod); err != nil {
		log.Error("fail to send message to telegram servers", logger.FError(err))
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
	localizer := b.container.GetLocalizer(*langTag)
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
	photoMedia := telegram.InputPhotoMedia{
		Type:    "photo",
		Media:   helpImageURL,
		Caption: utils.NewString(localizer.LocalizedString("help_cmd_text")),
	}
	editMessage := telegram.EditMessageMedia{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Media:       photoMedia,
		ReplyMarkup: backToMenuInlineKeyboardMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMessage, app.EditMessageMediaTelegramMethod); err != nil {
		log.Error("fail to send a EditMessage to telegram servers", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) developingCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
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
	if err := b.editMessageAndBackToMainMenu(ctx, callbackQuery); err != nil {
		log.Error("fail to send main menu to telegram servers", logger.FError(err))
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
	localizer := b.container.GetLocalizer(*langTag)
	telegramCallbackData, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if telegramCallbackData.Parameters == nil {
		return err
	}
	parameters := *telegramCallbackData.Parameters
	currentPage := utils.GetInt64(parameters[0])
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		log.Error("fail to send a AnswerCallbackQuery to telegram servers", logger.FError(err))
		return err
	}
	smsServices, err := b.smsActivateWorker.GetOrderedServices()
	if err != nil {
		return err
	}
	pagination := app.Pagination[sms.Service]{
		CurrentPage:  int(currentPage),
		ItemsPerPage: MaxInlineKeyboardRows * 2,
		DataSource:   smsServices,
	}
	replyMarkup, err := b.getServicesInlineKeyboardMarkup(ctx, callbackQuery, &pagination)
	if err != nil {
		return err
	}
	photoMedia := telegram.InputPhotoMedia{
		Type:    "photo",
		Media:   chooseServiceImageURL,
		Caption: utils.NewString(localizer.LocalizedString("select_sms_service")),
	}
	editMessageMedia := telegram.EditMessageMedia{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Media:       photoMedia,
		ReplyMarkup: replyMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMessageMedia, app.EditMessageMediaTelegramMethod); err != nil {
		log.Error("fail to send a EditMessage to telegram servers", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) selectServiceCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	telegramID := callbackQuery.From.ID
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	localizer := b.container.GetLocalizer(*langTag)
	telegramCallbackData, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if telegramCallbackData.Parameters == nil {
		return err
	}
	parameters := *telegramCallbackData.Parameters
	selectedService := parameters[0].(string)
	currentPage := utils.GetInt64(parameters[1])
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
		log.Error("fail to fetch profile by telegram id", logger.F("telegram_id", profile.TelegramID))
		return err
	}
	servicePrices, err := b.smsActivateWorker.GetPriceForService(selectedService)
	if err != nil {
		log.Error("fail to GetPriceForService", logger.FError(err))
		return err
	}
	countries, err := b.smsService.GetCountries()
	if err != nil {
		log.Error("fail to GetCountries", logger.FError(err))
		return err
	}
	pagination := app.Pagination[sms.PriceForService]{
		CurrentPage:  int(currentPage),
		ItemsPerPage: MaxInlineKeyboardRows,
		DataSource:   servicePrices,
	}

	replyMarkup, err := b.getServiceWithCountryInlineKeyboardMarkup(*langTag, *profile.PreferredCurrency, selectedService, &pagination, countries)
	if err != nil {
		log.Error("fail to getServiceWithCountryInlineKeyboardMarkup", logger.FError(err))
		return err
	}
	photoMedia := telegram.InputPhotoMedia{
		Type:    "photo",
		Media:   chooseCountryImageURL,
		Caption: utils.NewString(localizer.LocalizedString("select_sms_service_with_country")),
	}
	editMediaMessage := telegram.EditMessageMedia{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Media:       photoMedia,
		ReplyMarkup: replyMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMediaMessage, app.EditMessageMediaTelegramMethod); err != nil {
		log.Error("fail to send a EditMessage to telegram servers", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) preferredCurrenciesQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	telegramID := callbackQuery.From.ID
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Error("fail to get language code", logger.FError(err))
		return err
	}
	telegramProfile, err := b.profileRepository.FetchByTelegramID(ctx, telegramID)
	if err != nil {
		log.Error("fail to get profile by telegram id", logger.FError(err))
		return err
	}
	preferredCurrency := telegramProfile.PreferredCurrency
	if preferredCurrency == nil {
		log.Error("preferredCurrency is missing")
		return app.NilError
	}
	currency := b.container.GetConfig().CurrencyByAbbr(*preferredCurrency)
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
	inlineKeyboardMarkup, err := b.getPreferredCurrenciesKeyboardMarkup(*langTag)
	if err != nil {
		log.Error("fail to create a currencies keyboard markup", logger.FError(err))
		return err
	}
	photoMedia := telegram.InputPhotoMedia{
		Type:  "photo",
		Media: selectPreferredLanguageImageURL,
		Caption: utils.NewString(localizer.LocalizedStringWithTemplateData("choose_preferred_currency", map[string]any{
			"Currency": currency.ABBR,
		})),
	}
	editMessage := telegram.EditMessageMedia{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Media:       photoMedia,
		ReplyMarkup: inlineKeyboardMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMessage, app.EditMessageMediaTelegramMethod); err != nil {
		log.Error("fail to send a EditMessage to telegram servers", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) selectSMSServiceWithCountryQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	telegramID := callbackQuery.From.ID
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Error("fail to get language code", logger.FError(err))
		return err
	}
	localizer := b.container.GetLocalizer(*langTag)
	telegramCallbackData, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if err != nil {
		return err
	}
	parameters := *telegramCallbackData.Parameters
	serviceCode := parameters[0].(string)
	countryID := utils.GetInt64(parameters[1])
	maxPrice := utils.GetFloat64(parameters[2])
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		log.Error("fail to send a AnswerCallbackQuery to telegram servers", logger.FError(err))
		return err
	}
	domainProfile, err := b.profileRepository.FetchByTelegramID(ctx, telegramID)
	if err != nil {
		log.Error("fail to get profile", logger.FError(err))
		return err
	}
	requestedNumber, err := b.smsService.RequestNumber(serviceCode, countryID, maxPrice)
	smsError, ok := err.(sms.Error)
	if ok && strings.EqualFold(smsError.Name, sms.NoNumbersErrorName) {
		return b.EditMessageMedia(ctx, callbackQuery, localizer.LocalizedString("numbers_are_unavailable"), failReceivedCodeImageURL)
	} else if ok {
		return b.EditMessageMedia(ctx, callbackQuery, localizer.LocalizedString("fail_to_order_sms_code"), failReceivedCodeImageURL)
	} else if err != nil {
		log.Debug("unhandled error occurred", logger.FError(err))
		return b.EditMessageMedia(ctx, callbackQuery, localizer.LocalizedString("internal_error"), failReceivedCodeImageURL)
	}
	activationID, err := strconv.ParseInt(requestedNumber.ActivationID, 10, 64)
	if err != nil {
		log.Error("parse response ActivationID to int64 has failed", logger.FError(err))
		return err
	}
	domainSMSHistory := domain.SMSHistory{
		ProfileID:    domainProfile.ID,
		ActivationID: activationID,
		ServiceCode:  serviceCode,
		PhoneNumber:  requestedNumber.PhoneNumber,
	}
	if _, err := b.smsHistoryRepository.Create(ctx, &domainSMSHistory); err != nil {
		log.Error("fail to create sms history", logger.FError(err))
		return err
	}
	formattedText := localizer.LocalizedStringWithTemplateData("start_registration_form_with_sms_code", map[string]any{
		"PhoneNumber": requestedNumber.PhoneNumber,
	})
	return b.EditMessageMedia(ctx, callbackQuery, formattedText, avatarImageURL)
}

func (b *botController) emptyQueryCommandHandler(_ context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		log.Error("fail to send a AnswerCallbackQuery to telegram servers", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) selectPreferredCurrencyQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	telegramID := callbackQuery.From.ID
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		log.Error("fail to send a AnswerCallbackQuery to telegram servers", logger.FError(err))
		return err
	}
	telegramCallbackData, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if err != nil {
		return err
	}
	parameters := *telegramCallbackData.Parameters
	selectedPreferredCurrency := parameters[0].(string)
	if err := b.profileRepository.SetPreferredCurrency(ctx, telegramID, selectedPreferredCurrency); err != nil {
		log.Error("fail to set preferred currency to profile", logger.FError(err))
		return err
	}
	if err := b.editMessageAndBackToMainMenu(ctx, callbackQuery); err != nil {
		log.Error("fail to send main menu to telegram servers", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) EditMessageMedia(ctx context.Context, callbackQuery *telegram.CallbackQuery, text string, photoURL string) error {
	photoMedia := telegram.InputPhotoMedia{
		Type:      "photo",
		Media:     photoURL,
		Caption:   utils.NewString(text),
		ParseMode: utils.NewString("MarkdownV2"),
	}
	mainMenuInlineKeyboardMarkup, err := b.getMainMenuInlineKeyboardMarkup(ctx, callbackQuery.From)
	if err != nil {
		return err
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
