package workflow

import (
	"context"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/postpone"
	"go-ton-pass-telegram-bot/internal/repository"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/service/postpone/workflow/activity"
	"go-ton-pass-telegram-bot/pkg/logger"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"time"
)

const SMSActivateQueueName = "sms_activate"

type SMSActivateWorker interface {
	AddToQueue(ctx context.Context, smsActivation postpone.SMSActivation) error
	Prepare()
}
type smsActivateWorker struct {
	container container.Container
	client    client.Client
	activity  *activity.SMSActivity
}

func NewSMSActivateWorker(
	container container.Container,
	client client.Client,
	telegramService service.TelegramBotService,
	smsService service.SMSService,
	profileRepository repository.ProfileRepository,
	smsHistoryRepository repository.SMSHistoryRepository,
) SMSActivateWorker {
	a := activity.NewSMSActivity(container, telegramService, smsService, profileRepository, smsHistoryRepository)
	w := smsActivateWorker{
		container: container,
		client:    client,
		activity:  a,
	}
	return &w
}

func (s *smsActivateWorker) Prepare() {
	w := worker.New(s.client, SMSActivateQueueName, worker.Options{})
	w.RegisterWorkflow(SMSActivateStatusWorkflow)
	w.RegisterActivity(s.activity)
	go func() {
		_ = w.Run(worker.InterruptCh())
	}()
}

func (s *smsActivateWorker) AddToQueue(ctx context.Context, smsActivation postpone.SMSActivation) error {
	log := s.container.GetLogger()
	startWorkflowOptions := client.StartWorkflowOptions{
		StartDelay: 3 * time.Minute,
		TaskQueue:  SMSActivateQueueName,
	}
	workflowRun, err := s.client.ExecuteWorkflow(ctx, startWorkflowOptions, SMSActivateStatusWorkflow, smsActivation)
	if err != nil {
		return err
	}
	log.Debug(
		"success prepare workflow to execute",
		logger.F("id", workflowRun.GetID()),
		logger.F("ru_id", workflowRun.GetRunID()),
	)
	return nil
}
