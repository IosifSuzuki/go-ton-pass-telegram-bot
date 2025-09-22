package telegram

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/domain"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/pkg/logger"
)

func (b *botController) PreCheckoutHandler(ctx context.Context, ctxOptions *ContextOptions) error {
	log := b.container.GetLogger()

	preCheckoutQuery := ctxOptions.Update.PreCheckoutQuery
	answerPreCheckoutQuery := telegram.AnswerPreCheckoutQuery{
		PreCheckoutQueryID: preCheckoutQuery.ID,
		OK:                 true,
	}

	err := b.telegramBotService.SendResponse(answerPreCheckoutQuery, app.AnswerPreCheckoutQueryTelegramMethod)
	if err != nil {
		log.Error("fail to send answerPreCheckoutQuery message", logger.FError(err))
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}

	return nil
}

func (b *botController) RefundPaymentHandler(ctx context.Context, ctxOptions *ContextOptions) error {
	log := b.container.GetLogger()
	telegramID := ctxOptions.Profile.TelegramID

	refundedPayment := ctxOptions.Update.Message.RefundedPayment

	decodedPayload, err := base64.StdEncoding.DecodeString(refundedPayment.InvoicePayload)
	if err != nil {
		log.Error(
			"fail to decode refundedPayment payload",
			logger.FError(err),
		)
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}
	var telegramPaymentPayload app.TelegramPaymentPayload
	if err := json.Unmarshal(decodedPayload, &telegramPaymentPayload); err != nil {
		log.Error("fail to decode telegram payment payload", logger.FError(err))
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}

	telegramPaymentDomain, err := b.telegramPaymentRepository.FetchByTelegramPaymentChargeID(
		ctx,
		refundedPayment.TelegramPaymentChargeID,
	)
	if err != nil {
		log.Error(
			"fail to fetch telegram payment",
			logger.FError(err),
			logger.F("telegram_payment_charge_id", refundedPayment.TelegramPaymentChargeID),
		)
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}
	if telegramPaymentDomain != nil && telegramPaymentDomain.IsRefunded {
		log.Debug(
			"telegram payment is already refunded",
			logger.F("telegram_payment_charge_id", refundedPayment.TelegramPaymentChargeID),
		)
		return b.sendMessageMainMenu(ctx, ctxOptions)
	}

	log.Debug(
		"withdraw funds",
		logger.F("amount", telegramPaymentPayload.CreditBalance),
		logger.F("telegram_id", telegramID),
	)
	if err := b.profileRepository.Debit(ctx, telegramID, telegramPaymentPayload.CreditBalance); err != nil {
		log.Error(
			"fail to refund amount from balance",
			logger.FError(err),
			logger.F("telegram_id", telegramID),
			logger.F("amount", telegramPaymentPayload.CreditBalance),
		)
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}

	if err := b.telegramPaymentRepository.MarkRefunded(ctx, refundedPayment.TelegramPaymentChargeID); err != nil {
		log.Error(
			"fail to mark refunded telegram payment",
			logger.FError(err),
			logger.F("telegram_id", telegramID),
			logger.F("telegram_payment_charge_id", refundedPayment.TelegramPaymentChargeID),
		)
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}

	return b.sendMessageMainMenu(ctx, ctxOptions)
}

func (b *botController) SuccessfulPaymentHandler(ctx context.Context, ctxOptions *ContextOptions) error {
	log := b.container.GetLogger()

	successfulPayment := ctxOptions.Update.Message.SuccessfulPayment
	decodedPayload, err := base64.StdEncoding.DecodeString(successfulPayment.InvoicePayload)
	if err != nil {
		log.Error("fail to decode successfulPayment payload", logger.FError(err))
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}
	var telegramPaymentPayload app.TelegramPaymentPayload
	if err := json.Unmarshal(decodedPayload, &telegramPaymentPayload); err != nil {
		log.Error("fail to decode telegram payment payload", logger.FError(err))
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}

	err = b.profileRepository.TopUpBalanceByProfileID(
		ctx,
		telegramPaymentPayload.ProfileID,
		telegramPaymentPayload.CreditBalance,
	)
	if err != nil {
		log.Error(
			"fail to top up balance profile",
			logger.F("profile_id", telegramPaymentPayload.ProfileID),
			logger.FError(err),
		)
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}

	var telegramPaymentDomain = domain.TelegramPayment{
		ProfileID:               telegramPaymentPayload.ProfileID,
		TelegramPaymentChargeID: successfulPayment.TelegramPaymentChargeID,
		Currency:                successfulPayment.Currency,
		Amount:                  successfulPayment.TotalAmount,
		CreditAmount:            telegramPaymentPayload.CreditBalance,
	}
	if _, err := b.telegramPaymentRepository.Create(ctx, &telegramPaymentDomain); err != nil {
		log.Error(
			"fail to store telegram payment in db",
			logger.FError(err),
			logger.F("profile_id", telegramPaymentPayload.ProfileID),
		)
		return b.sendMessageInternalServerError(ctx, ctxOptions)
	}

	return b.sendMessageMainMenu(ctx, ctxOptions)
}
