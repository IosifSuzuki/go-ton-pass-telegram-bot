package telegram

import (
	"context"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/crypto/bot"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
	"strconv"
)

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
		log.Debug("fail to retrieve language code", logger.F("langTag", langTag))
	}
	localizer := b.container.GetLocalizer(langTag)
	telegramUser, err := getTelegramUser(update)
	if err != nil {
		log.Error("fail to get telegram user from update", logger.FError(err))
		return err
	}

	mainMenuInlineKeyboardMarkup, err := b.keyboardManager.MainMenuKeyboardMarkup()
	if err != nil {
		log.Debug("fail to get menu inline keyboard markup", logger.FError(err))
		return b.SendTextWithPhotoMedia(
			update,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			mainMenuInlineKeyboardMarkup,
		)
	}
	if update.Message.Text == nil {
		log.Debug("text has nil value")
		return b.SendTextWithPhotoMedia(
			update,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			mainMenuInlineKeyboardMarkup,
		)
	}
	text := *update.Message.Text
	amount, err := strconv.ParseFloat(text, 64)
	if err != nil {
		log.Debug("fail to parse number from user input", logger.F("text", text), logger.FError(err))
		return b.SendTextWithPhotoMedia(
			update,
			localizer.LocalizedString("enter_amount_for_payment_in_currency_error_markdown"),
			avatarImageURL,
			mainMenuInlineKeyboardMarkup,
		)
	}
	currency, err := b.sessionService.GetString(ctx, service.SelectedPayCurrencyAbbrKey, telegramUser.ID)
	if err != nil {
		log.Debug("fail to get pay currency from session service", logger.FError(err))
		return b.SendTextWithPhotoMedia(
			update,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			mainMenuInlineKeyboardMarkup,
		)
	}
	profile, err := b.profileRepository.FetchByTelegramID(ctx, telegramUser.ID)
	if err != nil {
		log.Debug("fail to get profile by telegram id", logger.F("telegram id", telegramUser.ID), logger.FError(err))
		return b.SendTextWithPhotoMedia(
			update,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			mainMenuInlineKeyboardMarkup,
		)
	}
	if currency == nil || profile.PreferredCurrency == nil {
		log.Debug(
			"currency or preferred_currency has nil value",
			logger.F("currency", currency),
			logger.F("preferred_currency", profile.PreferredCurrency),
		)
		return b.SendTextWithPhotoMedia(
			update,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			mainMenuInlineKeyboardMarkup,
		)
	}
	targetAmount, err := b.exchangeRateWorker.Convert(amount, *profile.PreferredCurrency, *currency)
	if err != nil {
		log.Debug(
			"fail to convert",
			logger.F("source_code", *profile.PreferredCurrency),
			logger.F("target_code", *currency),
		)
		return b.SendTextWithPhotoMedia(
			update,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			mainMenuInlineKeyboardMarkup,
		)
	}
	if targetAmount == nil {
		log.Debug("targetAmount must contains value")
		return b.SendTextWithPhotoMedia(
			update,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			mainMenuInlineKeyboardMarkup,
		)
	}
	invoicePayload := bot.InvoicePayload{
		ChatID:     update.Message.Chat.ID,
		TelegramID: telegramUser.ID,
	}
	encodedInvoicePayload, err := utils.EncodeCryptoBotInvoicePayload(invoicePayload)
	if err != nil {
		log.Debug("fail to encode a invoice payload", logger.FError(err))
		return b.SendTextWithPhotoMedia(
			update,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			mainMenuInlineKeyboardMarkup,
		)
	}
	invoice, err := b.cryptoPayBot.CreateInvoice(*currency, *targetAmount, *encodedInvoicePayload)
	if err != nil {
		log.Debug("fail to create a invoice", logger.FError(err))
		return b.SendTextWithPhotoMedia(
			update,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			mainMenuInlineKeyboardMarkup,
		)
	}
	cryptoPayInlineKeyboardMarkup, err := b.keyboardManager.CryptoPayBotKeyboardMarkup(invoice.BotInvoiceURL)
	if err != nil {
		log.Debug("fail to get cryptoPayInlineKeyboardMarkup", logger.FError(err))
		return b.SendTextWithPhotoMedia(
			update,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			mainMenuInlineKeyboardMarkup,
		)
	}
	return b.SendTextWithPhotoMedia(
		update,
		localizer.LocalizedString("crypto_bot_pay_title_markdown"),
		avatarImageURL,
		cryptoPayInlineKeyboardMarkup,
	)
}
