package telegram

import (
	"context"
	"errors"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/domain"
	model "go-ton-pass-telegram-bot/internal/model/postpone"
	"go-ton-pass-telegram-bot/internal/model/sms"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
	"strconv"
	"strings"
)

func (b *botController) selectedInitialLanguageCallbackQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
	callbackData *app.TelegramCallbackData,
) error {
	log := b.container.GetLogger()
	telegramID := ctxOptions.Profile.TelegramID
	if callbackData.Parameters == nil {
		log.Error("parameters must contains items")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	parameters := *callbackData.Parameters
	selectedLanguageCode, ok := parameters[0].(string)
	if !ok {
		log.Error("parameters[0] should be a string")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	ctxOptions.TelegramInlineKeyboardManager.Set(selectedLanguageCode)
	if err := b.profileRepository.SetPreferredLanguage(ctx, telegramID, selectedLanguageCode); err != nil {
		log.Error("fail to set preferred language for profile", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	profile, err := b.profileRepository.FetchByTelegramID(ctx, telegramID)
	if err != nil {
		log.Error(
			"fail to get refreshed profile from db",
			logger.FError(err),
			logger.F("telegram_id", telegramID),
		)
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	ctxOptions.Profile = profile
	deleteMessage := telegram.DeleteMessage{
		ChatID:    ctxOptions.Update.CallbackQuery.Message.Chat.ID,
		MessageID: ctxOptions.Update.CallbackQuery.Message.ID,
	}
	if err := b.telegramBotService.SendResponse(deleteMessage, app.DeleteMessageTelegramMethod); err != nil {
		log.Error("fail perform to delete a message", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	if err := b.sendMessageWelcome(ctx, ctxOptions); err != nil {
		log.Error("fail to send a welcome message", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	return b.sendMessageToSelectInitialPreferredCurrency(ctx, ctxOptions)
}

func (b *botController) selectedInitialPreferredCurrencyCallbackQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
	callbackData *app.TelegramCallbackData,
) error {
	log := b.container.GetLogger()
	telegramID := ctxOptions.Profile.TelegramID
	if callbackData.Parameters == nil {
		log.Error("parameters must contains items")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	parameters := *callbackData.Parameters
	selectedPreferredCurrency, ok := parameters[0].(string)
	if !ok {
		log.Error("parameters[0] should be a string")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	if err := b.profileRepository.SetPreferredCurrency(ctx, telegramID, selectedPreferredCurrency); err != nil {
		log.Error("fail to save preferred currency", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	return b.editMessageMainMenu(ctx, ctxOptions)
}

func (b *botController) balanceCallbackQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
) error {
	return b.editMessageProfileBalance(ctx, ctxOptions)
}

func (b *botController) listPayCurrenciesCallbackQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
	_ *app.TelegramCallbackData,
) error {
	return b.editMessageListPayCurrencies(ctx, ctxOptions)
}

func (b *botController) cancelEnteringAmountCallbackQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
) error {
	log := b.container.GetLogger()
	deleteMessage := telegram.DeleteMessage{
		ChatID:    ctxOptions.Update.GetChatID(),
		MessageID: ctxOptions.Update.CallbackQuery.Message.ID,
	}
	if err := b.deleteMessage(&deleteMessage); err != nil {
		log.Error("fail to delete message", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	return nil
}

func (b *botController) selectedPayCurrenciesCallbackQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
	callbackData *app.TelegramCallbackData,
) error {
	log := b.container.GetLogger()
	telegramID := ctxOptions.Profile.TelegramID
	if callbackData.Parameters == nil {
		log.Error("parameters must contains parameters")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	parameters := *callbackData.Parameters
	selectedPayCurrencyAbbr, ok := parameters[0].(string)
	if !ok {
		log.Error("parameters[0] should be a string")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	if err := b.sessionService.SaveString(ctx, service.SelectedPayCurrencyAbbrSessionKey, selectedPayCurrencyAbbr, telegramID); err != nil {
		log.Error("fail to save selected pay currency", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	if err := b.sendMessageEnterAmountCurrency(ctx, ctxOptions); err != nil {
		log.Error("fail to send message with enter amount of currency", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	if err := b.sessionService.SaveBotStateForUser(ctx, app.EnteringAmountCurrencyBotState, telegramID); err != nil {
		log.Error("fail to save the bot state entering amount currency", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	return nil
}

func (b *botController) languagesCallbackQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
	_ *app.TelegramCallbackData,
) error {
	return b.editMessageProfileLanguages(ctx, ctxOptions)
}

func (b *botController) selectedLanguageCallbackQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
	callbackData *app.TelegramCallbackData,
) error {
	log := b.container.GetLogger()
	telegramID := ctxOptions.Profile.TelegramID
	if callbackData.Parameters == nil {
		log.Error("parameters must contains items")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	parameters := *callbackData.Parameters
	selectedLanguageTag, ok := parameters[0].(string)
	if !ok {
		log.Error("parameters[0] should be a string")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	ctxOptions.TelegramInlineKeyboardManager.Set(selectedLanguageTag)
	if err := b.profileRepository.SetPreferredLanguage(ctx, telegramID, selectedLanguageTag); err != nil {
		log.Error(
			"fail to set preferred language for a profile",
			logger.F("preferredLanguage", selectedLanguageTag),
		)
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	return b.editMessageMainMenu(ctx, ctxOptions)
}

func (b *botController) historyCallbackQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
	callbackData *app.TelegramCallbackData,
) error {
	log := b.container.GetLogger()
	profileID := ctxOptions.Profile.ID
	if callbackData.Parameters == nil {
		log.Error("parameters must contains items")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	parameters := *callbackData.Parameters
	currentPage := utils.GetInt64(parameters[0])
	itemsPerPage := utils.GetInt64(parameters[1])
	historySMSCountPointer, err := b.smsHistoryRepository.GetNumberOfRows(ctx, profileID)
	if err != nil {
		log.Error(
			"fail to get number of rows of smsHistory by profile_id",
			logger.F("profile_id", profileID),
			logger.FError(err),
		)
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	if historySMSCountPointer == nil {
		log.Error("historySMSCountPointer has nil value")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	historySMSCount := *historySMSCountPointer
	pagination := app.Pagination{
		CurrentPage:  int(currentPage),
		LenItems:     int(historySMSCount),
		ItemsPerPage: int(itemsPerPage),
	}
	return b.editMessageHistories(ctx, ctxOptions, pagination)
}

func (b *botController) helpCallbackQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
) error {
	return b.editMessageHelp(ctx, ctxOptions)
}

func (b *botController) developingCallbackQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
) error {
	return b.editMessageDevelopingMode(ctx, ctxOptions)
}

func (b *botController) mainMenuCallbackQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
) error {
	return b.editMessageMainMenu(ctx, ctxOptions)
}

func (b *botController) servicesCallbackQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
	callbackData *app.TelegramCallbackData,
) error {
	log := b.container.GetLogger()
	parameters := *callbackData.Parameters
	currentPage := utils.GetInt64(parameters[0])
	itemsPerPage := 16
	smsServices, err := b.smsActivateWorker.GetOrderedServices()
	if err != nil {
		log.Error("fail to get ordered services", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	pagination := app.Pagination{
		CurrentPage:  int(currentPage),
		LenItems:     len(smsServices),
		ItemsPerPage: itemsPerPage,
	}
	return b.editMessageServices(ctx, ctxOptions, pagination, smsServices)
}

func (b *botController) selectServiceCallbackQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
	callbackData *app.TelegramCallbackData,
) error {
	log := b.container.GetLogger()
	parameters := *callbackData.Parameters
	selectedServiceCode, ok := parameters[0].(string)
	if !ok {
		log.Error("parameters[0] should be a string")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	currentPage := utils.GetInt64(parameters[1])
	itemsPerPage := 10
	servicePrices, err := b.smsActivateWorker.GetPriceForService(selectedServiceCode)
	if err != nil {
		log.Error("fail to fetch price for services", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	countries, err := b.smsService.GetCountries()
	if err != nil {
		log.Error("fail to fetch countries", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	pagination := app.Pagination{
		CurrentPage:  int(currentPage),
		ItemsPerPage: itemsPerPage,
		LenItems:     len(servicePrices),
	}
	return b.editMessageServiceCountries(ctx, ctxOptions, pagination, selectedServiceCode, servicePrices, countries)
}

func (b *botController) preferredCurrenciesQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
) error {
	return b.editMessagePreferredCurrencies(ctx, ctxOptions)
}

func (b *botController) payServiceQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
	callbackData *app.TelegramCallbackData,
) error {
	log := b.container.GetLogger()
	telegramID := ctxOptions.Profile.TelegramID
	if callbackData.Parameters == nil {
		log.Error("parameters has nil value", logger.FError(app.NilError))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	parameters := *callbackData.Parameters
	if len(parameters) < 3 {
		log.Error("not enough length parameters")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	serviceCode, ok := parameters[0].(string)
	if !ok {
		log.Error("parameters[0] should be a string")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	countryID := utils.GetInt64(parameters[1])
	maxPrice := utils.GetFloat64(parameters[2])
	priceWithFee := b.exchangeRateWorker.PriceWithFee(maxPrice)
	priceWithFeeUSD, err := b.exchangeRateWorker.ConvertToUSD(priceWithFee, "RUB")
	if err != nil {
		log.Error("fail to convert rubles to usd", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	hasSufficientFunds, err := b.profileRepository.HasSufficientFunds(ctx, telegramID, *priceWithFeeUSD)
	if err != nil {
		log.Error("fail to check that a profile has sufficient funds", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	if !hasSufficientFunds {
		log.Error("hasn't sufficient funds for buy service")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	requestedNumber, err := b.smsService.RequestNumber(serviceCode, countryID, maxPrice)
	smsError, ok := err.(sms.Error)
	if ok && strings.EqualFold(smsError.Name, sms.NoNumbersErrorName) {
		log.Error("no numbers available", logger.FError(smsError))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	} else if ok {
		log.Error("other sms activation error", logger.FError(smsError))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	} else if err != nil {
		log.Error("unhandled error occurred", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	activationID, err := strconv.ParseInt(requestedNumber.ActivationID, 10, 64)
	if err != nil {
		log.Error("convert activation_id to string has failed", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	country, err := b.smsActivateWorker.GetCountry(countryID)
	if err != nil {
		log.Error("fail to get country by id", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	smsService, err := b.smsActivateWorker.GetService(serviceCode)
	if err != nil {
		log.Error("fail to get sms service", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	fullPhoneNumber := utils.PhoneNumberTitle(requestedNumber.PhoneNumber)
	phoneNumber, err := utils.ParsePhoneNumber(fullPhoneNumber)
	if err != nil {
		log.Error("fail to parse phone number", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	domainSMSHistory := domain.SMSHistory{
		ProfileID:        ctxOptions.Profile.ID,
		ActivationID:     activationID,
		Status:           string(app.PendingSMSActivateState),
		ServiceCode:      smsService.Code,
		ServiceName:      smsService.Name,
		CountryID:        country.ID,
		CountryName:      country.Title,
		PhoneCodeNumber:  phoneNumber.CountryCode,
		PhoneShortNumber: phoneNumber.ShortPhoneNumber,
	}
	smsHistoryID, err := b.smsHistoryRepository.Create(ctx, &domainSMSHistory)
	if err != nil {
		log.Error("fail to create sms history", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	if err := b.profileRepository.Debit(ctx, telegramID, *priceWithFeeUSD); err != nil {
		log.Error("fail to withdraw money from account")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	workflow, err := b.postponeService.ScheduleCheckSMSActivation(ctx, telegramID, activationID, *priceWithFeeUSD)
	if err != nil {
		log.Error("fail to prepare schedule to check the sms activation", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	temporalWorkflowDomain := domain.TemporalWorkflow{
		SMSHistoryID:  *smsHistoryID,
		TemporalID:    workflow.ID,
		TemporalRunID: workflow.RunID,
	}
	_, err = b.temporalWorkflowRepository.Create(ctx, &temporalWorkflowDomain)
	if err != nil {
		log.Error("fail to record temporal workflow to db", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	if err := b.sendMessageStartSMSActivation(ctx, ctxOptions, &domainSMSHistory, *smsHistoryID); err != nil {
		log.Error("fail to send message with sms activation", logger.FError(err))
		return nil
	}
	return b.sendMessageMainMenu(ctx, ctxOptions)
}

func (b *botController) emptyQueryCommandHandler(_ context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	if err := b.AnswerCallbackQuery(callbackQuery, nil, false); err != nil {
		log.Error("fail to answer callback query", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) selectPreferredCurrencyQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
	callbackData *app.TelegramCallbackData,
) error {
	log := b.container.GetLogger()
	telegramID := ctxOptions.Update.GetTelegramID()
	if callbackData.Parameters == nil {
		log.Error("parameters has nil value", logger.FError(app.NilError))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	parameters := *callbackData.Parameters
	selectedPreferredCurrency, ok := parameters[0].(string)
	if !ok {
		log.Error("parameters[0] should be a string")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	if err := b.profileRepository.SetPreferredCurrency(ctx, telegramID, selectedPreferredCurrency); err != nil {
		log.Error("fail to set preferred currency to profile", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	return b.editMessageMainMenu(ctx, ctxOptions)
}

func (b *botController) deleteCryptoBotQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
	callbackData *app.TelegramCallbackData,
) error {
	log := b.container.GetLogger()
	if callbackData.Parameters == nil {
		log.Error("parameters has nil value", logger.FError(app.NilError))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	parameters := *callbackData.Parameters
	invoiceID := utils.GetInt64(parameters[0])
	removeInvoiceErr := b.cryptoPayBot.RemoveInvoice(invoiceID)
	if removeInvoiceErr != nil && errors.Is(removeInvoiceErr, app.DeleteInvoiceError) {
		log.Error("fail to remove invoice", logger.FError(removeInvoiceErr))
	} else if removeInvoiceErr != nil {
		log.Error("fail to remove invoice", logger.FError(removeInvoiceErr))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	deleteMessage := telegram.DeleteMessage{
		ChatID:    ctxOptions.Update.GetChatID(),
		MessageID: ctxOptions.Update.CallbackQuery.Message.ID,
	}
	if err := b.deleteMessage(&deleteMessage); err != nil {
		log.Error("fail to delete message in bot chat", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	if removeInvoiceErr != nil && errors.Is(removeInvoiceErr, app.DeleteInvoiceError) {
		log.Error("fail to remove invoice", logger.FError(removeInvoiceErr))
		return nil
	} else if removeInvoiceErr != nil {
		log.Error("fail to remove invoice", logger.FError(removeInvoiceErr))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	if err := b.sendMessageSuccessfullyDeletedInvoice(ctx, ctxOptions); err != nil {
		log.Error(
			"fail to send message with notice about delete the invoice",
			logger.FError(err),
		)
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	return b.sendMessageMainMenu(ctx, ctxOptions)
}

func (b *botController) confirmServiceQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
	callbackData *app.TelegramCallbackData,
) error {
	log := b.container.GetLogger()
	if callbackData.Parameters == nil {
		log.Error("parameters has nil value", logger.FError(app.NilError))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	parameters := *callbackData.Parameters
	serviceCode, ok := parameters[0].(string)
	if !ok {
		log.Error("parameters[0] should be a string")
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	countryID := utils.GetInt64(parameters[1])
	priceInRub := utils.GetFloat64(parameters[2])
	priceWithFeeInPreferredCurrency := utils.GetFloat64(parameters[3])
	country, err := b.smsActivateWorker.GetCountry(countryID)
	if err != nil {
		log.Error("fail to get country", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	selectedService, err := b.smsActivateWorker.GetService(serviceCode)
	if err != nil {
		log.Error("fail to get service", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	return b.editMessageConfirmService(
		ctx,
		ctxOptions,
		selectedService,
		country,
		priceInRub,
		priceWithFeeInPreferredCurrency,
	)
}

func (b *botController) refundAmountFromSMSActivationQueryCommandHandler(
	ctx context.Context,
	ctxOptions *ContextOptions,
	callbackData *app.TelegramCallbackData,
) error {
	log := b.container.GetLogger()
	if callbackData.Parameters == nil {
		log.Error("parameters has nil value", logger.FError(app.NilError))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	parameters := *callbackData.Parameters
	smsHistoryID := utils.GetInt64(parameters[0])
	temporalWorkflowDomain, err := b.temporalWorkflowRepository.GetBySMSHistoryID(ctx, smsHistoryID)
	if err != nil {
		log.Error("fetch workflow by sms history id has failed",
			logger.F("sms_history_id", smsHistoryID),
			logger.FError(err),
		)
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	workflow := model.Workflow{
		ID:    temporalWorkflowDomain.TemporalID,
		RunID: temporalWorkflowDomain.TemporalRunID,
	}
	if err := b.postponeService.CancelSMSActivation(ctx, workflow); err != nil {
		log.Error("fail to cancel sms activation", logger.FError(err))
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	return b.AnswerCallbackQuery(ctxOptions.Update.CallbackQuery, nil, false)
}
