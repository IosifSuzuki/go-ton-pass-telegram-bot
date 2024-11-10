package postpone

import (
	"context"
	"go-ton-pass-telegram-bot/internal/container"
	model "go-ton-pass-telegram-bot/internal/model/postpone"
	"go-ton-pass-telegram-bot/internal/repository"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/service/postpone/workflow"
	"go-ton-pass-telegram-bot/pkg/logger"
	"go.temporal.io/sdk/client"
)

type Postpone interface {
	ScheduleCheckSMSActivation(ctx context.Context, telegramID int64, activationID int64, amount float64) (*model.Workflow, error)
	CancelSMSActivation(ctx context.Context, workflow model.Workflow) error
	Prepare() error
}

type postpone struct {
	container            container.Container
	smsWorker            workflow.SMSActivateWorker
	profileRepository    repository.ProfileRepository
	smsHistoryRepository repository.SMSHistoryRepository
}

func NewPostpone(
	container container.Container,
	client client.Client,
	profileRepository repository.ProfileRepository,
	smsHistoryRepository repository.SMSHistoryRepository,
) Postpone {
	telegramService := service.NewTelegramBot(container)
	smsService := service.NewSMSService(container)
	smsWorker := workflow.NewSMSActivateWorker(container, client, telegramService, smsService, profileRepository, smsHistoryRepository)
	return &postpone{
		container:            container,
		smsWorker:            smsWorker,
		profileRepository:    profileRepository,
		smsHistoryRepository: smsHistoryRepository,
	}
}

func (p *postpone) ScheduleCheckSMSActivation(ctx context.Context, telegramID int64, activationID int64, amount float64) (*model.Workflow, error) {
	log := p.container.GetLogger()
	profile, err := p.profileRepository.FetchByTelegramID(ctx, telegramID)
	if err != nil {
		log.Error("fail to fetch profile", logger.F("telegram_id", telegramID), logger.FError(err))
		return nil, err
	}
	input := model.SMSActivation{
		ActivationID: activationID,
		ProfileID:    profile.ID,
		ChatID:       profile.TelegramChatID,
		Amount:       amount,
	}
	return p.smsWorker.AddToQueue(ctx, input)
}

func (p *postpone) CancelSMSActivation(ctx context.Context, workflow model.Workflow) error {
	return p.smsWorker.ExecuteCancel(ctx, workflow)
}

func (p *postpone) Prepare() error {
	p.smsWorker.Prepare()
	return nil
}
