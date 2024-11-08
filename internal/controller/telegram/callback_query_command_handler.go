package telegram

import (
	"context"
	"errors"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/domain"
	"go-ton-pass-telegram-bot/internal/model/sms"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
	"strconv"
	"strings"
)

func (b *botController) selectedInitialLanguageCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Debug("fail to retrieve language code", logger.F("langTag", langTag))
	}
	telegramID := callbackQuery.From.ID
	localizer := b.container.GetLocalizer(langTag)
	telegramCallbackQueryData, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if err != nil {
		log.Error("fail to decode telegram callback data", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			nil,
		)
	}
	if telegramCallbackQueryData.Parameters == nil {
		log.Error("parameters must contains parameters")
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			nil,
		)
	}
	parameters := *telegramCallbackQueryData.Parameters
	selectedLanguageCode, ok := parameters[0].(string)
	if !ok {
		log.Error("first parameter should be string", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			nil,
		)
	}
	localizer = b.container.GetLocalizer(selectedLanguageCode)
	b.keyboardManager.Set(selectedLanguageCode)
	if err := b.profileRepository.SetPreferredLanguage(ctx, telegramID, selectedLanguageCode); err != nil {
		log.Error("fail to set initial preferred language", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			nil,
		)
	}
	deleteMessage := telegram.DeleteMessage{
		ChatID:    callbackQuery.Message.Chat.ID,
		MessageID: callbackQuery.Message.ID,
	}
	if err := b.telegramBotService.SendResponse(deleteMessage, app.DeleteMessageTelegramMethod); err != nil {
		log.Error("fail perform delete message", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			nil,
		)
	}
	if err := b.messageWelcome(ctx, callbackQuery.Message.Chat.ID, &callbackQuery.From); err != nil {
		log.Error("fail to send welcome message", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			nil,
		)
	}
	if err := b.messageToSelectInitialPreferredCurrency(ctx, callbackQuery.Message.Chat.ID, &callbackQuery.From); err != nil {
		log.Error("fail to select preferred currency", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			nil,
		)
	}
	return nil
}

