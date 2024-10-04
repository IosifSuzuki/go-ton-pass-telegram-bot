package telegram

import (
	"context"
	"fmt"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/crypto/bot"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
	"strconv"
)

func (b *botController) userSelectedLanguageBotStageHandler(ctx context.Context, update *telegram.Update) error {
	telegramID := update.Message.From.ID
	if update.Message.Text == nil {
		return app.NilError
	}
	selectedLanguageNativeName := *update.Message.Text

	availableLanguages := b.container.GetConfig().AvailableLanguages()
	filteredLanguages := utils.Filter(availableLanguages, func(language app.Language) bool {
		presentableLanguageText := fmt.Sprintf("%s %s", language.FlagEmoji, language.NativeName)
		return presentableLanguageText == selectedLanguageNativeName
	})
	if len(filteredLanguages) == 0 {
		return app.UnknownLanguageError
	}
	selectedLanguage := filteredLanguages[0]

	if err := b.profileRepository.SetPreferredLanguage(ctx, telegramID, selectedLanguage.Code); err != nil {
		return err
	}
	if err := b.messageWelcome(ctx, update); err != nil {
		return err
	}
	return b.messageToSelectPreferredCurrency(ctx, update)
}

func (b *botController) userSelectedPreferredCurrencyBotStageHandler(ctx context.Context, update *telegram.Update) error {
	telegramID := update.Message.From.ID
	if update.Message.Text == nil {
		return app.NilError
	}
	selectedCurrencyText := *update.Message.Text

	availableCurrencies := b.container.GetConfig().AvailablePreferredCurrencies()
	filteredCurrencies := utils.Filter(availableCurrencies, func(currency app.Currency) bool {
		presentableCurrencyText := fmt.Sprintf("%s %s", currency.Symbol, currency.ABBR)
		return presentableCurrencyText == selectedCurrencyText
	})
	if len(filteredCurrencies) == 0 {
		return app.UnknownLanguageError
	}
	selectedCurrency := filteredCurrencies[0]
	if err := b.profileRepository.SetPreferredCurrency(ctx, telegramID, selectedCurrency.ABBR); err != nil {
		return err
	}
	if err := b.sessionService.ClearBotStateForUser(ctx, telegramID); err != nil {
		return err
	}
	return b.messageMainMenu(ctx, update)
}

func (b *botController) enterAmountCurrencyBotStageHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	telegramID := callbackQuery.From.ID
	telegramCallbackData, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if err != nil {
		return err
	}
	parameters := *telegramCallbackData.Parameters
	selectedPayCurrencyAbbr := parameters[0].(string)
	if err := b.sessionService.SaveString(ctx, service.SelectedPayCurrencyAbbrKey, selectedPayCurrencyAbbr, telegramID); err != nil {
		return err
	}
	if err := b.sessionService.SaveBotStateForUser(ctx, app.EnteringAmountCurrencyBotState, telegramID); err != nil {
		return err
	}
	return b.messageEnterAmountCurrency(ctx, callbackQuery)
}

func (b *botController) enteringAmountCurrencyBotStageHandler(ctx context.Context, update *telegram.Update) error {
	log := b.container.GetLogger()
	langTag, err := b.getLanguageCode(ctx, *update.Message.From)
	if err != nil {
		log.Error("getLanguageCode has failed", logger.FError(err))
		return err
	}
	localizer := b.container.GetLocalizer(*langTag)
	telegramID, err := getTelegramID(update)
	if err != nil {
		log.Error("telegram id is missing", logger.FError(err))
		return err
	}
	if update.Message.Text == nil {
		log.Error("text has nil value")
		return app.NilError
	}
	text := *update.Message.Text
	amount, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return b.messageWithPlainText(ctx, localizer.LocalizedString("enter_amount_for_payment_in_currency_error"), update)
	}
	currency, err := b.sessionService.GetString(ctx, service.SelectedPayCurrencyAbbrKey, *telegramID)
	if err != nil {
		log.Debug("fail to get selected pay currency", logger.FError(err))
		return b.messageMainMenu(ctx, update)
	}
	invoicePayload := bot.InvoicePayload{
		ChatID:     update.Message.Chat.ID,
		TelegramID: *telegramID,
	}
	encodedInvoicePayload, err := utils.EncodeCryptoBotInvoicePayload(invoicePayload)
	if err != nil {
		log.Error("fail to encode a invoice payload", logger.FError(err))
		return err
	}
	invoice, err := b.cryptoPayBot.CreateInvoice(*currency, amount, *encodedInvoicePayload)
	if err != nil {
		log.Error("fail to create a invoice", logger.FError(err))
		return err
	}
	replyMarkup, err := b.getCryptoPayBotKeyboardMarkup(*langTag, invoice.BotInvoiceURL)
	if err != nil {
		log.Error("fail to get cryptoPayBotKeyboardMarkup", logger.FError(err))
		return err
	}
	resp := telegram.SendResponse{
		ChatID:      update.Message.Chat.ID,
		Text:        localizer.LocalizedString("crypto_bot_pay_title"),
		ReplyMarkup: replyMarkup,
	}
	return b.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}
