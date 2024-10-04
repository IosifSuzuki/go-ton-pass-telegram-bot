package crypto

import (
	"context"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/crypto/bot"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/repository"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
	"strconv"
)

type CryptoController interface {
	Serve(update *bot.WebhookUpdates) error
}

type cryptoController struct {
	container          container.Container
	telegramBotService service.TelegramBotService
	cryptoPayBot       service.CryptoPayBot
	sessionService     service.SessionService
	profileRepository  repository.ProfileRepository
}

func NewCryptoController(
	container container.Container,
	sessionService service.SessionService,
	profileRepository repository.ProfileRepository,
) CryptoController {
	return &cryptoController{
		container:          container,
		telegramBotService: service.NewTelegramBot(container),
		cryptoPayBot:       service.NewCryptoPayBot(container),
		sessionService:     sessionService,
		profileRepository:  profileRepository,
	}
}

func (c *cryptoController) Serve(update *bot.WebhookUpdates) error {
	ctx := context.Background()
	log := c.container.GetLogger()
	invoice := update.PayloadInvoice
	if invoice == nil {
		log.Debug("invoice is missing")
		return app.NilError
	}
	payloadInvoiceEncodedText := invoice.Payload
	if payloadInvoiceEncodedText == nil {
		log.Error("expected payload invoice")
		return app.NilError
	}
	payloadInvoice, err := utils.DecodeCryptoBotInvoicePayload(*payloadInvoiceEncodedText)
	if err != nil {
		log.Error("decoded CryptoBotInvoicePayload has failed", logger.FError(err))
		return err
	}
	profile, err := c.profileRepository.FetchByTelegramID(ctx, payloadInvoice.TelegramID)
	if err != nil {
		log.Error("fetchByTelegramID has failed", logger.FError(err))
		return err
	}
	localizer := c.container.GetLocalizer(*profile.PreferredLanguage)
	paidUsdRate, err := strconv.ParseFloat(*update.PayloadInvoice.PaidUsdRate, 64)
	if err != nil {
		log.Error("paidUsdRate has unknown float format", logger.FError(err))
		return err
	}
	amount, err := strconv.ParseFloat(update.PayloadInvoice.Amount, 64)
	if err != nil {
		log.Error("amount has unknown float format", logger.FError(err))
		return err
	}
	amountInUSD := amount * paidUsdRate
	log.Debug("will top up balance", logger.F("amountInUSD", amountInUSD))
	if err := c.profileRepository.TopUpBalance(ctx, payloadInvoice.TelegramID, amountInUSD); err != nil {
		return err
	}
	resp := telegram.SendResponse{
		ChatID: payloadInvoice.ChatID,
		Text:   localizer.LocalizedString("balance_updated"),
	}
	if err := c.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod); err != nil {
		log.Error("fail to send message with success top up balance", logger.FError(err))
		return err
	}
	return nil
}
