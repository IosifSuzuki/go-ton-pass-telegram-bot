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

func (b *botController) userSelectedLanguageBotStageHandler(ctx context.Context, update *telegram.Update) error {
	telegramID := update.Message.From.ID
	if update.Message.Text == nil {
		return app.NilError
	}
	selectedLanguageNativeName := *update.Message.Text

	availableLanguages := b.container.GetConfig().AvailableLanguages()
	filteredLanguages := utils.Filter(availableLanguages, func(language app.Language) bool {
		presentableLanguageText := utils.LanguageTextFormat(language)
		return presentableLanguageText == selectedLanguageNativeName
	})
	if len(filteredLanguages) == 0 {
		return app.UnknownLanguageError
	}
	selectedLanguage := filteredLanguages[0]

	b.keyboardManager.Set(selectedLanguage.Code)
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
		presentableCurrencyText := utils.ShortCurrencyTextFormat(currency)
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
		log.Debug("fail to retrieve language code", logger.F("langTag", langTag))
	}
	localizer := b.container.GetLocalizer(langTag)
	telegramUser, err := getTelegramUser(update)
	if err != nil {
		log.Error("fail to get telegram user from update", logger.FError(err))
		return err
	}

	mainMenuInlineKeyboardMarkup, err := b.keyboardManager.MainMenuInlineKeyboardMarkup()
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
	cryptoPayInlineKeyboardMarkup, err := b.getCryptoPayBotKeyboardMarkup(langTag, invoice.BotInvoiceURL)
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
