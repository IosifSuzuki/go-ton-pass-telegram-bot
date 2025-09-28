package telegram

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
	"math"
)

func (b *botController) enteringAmountCurrencyBotStageHandler(ctx context.Context, ctxOptions *ContextOptions) error {
	log := b.container.GetLogger()
	text := ctxOptions.Update.Message.Text
	telegramID := ctxOptions.Update.GetTelegramID()
	if text == nil {
		log.Error("text has nil value")
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}

	amount, err := utils.ParseFloat64FromText(*text)
	if err != nil {
		log.Error(
			"fail to parse number from user input",
			logger.F("text", text),
			logger.FError(err),
		)
		return b.editMessageEnterAmountPayError(ctx, ctxOptions)
	}
	paymentMethod, err := b.sessionService.GetString(ctx, service.SelectedPaymentMethodSessionKey, telegramID)
	if err != nil {
		log.Error("fail to get payment method from session service", logger.FError(err))
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}
	if paymentMethod == nil {
		log.Error("paymentMethod is nil")
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}
	var (
		selectedCurrency *string
	)
	switch *paymentMethod {
	case app.TelegramStarsPaymentMethod:
		selectedCurrency = utils.NewString("XTR")
	case app.CryptoBotPaymentMethod:
		currency, err := b.sessionService.GetString(ctx, service.SelectedCryptoBotPayCurrencyAbbrSessionKey, telegramID)
		if err != nil {
			log.Error(
				"fail to get the pay currency from session service",
				logger.FError(err),
			)
			return b.sendMessageInternalServerError(ctx, ctxOptions)
		}
		selectedCurrency = currency
	case app.StripePaymentMethod:
		selectedCurrency = ctxOptions.Profile.PreferredCurrency
	default:
		log.Error("detected unimplemented payment method")
	}
	if selectedCurrency == nil || ctxOptions.Profile.PreferredCurrency == nil {
		log.Error(
			"currency or preferred_currency has nil value",
			logger.F("currency", selectedCurrency),
			logger.F("preferred_currency", ctxOptions.Profile.PreferredCurrency),
		)
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}
	preferredCurrency := *ctxOptions.Profile.PreferredCurrency
	convertedAmount, err := b.exchangeRateWorker.Convert(amount, preferredCurrency, *selectedCurrency)
	if err != nil {
		log.Error(
			"fail to convert",
			logger.F("source_code", preferredCurrency),
			logger.F("target_code", *selectedCurrency),
		)
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}
	if convertedAmount == nil {
		log.Error("convertedAmount must contains value")
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}

	switch *paymentMethod {
	case app.CryptoBotPaymentMethod:
		return b.sendCryptoBotInvoice(ctx, ctxOptions, selectedCurrency, convertedAmount)
	case app.TelegramStarsPaymentMethod:
		amountInUSD, err := b.exchangeRateWorker.ConvertToUSD(amount, preferredCurrency)
		if err != nil {
			log.Error(
				"fail to convert to USD",
				logger.F("amount", amount),
				logger.F("currency", selectedCurrency),
				logger.FError(err),
			)
			return b.sendMessageInternalServerError(ctx, ctxOptions)
		}
		if amountInUSD == nil {
			err := app.NilError
			log.Error("amountInUSD must contains value", logger.FError(err))
			return b.sendMessageInternalServerError(ctx, ctxOptions)
		}
		amountInXTR := int64(math.Ceil(*convertedAmount))
		return b.sendTelegramStarsInvoice(ctx, ctxOptions, *amountInUSD, amountInXTR)
	case app.StripePaymentMethod:
		return b.sendStripeInvoice(ctx, ctxOptions, selectedCurrency, amount)
	default:
		return nil
	}
}

func (b *botController) sendTelegramStarsInvoice(_ context.Context, ctxOptions *ContextOptions, creditBalance float64, stars int64) error {
	log := b.container.GetLogger()
	telegramID := ctxOptions.Update.GetTelegramID()
	preferredLanguage := b.getPreferredLanguage(ctxOptions)
	localized := b.container.GetLocalizer(preferredLanguage)

	telegramPaymentPayload := app.TelegramPaymentPayload{
		ProfileID:     ctxOptions.Profile.ID,
		CreditBalance: creditBalance,
	}
	payloadData, err := json.Marshal(telegramPaymentPayload)
	if err != nil {
		log.Error("failed to marshal payload data", logger.FError(err))
		return err
	}
	encodedPayloadData := base64.StdEncoding.EncodeToString(payloadData)
	replyMarkup, err := ctxOptions.TelegramInlineKeyboardManager.TelegramStarsPayInlineKeyboardMarkup(stars)
	if err != nil {
		log.Error("fail to get telegram stars inline keyboard markup", logger.FError(err))
		return err
	}
	sendInvoice := telegram.SendInvoice{
		ChatID:        ctxOptions.Update.GetChatID(),
		Title:         localized.LocalizedString("invoice_telegram_stars_title"),
		Description:   localized.LocalizedString("invoice_telegram_stars_description"),
		Payload:       encodedPayloadData,
		ProviderToken: "",
		Currency:      "XTR",
		Prices: []telegram.LabeledPrice{
			{
				Label: "Price",
				Price: stars,
			},
		},
		PhotoURL:       avatarImageURL,
		ProtectContent: true,
		ReplyMarkup:    replyMarkup,
	}

	if err := b.telegramBotService.SendResponse(sendInvoice, app.SendInvoiceTelegramMethod); err != nil {
		log.Error(
			"fail to send invoice",
			logger.FError(err),
			logger.F("telegram_id", telegramID),
		)
		return err
	}

	return nil
}
