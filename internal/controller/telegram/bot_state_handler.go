package telegram

import (
	"context"
	"go-ton-pass-telegram-bot/internal/model/crypto/bot"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
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
	currency, err := b.sessionService.GetString(ctx, service.SelectedPayCurrencyAbbrSessionKey, telegramID)
	if err != nil {
		log.Error(
			"fail to get pay currency from session service",
			logger.FError(err),
		)
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}
	if currency == nil || ctxOptions.Profile.PreferredCurrency == nil {
		log.Error(
			"currency or preferred_currency has nil value",
			logger.F("currency", currency),
			logger.F("preferred_currency", ctxOptions.Profile.PreferredCurrency),
		)
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}
	preferredCurrency := *ctxOptions.Profile.PreferredCurrency
	convertedAmount, err := b.exchangeRateWorker.Convert(amount, preferredCurrency, *currency)
	if err != nil {
		log.Error(
			"fail to convert",
			logger.F("source_code", preferredCurrency),
			logger.F("target_code", *currency),
		)
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}
	if convertedAmount == nil {
		log.Error("convertedAmount must contains value")
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}
	invoicePayload := bot.InvoicePayload{
		ChatID:     ctxOptions.Update.GetChatID(),
		TelegramID: telegramID,
	}
	encodedInvoicePayload, err := utils.EncodeCryptoBotInvoicePayload(invoicePayload)
	if err != nil {
		log.Error("fail to encode a invoice payload", logger.FError(err))
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}
	invoice, err := b.cryptoPayBot.CreateInvoice(*currency, *convertedAmount, *encodedInvoicePayload)
	if err != nil {
		log.Error("fail to create a invoice", logger.FError(err))
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}
	if err := b.sessionService.ClearBotStateForUser(ctx, telegramID); err != nil {
		log.Error(
			"fail to clear bot state",
			logger.FError(err),
			logger.F("telegram_id", telegramID),
		)
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	if err := b.sessionService.ClearString(ctx, service.SelectedPayCurrencyAbbrSessionKey, telegramID); err != nil {
		log.Error(
			"fail to clear selected pay currency string",
			logger.FError(err),
			logger.F("telegram_id", telegramID),
		)
		return b.editMessageInternalServerError(ctx, ctxOptions)
	}
	return b.sendMessageConfirmTouchUpBalance(ctx, ctxOptions, invoice)
}
