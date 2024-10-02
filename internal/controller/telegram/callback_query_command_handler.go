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
	backToMenuInlineKeyboardMarkup, err := b.getMenuInlineKeyboardMarkup(*langTag)
	if err != nil {
		return err
	}
	editMessage := telegram.EditMessage{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Text:        balanceText,
		ReplyMarkup: backToMenuInlineKeyboardMarkup,
	}
	if err := b.telegramBotService.SendResponse(editMessage, app.EditMessageTextTelegramMethod); err != nil {
		log.Error("fail to send a EditMessage to telegram servers", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) languagesCallbackQueryCommandHandle(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	telegramID := callbackQuery.From.ID
	telegramProfile, err := b.profileRepository.FetchByTelegramID(ctx, telegramID)
	b.container.GetConfig().AvailableLanguages()
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

func (b *botController) selectLanguageCallbackQueryCommandHandle(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
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

func (b *botController) historyCallbackQueryCommandHandle(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
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

func (b *botController) helpCallbackQueryCommandHandle(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
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

func (b *botController) unsupportedCallbackQueryCommandHandle(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
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

func (b *botController) servicesCallbackQueryCommandHandle(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
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

func (b *botController) selectServiceCallbackQueryCommandHandle(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
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
