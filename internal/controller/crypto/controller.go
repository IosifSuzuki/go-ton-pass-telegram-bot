package crypto

import (
	"context"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/crypto/bot"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/repository"
	"go-ton-pass-telegram-bot/internal/service"
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
	invoiceRepository  repository.InvoiceRepository
}

func NewCryptoController(
	container container.Container,
	sessionService service.SessionService,
	profileRepository repository.ProfileRepository,
	invoiceRepository repository.InvoiceRepository,
) CryptoController {
	return &cryptoController{
		container:          container,
		telegramBotService: service.NewTelegramBot(container),
		cryptoPayBot:       service.NewCryptoPayBot(container),
		sessionService:     sessionService,
		profileRepository:  profileRepository,
		invoiceRepository:  invoiceRepository,
	}
}

func (c *cryptoController) Serve(update *bot.WebhookUpdates) error {
	ctx := context.Background()
	log := c.container.GetLogger()
	payloadInvoice := update.PayloadInvoice
	if payloadInvoice == nil {
		return app.NilError
	}
	invoice, err := c.invoiceRepository.GetInvoiceByInvoiceID(ctx, payloadInvoice.ID)
	if err != nil {
		return err
	}
	profile, err := c.profileRepository.FetchByID(ctx, invoice.ProfileID)
	if err != nil {
		return err
	}
	localizer := c.container.GetLocalizer(*profile.PreferredLanguage)
	if err := c.invoiceRepository.UpdateStatus(ctx, invoice.InvoiceID, invoice.Status); err != nil {
		return err
	}
	if update.PayloadInvoice.PaidUsdRate == nil {
		log.Error("PaidUsdRate has nil")
		return app.NilError
	}
	paidUsdRate, err := strconv.ParseFloat(*update.PayloadInvoice.PaidUsdRate, 64)
	if err != nil {
		log.Error("ParseFloat of PaidUsdRate has failed", logger.FError(err))
		return err
	}
	amount, err := strconv.ParseFloat(update.PayloadInvoice.Amount, 64)
	if err != nil {
		log.Error("ParseFloat of amount has failed", logger.FError(err))
		return err
	}
	log.Debug("will top up balance", logger.F("paidUsdRate", paidUsdRate))
	if err := c.profileRepository.TopUpBalance(ctx, invoice.ProfileID, paidUsdRate*amount); err != nil {
		return err
	}
	resp := telegram.SendResponse{
		ChatID: invoice.ChatID,
		Text:   localizer.LocalizedString("balance_updated"),
	}
	return c.telegramBotService.SendResponse(resp, app.SendMessageTelegramMethod)
}
