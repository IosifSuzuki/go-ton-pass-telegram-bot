package workflow

import (
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/postpone"
	"go-ton-pass-telegram-bot/internal/service/postpone/workflow/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"time"
)

func SMSActivateStatusWorkflow(ctx workflow.Context, input postpone.SMSActivation) (string, error) {
	successMsg := "success complete operation"
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    100 * time.Second,
		MaximumAttempts:    500,
	}
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, options)
	var a *activity.SMSActivity
	var result string
	if err := workflow.ExecuteActivity(ctx, a.GetStatus, input.ActivationID).Get(ctx, &result); err != nil {
		return "", err
	}
	activationStatus := app.SMSActivationState(result)
	if err := workflow.ExecuteActivity(ctx, a.SaveStatusInDB, input.ActivationID, activationStatus).Get(ctx, nil); err != nil {
		return "", err
	}
	if activationStatus == app.DoneSMSActivateState {
		return successMsg, nil
	}
	if activationStatus == app.PendingSMSActivateState {
		if err := workflow.ExecuteActivity(ctx, a.CancelStatus, input.ActivationID).Get(ctx, nil); err != nil {
			return "", err
		}
		activationStatus = app.CancelSMSActivateState
		if err := workflow.ExecuteActivity(ctx, a.SaveStatusInDB, input.ActivationID, activationStatus).Get(ctx, nil); err != nil {
			return "", err
		}
	}
	if err := workflow.ExecuteActivity(ctx, a.RefundAmount, input.ProfileID, input.Amount).Get(ctx, nil); err != nil {
		return "", err
	}
	if err := workflow.ExecuteActivity(ctx, a.RefundTimeOutMessage, input.ChatID, input.ProfileID, input.ActivationID).Get(ctx, nil); err != nil {
		return "", err
	}
	return successMsg, nil
}

func CancelSMSActivateStatusWorkflow(ctx workflow.Context, input postpone.SMSActivation) (string, error) {
	successMsg := "success cancel operation"
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    100 * time.Second,
		MaximumAttempts:    500,
	}
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, options)
	var a *activity.SMSActivity
	if err := workflow.ExecuteActivity(ctx, a.SaveStatusInDB, input.ActivationID, app.CancelSMSActivateState).Get(ctx, nil); err != nil {
		return "", err
	}
	if err := workflow.ExecuteActivity(ctx, a.RefundAmount, input.ProfileID, input.Amount).Get(ctx, nil); err != nil {
		return "", err
	}
	if err := workflow.ExecuteActivity(ctx, a.UserRefundMessage, input.ChatID, input.ProfileID, input.ActivationID).Get(ctx, nil); err != nil {
		return "", err
	}
	return successMsg, nil
}
