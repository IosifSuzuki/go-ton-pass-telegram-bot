package telegram

import (
	"context"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/sms"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/service"
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
	editMessage := telegram.EditMessage{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Text:        balanceText,
		ReplyMarkup: b.telegramBotService.GetBackToMenuInlineKeyboardMarkup(*langTag),
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
	editMessage := telegram.EditMessage{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Text:        b.container.GetLocalizer(*langTag).LocalizedString("short_description"),
		ReplyMarkup: b.telegramBotService.GetMenuInlineKeyboardMarkup(*langTag),
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
		Text:      utils.NewString(b.container.GetLocalizer(*langTag).LocalizedString("development_process")),
		ShowAlert: true,
	}
	if err := b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod); err != nil {
		log.Error("fail to send a AnswerCallbackQuery to telegram servers", logger.FError(err))
		return err
	}
	smsServices, err := b.smsService.GetListOfAllServices()
	if err != nil {
		return err
	}
	editMessage := telegram.EditMessage{
		ChatID:      &callbackQuery.Message.Chat.ID,
		MessageID:   &callbackQuery.Message.ID,
		Text:        b.container.GetLocalizer(*langTag).LocalizedString("select_sms_service"),
		ReplyMarkup: b.getSMSServicesInlineKeyboardMarkup(smsServices),
	}
	if err := b.telegramBotService.SendResponse(editMessage, app.EditMessageTextTelegramMethod); err != nil {
		log.Error("fail to send a EditMessage to telegram servers", logger.FError(err))
		return err
	}
	return nil
}

func (b *botController) getSMSServicesInlineKeyboardMarkup(smsServices []sms.Service) *telegram.InlineKeyboardMarkup {
	inlineKeyboardButtons := make([]telegram.InlineKeyboardButton, 0, len(smsServices))
	for _, smsService := range smsServices {
		inlineKeyboardButtons = append(inlineKeyboardButtons, telegram.InlineKeyboardButton{
			Text: smsService.Name,
			Data: utils.NewString(service.MainMenuCallbackQueryCmdText),
		})
	}
	columns := 3
	rows := len(inlineKeyboardButtons) / columns

	orderedInlineKeyboardButtons := make([][]telegram.InlineKeyboardButton, 0, rows)
	for i := 0; i < rows; i++ {
		start := i * columns
		end := start + columns
		orderedInlineKeyboardButtons = append(orderedInlineKeyboardButtons, inlineKeyboardButtons[start:end])
	}
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: orderedInlineKeyboardButtons,
	}
}
