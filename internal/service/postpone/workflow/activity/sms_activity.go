package activity

import (
	"context"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/repository"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/internal/worker"
	"go-ton-pass-telegram-bot/pkg/logger"
)

const (
	failReceivedCodeImageURL = "https://www.imghippo.com/i/D8eey1728514326.png"
)

type SMSActivity struct {
	container            container.Container
	telegramService      service.TelegramBotService
	smsService           service.SMSService
	profileRepository    repository.ProfileRepository
	smsHistoryRepository repository.SMSHistoryRepository
	formatterWorker      worker.Formatter
}

func NewSMSActivity(
	container container.Container,
	telegramService service.TelegramBotService,
	smsService service.SMSService,
	profileRepository repository.ProfileRepository,
	smsHistoryRepository repository.SMSHistoryRepository,
) *SMSActivity {
	return &SMSActivity{
		container:            container,
		telegramService:      telegramService,
		smsService:           smsService,
		profileRepository:    profileRepository,
		smsHistoryRepository: smsHistoryRepository,
		formatterWorker:      worker.NewFormatter(container),
	}
}

func (s *SMSActivity) GetStatus(_ context.Context, activityID int64) (string, error) {
	status, err := s.smsService.GetStatus(activityID)
	if err != nil {
		return string(app.UnknownSMSActivateState), err
	}
	return string(status), nil
}

func (s *SMSActivity) CancelStatus(_ context.Context, activationID int64) error {
	log := s.container.GetLogger()
	if err := s.smsService.CancelActivation(activationID); err != nil {
		log.Error("cancel activation has failed", logger.FError(err))
		return err
	}
	return nil
}

func (s *SMSActivity) SaveStatusInDB(ctx context.Context, activationID int64, activationStatus app.SMSActivationState) (string, error) {
	return "", s.smsHistoryRepository.ChangeActivationStatus(ctx, activationID, string(activationStatus))
}

func (s *SMSActivity) RefundAmount(ctx context.Context, profileID int64, amount float64) (string, error) {
	log := s.container.GetLogger()
	log.Debug("will refund amount",
		logger.F("profile_id", profileID),
		logger.F("amount", amount),
	)
	if err := s.profileRepository.TopUpBalanceByProfileID(ctx, profileID, amount); err != nil {
		log.Debug("fail to top up balance", logger.F("profile_id", profileID), logger.F("amount", amount))
		return "", err
	}
	return "", nil
}

func (s *SMSActivity) RefundTimeOutMessage(ctx context.Context, chatID int64, profileID int64, activationID int64) (string, error) {
	log := s.container.GetLogger()
	profile, err := s.profileRepository.FetchByID(ctx, profileID)
	if err != nil {
		log.Debug("fail to get profile by id", logger.F("profile_id", profileID))
		return "", err
	}
	langCode := profile.PreferredLanguage
	smsHistory, err := s.smsHistoryRepository.GetByActivationID(ctx, activationID)
	if err != nil {
		log.Debug("fail to get sms history by id", logger.F("activation_id", activationID))
		return "", err
	}
	respText := s.formatterWorker.FailSMSActivation(*langCode, smsHistory)
	replyKeyboardRemove := telegram.ReplyKeyboardRemove{RemoveKeyboard: true}
	sendPhoto := telegram.SendPhoto{
		ChatID:      chatID,
		Photo:       failReceivedCodeImageURL,
		Caption:     respText,
		ReplyMarkup: replyKeyboardRemove,
		ParseMode:   utils.NewString("MarkdownV2"),
	}
	return "", s.telegramService.SendResponse(sendPhoto, app.SendPhotoTelegramMethod)
}

func (s *SMSActivity) UserRefundMessage(ctx context.Context, chatID int64, profileID int64, activationID int64) (string, error) {
	log := s.container.GetLogger()
	profile, err := s.profileRepository.FetchByID(ctx, profileID)
	if err != nil {
		log.Debug("fail to get profile by id", logger.F("profile_id", profileID))
		return "", err
	}
	langCode := profile.PreferredLanguage
	smsHistory, err := s.smsHistoryRepository.GetByActivationID(ctx, activationID)
	if err != nil {
		log.Debug("fail to get sms history by id", logger.F("activation_id", activationID))
		return "", err
	}
	respText := s.formatterWorker.ManualCancelActivation(*langCode, smsHistory)
	replyKeyboardRemove := telegram.ReplyKeyboardRemove{RemoveKeyboard: true}
	sendPhoto := telegram.SendPhoto{
		ChatID:      chatID,
		Photo:       failReceivedCodeImageURL,
		Caption:     respText,
		ReplyMarkup: replyKeyboardRemove,
		ParseMode:   utils.NewString("MarkdownV2"),
	}
	return "", s.telegramService.SendResponse(sendPhoto, app.SendPhotoTelegramMethod)
}
