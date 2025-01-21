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

const avatarImageURL = "https://i.ibb.co/rmqsKty/avatar.png"

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
	replyMarkup := telegram.ReplyKeyboardRemove{
		RemoveKeyboard: true,
	}
	paidUsdRate, err := strconv.ParseFloat(*update.PayloadInvoice.PaidUsdRate, 64)
	if err != nil {
		log.Debug("paidUsdRate has unknown float format", logger.FError(err))
		return c.SendTextWithPhotoMedia(
			profile.TelegramChatID,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	amount, err := strconv.ParseFloat(update.PayloadInvoice.Amount, 64)
	if err != nil {
		log.Debug("amount has unknown float format", logger.FError(err))
		return c.SendTextWithPhotoMedia(
			profile.TelegramChatID,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	amountInUSD := amount * paidUsdRate
	log.Debug("will top up balance", logger.F("amountInUSD", amountInUSD))
	if err := c.profileRepository.TopUpBalanceByTelegramID(
		ctx,
		payloadInvoice.TelegramID,
		amountInUSD,
	); err != nil {
		log.Debug("fail to top up balance", logger.FError(err))
		return c.SendTextWithPhotoMedia(
			profile.TelegramChatID,
			localizer.LocalizedString("internal_error_markdown"),
			avatarImageURL,
			replyMarkup,
		)
	}
	return c.SendTextWithPhotoMedia(
		profile.TelegramChatID,
		localizer.LocalizedString("balance_updated_markdown"),
		avatarImageURL,
		replyMarkup,
	)
}

func (c *cryptoController) SendTextWithPhotoMedia(chatID int64, text string, photoURL string, replyMarkup any) error {
	log := c.container.GetLogger()
	resp := telegram.SendPhoto{
		ChatID:      chatID,
		Caption:     text,
		Photo:       photoURL,
		ParseMode:   utils.NewString("MarkdownV2"),
		ReplyMarkup: replyMarkup,
	}
	if err := c.telegramBotService.SendResponse(resp, app.SendPhotoTelegramMethod); err != nil {
		log.Debug("fail to send message with photo media", logger.FError(err))
		return err
	}
	return nil
}
