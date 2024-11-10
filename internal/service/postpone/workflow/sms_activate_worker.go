package workflow

import (
	"context"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/postpone"
	"go-ton-pass-telegram-bot/internal/repository"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/service/postpone/workflow/activity"
	"go-ton-pass-telegram-bot/pkg/logger"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/worker"
	"time"
)

const SMSActivateQueueName = "sms_activate"

type SMSActivateWorker interface {
	AddToQueue(ctx context.Context, smsActivation postpone.SMSActivation) (*postpone.Workflow, error)
	ExecuteCancel(ctx context.Context, workflow postpone.Workflow) error
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
	w.RegisterWorkflow(CancelSMSActivateStatusWorkflow)
	w.RegisterActivity(s.activity)
	go func() {
		_ = w.Run(worker.InterruptCh())
	}()
}

func (s *smsActivateWorker) AddToQueue(ctx context.Context, smsActivation postpone.SMSActivation) (*postpone.Workflow, error) {
	log := s.container.GetLogger()
	startWorkflowOptions := client.StartWorkflowOptions{
		StartDelay: 20 * time.Minute,
		TaskQueue:  SMSActivateQueueName,
	}
	workflowRun, err := s.client.ExecuteWorkflow(ctx, startWorkflowOptions, SMSActivateStatusWorkflow, smsActivation)
	if err != nil {
		return nil, err
	}
	workflow := postpone.Workflow{
		ID:    workflowRun.GetID(),
		RunID: workflowRun.GetRunID(),
	}
	log.Debug("success prepare workflow to execute", logger.F("workflow", workflow))
	return &workflow, nil
}

func (s *smsActivateWorker) ExecuteCancel(ctx context.Context, workflow postpone.Workflow) error {
	smsActivation, err := s.getWorkflowSMSActivation(workflow.ID, workflow.RunID)
	if err != nil {
		return err
	}
	if err := s.client.CancelWorkflow(ctx, workflow.ID, workflow.RunID); err != nil {
		return err
	}
	startWorkflowOptions := client.StartWorkflowOptions{
		TaskQueue: SMSActivateQueueName,
	}
	_, err = s.client.ExecuteWorkflow(ctx, startWorkflowOptions, CancelSMSActivateStatusWorkflow, *smsActivation)
	return err
}

func (s *smsActivateWorker) getWorkflowSMSActivation(id string, runID string) (*postpone.SMSActivation, error) {
	historyIterator := s.client.GetWorkflowHistory(context.Background(), id, runID, false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	var input = new(postpone.SMSActivation)
	dataConverter := converter.GetDefaultDataConverter()
	for historyIterator.HasNext() {
		event, err := historyIterator.Next()
		if err != nil {
			return nil, err
		}

		if event.GetEventType() == enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED {
			attributes := event.GetWorkflowExecutionStartedEventAttributes()
			err := dataConverter.FromPayloads(attributes.Input, &input)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	return input, nil
}
