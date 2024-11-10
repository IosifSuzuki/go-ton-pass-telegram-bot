package sms

import (
	"context"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/sms"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/repository"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/internal/worker"
	"go-ton-pass-telegram-bot/pkg/logger"
)

const (
	successReceivedCodeImageURL = "https://www.imghippo.com/i/yOFIj1728463916.png"
)

type SMSActivateController interface {
	Serve(update *sms.WebhookUpdates) error
}

type smsActivateController struct {
	container            container.Container
	telegramBotService   service.TelegramBotService
	profileRepository    repository.ProfileRepository
	smsHistoryRepository repository.SMSHistoryRepository
	formatterWorker      worker.Formatter
}

func NewSMSActivateController(
	container container.Container,
	profileRepository repository.ProfileRepository,
	smsHistoryRepository repository.SMSHistoryRepository,
) *smsActivateController {
	return &smsActivateController{
		container:            container,
		profileRepository:    profileRepository,
		smsHistoryRepository: smsHistoryRepository,
		telegramBotService:   service.NewTelegramBot(container),
		formatterWorker:      worker.NewFormatter(container),
	}
}

func (s *smsActivateController) Serve(update *sms.WebhookUpdates) error {
	ctx := context.Background()
	log := s.container.GetLogger()
	domainSMSHistory, err := s.smsHistoryRepository.GetByActivationID(ctx, update.ActivationID)
	if err != nil {
		log.Error("fail to get sms history from db", logger.FError(err))
		return err
	}
	domainSMSHistory.SMSText = utils.NewString(update.Text)
	domainSMSHistory.SMSCode = utils.NewString(update.Code)
	if err := s.smsHistoryRepository.ReceiveSMSCode(ctx, domainSMSHistory); err != nil {
		log.Error("fail to get sms history from db", logger.FError(err))
		return err
	}
	domainProfile, err := s.profileRepository.FetchByID(ctx, domainSMSHistory.ProfileID)
	if err != nil {
		log.Error("fail to get fetch profile from db by id", logger.FError(err))
		return err
	}
	langCode := *domainProfile.PreferredLanguage
	replyKeyboardRemove := telegram.ReplyKeyboardRemove{
		RemoveKeyboard: true,
	}
	respText := s.formatterWorker.CompleteSMSActivation(langCode, domainSMSHistory)
	sendPhoto := telegram.SendPhoto{
		ChatID:      domainProfile.TelegramChatID,
		Photo:       successReceivedCodeImageURL,
		Caption:     respText,
		ReplyMarkup: replyKeyboardRemove,
	}
	if err := s.telegramBotService.SendResponse(sendPhoto, app.SendPhotoTelegramMethod); err != nil {
		log.Error("send code to telegram chat has failed", logger.FError(err))
		return err
	}
	return nil
}