func (b *botController) selectedInitialPreferredCurrencyCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	telegramID := callbackQuery.From.ID
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Debug("fail to retrieve language code", logger.F("langTag", langTag))
	}
	localizer := b.container.GetLocalizer(langTag)
	telegramCallbackQueryData, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if err != nil {
		log.Error("fail to decode telegram callback data", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			nil,
		)
	}
	if telegramCallbackQueryData.Parameters == nil {
		log.Error("parameters must contains parameters")
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			nil,
		)
	}
	parameters := *telegramCallbackQueryData.Parameters
	selectedPreferredCurrency, ok := parameters[0].(string)
	if !ok {
		log.Error("first parameter should be string", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			nil,
		)
	}
	if err := b.profileRepository.SetPreferredCurrency(ctx, telegramID, selectedPreferredCurrency); err != nil {
		log.Error("fail to save preferred currency", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			nil,
		)
	}
	if err := b.editMessageMainMenu(ctx, callbackQuery); err != nil {
		log.Error("fail to edit message on main menu", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) balanceCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	telegramID := callbackQuery.From.ID
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Debug("fail to retrieve language code", logger.F("langTag", langTag))
	}
	replyMarkup, err := b.keyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Error("fail to get main menu inline keyboard", logger.FError(err))
		return err
	}
	localizer := b.container.GetLocalizer(langTag)
	telegramProfile, err := b.profileRepository.FetchByTelegramID(ctx, telegramID)
	if err != nil {
		log.Error("fail to fetchByTelegramID", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	currency := b.container.GetConfig().CurrencyByAbbr(*telegramProfile.PreferredCurrency)
	preferredCurrency := telegramProfile.PreferredCurrency
	if preferredCurrency == nil {
		log.Error("preferredCurrency is missing", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	convertedBalance, err := b.exchangeRateWorker.ConvertFromUSD(telegramProfile.Balance, *preferredCurrency)
	if err != nil {
		log.Error("convert currency failed", logger.F("from", "usd"), logger.F("to", *preferredCurrency))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	balanceText := localizer.LocalizedStringWithTemplateData("your_balance_is_markdown", map[string]any{
		"Balance": utils.EscapeMarkdownText(utils.CurrencyAmountTextFormat(*convertedBalance, *currency)),
	})
	replyMarkup, err = b.keyboardManager.TopUpBalanceKeyboardMarkup()
	if err != nil {
		log.Error("fail to get inline keyboard", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	return b.AnswerCallbackQueryWithEditMessageMedia(
		callbackQuery,
		balanceText,
		avatarImageURL,
		replyMarkup,
	)
}

func (b *botController) listPayCurrenciesCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	localizer := b.container.GetLocalizer(langTag)
	if err != nil {
		log.Debug("fail to retrieve language code", logger.F("langTag", langTag))
	}
	replyMarkup, err := b.keyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Error("fail to get main menu inline keyboard", logger.FError(err))
		return err
	}
	if err := b.messageListPayCurrencies(ctx, callbackQuery); err != nil {
		log.Error("fail to send message select pay currency", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	return nil
}

func (b *botController) selectedPayCurrenciesCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Debug("fail to retrieve language code", logger.F("langTag", langTag))
	}
	localizer := b.container.GetLocalizer(langTag)
	replyMarkup, err := b.keyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Error("fail to get main menu inline keyboard", logger.FError(err))
		return err
	}
	telegramID := callbackQuery.From.ID
	telegramCallbackData, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if err != nil {
		log.Error("fail to decode telegram callback data", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	parameters := *telegramCallbackData.Parameters
	selectedPayCurrencyAbbr, ok := parameters[0].(string)
	if !ok {
		log.Error("parameter must contains selected pay currency")
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	if err := b.sessionService.SaveString(ctx, service.SelectedPayCurrencyAbbrKey, selectedPayCurrencyAbbr, telegramID); err != nil {
		log.Error("fail to save selected pay currency", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	if err := b.messageEnterAmountCurrency(ctx, callbackQuery); err != nil {
		log.Error("fail to send message with enter amount of currency", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	if err := b.sessionService.SaveBotStateForUser(ctx, app.EnteringAmountCurrencyBotState, telegramID); err != nil {
		log.Error("fail to save bot state entering amount currency", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) languagesCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Debug("fail to retrieve language code", logger.F("langTag", langTag))
	}
	localizer := b.container.GetLocalizer(langTag)
	replyMarkup, err := b.keyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Error("fail to get main menu inline keyboard", logger.FError(err))
		return err
	}
	language := b.container.GetConfig().LanguageByCode(langTag)
	return b.AnswerCallbackQueryWithEditMessageMedia(
		callbackQuery,
		localizer.LocalizedStringWithTemplateData("choose_language_markdown", map[string]any{
			"Language": utils.EscapeMarkdownText(utils.LanguageTextFormat(*language)),
		}),
		selectPreferredLanguageImageURL,
		replyMarkup,
	)
}

func (b *botController) selectLanguageCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Debug("fail to retrieve language code", logger.F("langTag", langTag))
	}
	localizer := b.container.GetLocalizer(langTag)
	telegramID := callbackQuery.From.ID
	replyMarkup, err := b.keyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Error("fail to get main menu inline keyboard", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	telegramCallbackData, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if err != nil {
		log.Error("fail to decode telegram callback data", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	parameters := *telegramCallbackData.Parameters
	langTag, ok := parameters[0].(string)
	if !ok {
		log.Error("fail to get selected language from telegram callback data", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	localizer = b.container.GetLocalizer(langTag)
	b.keyboardManager.Set(langTag)
	replyMarkup, err = b.keyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Error("fail to get main menu inline keyboard", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	if err := b.profileRepository.SetPreferredLanguage(ctx, telegramID, langTag); err != nil {
		log.Error("fail to SetPreferredLanguage", logger.F("preferredLanguage", langTag))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	if err := b.editMessageMainMenu(ctx, callbackQuery); err != nil {
		log.Error("fail to send message with main menu to telegram servers", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	return nil
}

func (b *botController) historyCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Debug("fail to getLanguageCode", logger.F("langTag", langTag), logger.FError(err))
	}
	telegramID := callbackQuery.From.ID
	localizer := b.container.GetLocalizer(langTag)
	replyMarkup, err := b.keyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Error("fail to get main menu inline keyboard", logger.FError(err))
		return err
	}
	telegramCallbackQueryData, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if err != nil {
		log.Error("fail to decode telegram callback data", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	if telegramCallbackQueryData.Parameters == nil {
		log.Error("parameters must contains parameters")
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	parameters := *telegramCallbackQueryData.Parameters
	currentPage := utils.GetInt64(parameters[0])
	itemsPerPage := utils.GetInt64(parameters[1])
	profile, err := b.profileRepository.FetchByTelegramID(ctx, telegramID)
	if err != nil {
		log.Error("fail to fetch profile by telegram id", logger.F("telegram_id", telegramID), logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	historySMSCountPointer, err := b.smsHistoryRepository.GetNumberOfRows(ctx, profile.ID)
	if err != nil {
		log.Error("fail to get number of rows of smsHistory by prifile id", logger.F("profile_id", profile.ID), logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	historySMSCount := *historySMSCountPointer
	if historySMSCount == 0 {
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("empty_history_markdown"),
			historyImageURL,
			replyMarkup,
		)
	}
	pagination := app.Pagination{
		CurrentPage:  int(currentPage),
		LenItems:     int(historySMSCount),
		ItemsPerPage: int(itemsPerPage),
	}
	offset := pagination.CurrentPage * pagination.ItemsPerPage
	smsHistories, err := b.smsHistoryRepository.FetchList(ctx, profile.ID, offset, pagination.ItemsPerPage)
	if err != nil {
		log.Error("fail to get list of sms history",
			logger.F("profile_id", profile.ID),
			logger.F("offset", offset),
			logger.F("itemsPerPage", itemsPerPage),
			logger.FError(err),
		)
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	prevPageParameters := []any{pagination.PrevPage(), pagination.ItemsPerPage}
	nextPageParameters := []any{pagination.NextPage(), pagination.ItemsPerPage}
	pageControlButtons, err := b.keyboardManager.PageControlKeyboardButtons(app.HistoryCallbackQueryCmdText, pagination, prevPageParameters, nextPageParameters)
	if err != nil {
		log.Error("fail to get page control keyboard buttons", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	mainMenuButton := b.keyboardManager.MainMenuKeyboardButton()
	buttons := [][]telegram.InlineKeyboardButton{
		pageControlButtons,
		{*mainMenuButton},
	}
	replyMarkup = &telegram.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}
	responseText := b.formatterWorker.SHSHistories(langTag, smsHistories)
	return b.AnswerCallbackQueryWithEditMessageMedia(
		callbackQuery,
		responseText,
		historyImageURL,
		replyMarkup,
	)
}

func (b *botController) helpCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Debug("fail to retrieve language code", logger.F("langTag", langTag))
	}
	localizer := b.container.GetLocalizer(langTag)
	replyMarkup, err := b.keyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Error("fail to get main menu keyboard markup", logger.FError(err))
		return err
	}
	return b.AnswerCallbackQueryWithEditMessageMedia(
		callbackQuery,
		localizer.LocalizedString("help_cmd_text_markdown"),
		helpImageURL,
		replyMarkup,
	)
}

func (b *botController) developingCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Debug("fail to retrieve language code", logger.F("langTag", langTag))
		return err
	}
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      utils.NewString(b.container.GetLocalizer(langTag).LocalizedString("development_process_markdown")),
		ShowAlert: true,
	}
	return b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod)
}

func (b *botController) mainMenuCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	telegramID := callbackQuery.From.ID
	if err := b.sessionService.ClearBotStateForUser(ctx, telegramID); err != nil {
		log.Error("fail to clear bot state for user", logger.FError(err), logger.F("telegramID", telegramID))
		return err
	}
	if err := b.editMessageMainMenu(ctx, callbackQuery); err != nil {
		log.Error("fail to send main menu to telegram servers", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) servicesCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Debug("fail to retrieve language code", logger.F("langTag", langTag))
	}
	localizer := b.container.GetLocalizer(langTag)
	replyMarkup, err := b.keyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Error("fail to get main menu inline keyboard", logger.FError(err))
		return err
	}
	telegramCallbackData, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if err != nil {
		log.Error("fail to decode telegram callback data", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	if telegramCallbackData.Parameters == nil {
		log.Error("parameter must contains not empty value")
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	parameters := *telegramCallbackData.Parameters
	currentPage := utils.GetInt64(parameters[0])
	itemsPerPage := 16
	smsServices, err := b.smsActivateWorker.GetOrderedServices()
	if err != nil {
		log.Error("fail to get ordered services", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	pagination := app.Pagination{
		CurrentPage:  int(currentPage),
		LenItems:     len(smsServices),
		ItemsPerPage: itemsPerPage,
	}
	replyMarkup, err = b.keyboardManager.ServicesInlineKeyboardMarkup(smsServices, pagination)
	if err != nil {
		log.Error("fail to get services inline keyboard markup", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	return b.AnswerCallbackQueryWithEditMessageMedia(
		callbackQuery,
		localizer.LocalizedString("select_sms_service_markdown"),
		chooseServiceImageURL,
		replyMarkup,
	)
}

func (b *botController) selectServiceCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	telegramID := callbackQuery.From.ID
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Debug("fail to retrieve language code", logger.F("langTag", langTag))
	}
	localizer := b.container.GetLocalizer(langTag)
	replyMarkup, err := b.keyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Error("fail to get main menu inline keyboard", logger.FError(err))
		return err
	}
	telegramCallbackData, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if err != nil {
		log.Error("fail to decode telegram callback data", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	if telegramCallbackData.Parameters == nil {
		log.Error("parameter must contains not empty value")
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	parameters := *telegramCallbackData.Parameters
	selectedServiceCode, ok := parameters[0].(string)
	if !ok {
		log.Error("parameters at first position must contains selected service parameter")
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	currentPage := utils.GetInt64(parameters[1])
	itemsPerPage := 16
	profile, err := b.profileRepository.FetchByTelegramID(ctx, telegramID)
	if err != nil {
		log.Error("fail to fetch profile by telegram id", logger.F("telegram_id", profile.TelegramID))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	if profile.PreferredCurrency == nil {
		log.Error("profile must have preferred currency")
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	preferredCurrency := *profile.PreferredCurrency
	servicePrices, err := b.smsActivateWorker.GetPriceForService(selectedServiceCode)
	if err != nil {
		log.Error("fail to GetPriceForService", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	countries, err := b.smsService.GetCountries()
	if err != nil {
		log.Error("fail to GetCountries", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	pagination := app.Pagination{
		CurrentPage:  int(currentPage),
		ItemsPerPage: itemsPerPage,
		LenItems:     len(servicePrices),
	}
	replyMarkup, err = b.keyboardManager.ServiceCountriesInlineKeyboardMarkup(selectedServiceCode, preferredCurrency, pagination, servicePrices, countries)
	if err != nil {
		log.Error("fail to get service countries inline keyboard markup", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	return b.AnswerCallbackQueryWithEditMessageMedia(
		callbackQuery,
		localizer.LocalizedString("select_sms_service_with_country_markdown"),
		chooseCountryImageURL,
		replyMarkup,
	)
}

func (b *botController) preferredCurrenciesQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	telegramID := callbackQuery.From.ID
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Debug("fail to retrieve language code", logger.F("langTag", langTag))
	}
	localizer := b.container.GetLocalizer(langTag)
	replyMarkup, err := b.keyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Error("fail to get main menu inline keyboard", logger.FError(err))
		return err
	}
	telegramProfile, err := b.profileRepository.FetchByTelegramID(ctx, telegramID)
	if err != nil {
		log.Error("fail to get profile by telegram id", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	preferredCurrency := telegramProfile.PreferredCurrency
	if preferredCurrency == nil {
		log.Error("preferredCurrency must not has nil value")
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	currency := b.container.GetConfig().CurrencyByAbbr(*preferredCurrency)
	replyMarkup, err = b.keyboardManager.PreferredCurrenciesKeyboardMarkup()
	if err != nil {
		log.Error("fail to get a currencies keyboard markup", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	return b.AnswerCallbackQueryWithEditMessageMedia(
		callbackQuery,
		localizer.LocalizedStringWithTemplateData("choose_preferred_currency_markdown", map[string]any{
			"Currency": currency.ABBR,
		}),
		selectPreferredCurrencyImageURL,
		replyMarkup,
	)
}

func (b *botController) payServiceQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	telegramID := callbackQuery.From.ID
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Debug("fail to retrieve language code", logger.F("langTag", langTag))
	}
	localizer := b.container.GetLocalizer(langTag)
	replyMarkup, err := b.keyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Error("fail to get main menu inline keyboard", logger.FError(err))
		return err
	}
	telegramCallbackData, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if err != nil {
		log.Error("fail to decode telegram callback data", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	if telegramCallbackData.Parameters == nil {
		log.Error("parameter must contains not empty value")
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	parameters := *telegramCallbackData.Parameters
	if len(parameters) < 3 {
		log.Error("not enough length parameters")
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	serviceCode, ok := parameters[0].(string)
	if !ok {
		log.Error("serviceCode isn't string")
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	countryID := utils.GetInt64(parameters[1])
	maxPrice := utils.GetFloat64(parameters[2])
	priceWithFee := b.exchangeRateWorker.PriceWithFee(maxPrice)
	hasSufficientFunds, err := b.profileRepository.HasSufficientFunds(ctx, telegramID, priceWithFee)
	if err != nil {
		log.Error("fail to check that a profile has sufficient funds", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	if !hasSufficientFunds {
		log.Error("hasn't sufficient funds for buy service")
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("insufficient_funds_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	domainProfile, err := b.profileRepository.FetchByTelegramID(ctx, telegramID)
	if err != nil {
		log.Debug("fail to get profile", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("insufficient_funds_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	requestedNumber, err := b.smsService.RequestNumber(serviceCode, countryID, maxPrice)
	smsError, ok := err.(sms.Error)
	if ok && strings.EqualFold(smsError.Name, sms.NoNumbersErrorName) {
		log.Error("no numbers available", logger.FError(smsError))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("numbers_are_unavailable_markdown"),
			failReceivedCodeImageURL,
			replyMarkup,
		)
	} else if ok {
		log.Error("other sms activation error", logger.FError(smsError))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("fail_to_order_sms_code"),
			failReceivedCodeImageURL,
			replyMarkup,
		)
	} else if err != nil {
		log.Debug("unhandled error occurred", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			failReceivedCodeImageURL,
			replyMarkup,
		)
	}
	activationID, err := strconv.ParseInt(requestedNumber.ActivationID, 10, 64)
	if err != nil {
		log.Error("convert ActivationID to string has failed", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	domainSMSHistory := domain.SMSHistory{
		ProfileID:    domainProfile.ID,
		ActivationID: activationID,
		Status:       string(app.PendingSMSActivateState),
		ServiceCode:  serviceCode,
		PhoneNumber:  utils.PhoneNumberTitle(requestedNumber.PhoneNumber),
	}
	if _, err := b.smsHistoryRepository.Create(ctx, &domainSMSHistory); err != nil {
		log.Error("fail to create sms history", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	if err := b.profileRepository.Debit(ctx, telegramID, priceWithFee); err != nil {
		log.Error("fail to withdraw money from account")
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	responseText := localizer.LocalizedStringWithTemplateData("start_registration_form_with_sms_code_markdown", map[string]any{
		"PhoneNumber": utils.PhoneNumberTitle(requestedNumber.PhoneNumber),
	})
	if err := b.postponeService.ScheduleCheckSMSActivation(ctx, telegramID, activationID, priceWithFee); err != nil {
		log.Error("fail to prepare schedule to check the sms activation", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	return b.AnswerCallbackQueryWithEditMessageMedia(callbackQuery, responseText, avatarImageURL, replyMarkup)
}

func (b *botController) emptyQueryCommandHandler(_ context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	if err := b.AnswerCallbackQuery(callbackQuery); err != nil {
		log.Error("fail to answer callback query", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) selectPreferredCurrencyQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	telegramID := callbackQuery.From.ID
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	if err != nil {
		log.Debug("fail to retrieve language code", logger.F("langTag", langTag))
	}
	localizer := b.container.GetLocalizer(langTag)
	replyMarkup, err := b.keyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Error("fail to get main menu inline keyboard", logger.FError(err))
		return err
	}
	telegramCallbackData, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if err != nil {
		log.Error("fail to decode telegram callback data", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	if telegramCallbackData.Parameters == nil {
		log.Error("parameter must contains not empty value")
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	parameters := *telegramCallbackData.Parameters
	selectedPreferredCurrency := parameters[0].(string)
	if err := b.profileRepository.SetPreferredCurrency(ctx, telegramID, selectedPreferredCurrency); err != nil {
		log.Error("fail to set preferred currency to profile", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	return b.editMessageMainMenu(ctx, callbackQuery)
}

func (b *botController) deleteCryptoBotQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, callbackQuery.From)
	telegramUser := callbackQuery.From
	if err != nil {
		log.Debug("fail to retrieve language code", logger.F("langTag", langTag))
	}
	localizer := b.container.GetLocalizer(langTag)
	replyMarkup, err := b.keyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Error("fail to get main menu inline keyboard", logger.FError(err))
		return err
	}
	telegramCallbackQueryData, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if err != nil {
		log.Error("fail to decode telegram callback data", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	if telegramCallbackQueryData.Parameters == nil {
		log.Error("parameters must contains parameters")
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	parameters := *telegramCallbackQueryData.Parameters
	invoiceID, ok := parameters[0].(int64)
	if !ok {
		log.Error("parameters must contains invoiceID")
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	if err := b.cryptoPayBot.RemoveInvoice(invoiceID); err != nil {
		log.Error("fail to remove invoice", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	deleteMessage := telegram.DeleteMessage{
		ChatID:    callbackQuery.Message.Chat.ID,
		MessageID: callbackQuery.Message.ID,
	}
	err = b.telegramBotService.SendResponse(deleteMessage, app.DeleteMessageTelegramMethod)
	if err != nil && errors.Is(err, app.DeleteInvoiceError) {
		log.Debug("fail to remove invoice", logger.FError(err))
	} else if err != nil {
		log.Debug("other unknown error while perform remove invoice", logger.FError(err))
		return b.AnswerCallbackQueryWithEditMessageMedia(
			callbackQuery,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	if err == nil {
		err = b.SendTextWithPhotoMedia(
			callbackQuery.Message.Chat.ID,
			localizer.LocalizedString("success_deleted_invoice_markdown"),
			avatarImageURL,
			nil,
		)
		if err != nil {
			log.Error("fail to send message to telegram")
			return b.AnswerCallbackQueryWithEditMessageMedia(
				callbackQuery,
				localizer.LocalizedString("internal_error_markdown"),
				avatarImageURL,
				replyMarkup,
			)
		}
	}
	return b.messageMainMenu(ctx, callbackQuery.Message.Chat.ID, &telegramUser)
}
