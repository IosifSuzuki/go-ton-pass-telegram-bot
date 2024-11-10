package service

import (
	"context"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
)

const (
	PushOperationCallbackDataStack = "push"
	PopOperationCallbackDataStack  = "pop"
)

type CallbackDataStack interface {
	Push(ctx context.Context, callbackQuery *telegram.CallbackQuery) error
	Pop(ctx context.Context, callbackQuery *telegram.CallbackQuery) (*app.TelegramCallbackData, error)
}
type callbackDataStack struct {
	container container.Container
	cache     Cache
}

func NewCallbackDataStack(container container.Container, cache Cache) CallbackDataStack {
	return &callbackDataStack{
		container: container,
		cache:     cache,
	}
}

func (c *callbackDataStack) Push(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	userID := callbackQuery.Message.From.ID
	chatID := callbackQuery.Message.Chat.ID
	messageID := callbackQuery.Message.ID
	callbackData, _ := c.cache.GetTelegramCallbackData(ctx, userID, chatID, messageID)
	if callbackData == nil {
		callbackData = make([]app.TelegramCallbackData, 0)
	}
	item, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if err != nil {
		return err
	}
	callbackData = append(callbackData, *item)
	if err := c.cache.SetLastTelegramCallbackDataOperation(ctx, PushOperationCallbackDataStack, userID, chatID, messageID); err != nil {
		return err
	}
	return c.cache.SaveTelegramCallbackData(ctx, callbackData, userID, chatID, messageID)
}

func (c *callbackDataStack) Pop(ctx context.Context, callbackQuery *telegram.CallbackQuery) (*app.TelegramCallbackData, error) {
	userID := callbackQuery.Message.From.ID
	chatID := callbackQuery.Message.Chat.ID
	messageID := callbackQuery.Message.ID
	operation, _ := c.cache.GetLastTelegramCallbackDataOperation(ctx, userID, chatID, messageID)
	callbackData, err := c.cache.GetTelegramCallbackData(ctx, userID, chatID, messageID)
	if err != nil {
		return nil, err
	}
	endIndex := len(callbackData) - 1
	if operation == PushOperationCallbackDataStack {
		endIndex -= 1
	}
	if err := c.cache.SetLastTelegramCallbackDataOperation(ctx, PopOperationCallbackDataStack, userID, chatID, messageID); err != nil {
		return nil, err
	}
	if endIndex < 0 {
		return nil, app.NilError
	}
	callbackData = callbackData[:endIndex]
	item := callbackData[len(callbackData)-1]
	if err := c.cache.SaveTelegramCallbackData(ctx, callbackData, userID, chatID, messageID); err != nil {
		return nil, err
	}
	return &item, nil
}
