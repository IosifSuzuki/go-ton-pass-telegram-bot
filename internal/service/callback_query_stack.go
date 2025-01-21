package service

import (
	"context"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
	"strings"
)

type CallbackDataStack interface {
	Push(ctx context.Context, callbackQuery *telegram.CallbackQuery) error
	Pop(ctx context.Context, callbackQuery *telegram.CallbackQuery) (*app.TelegramCallbackData, error)
	Top(ctx context.Context, callbackQuery *telegram.CallbackQuery) (*app.TelegramCallbackData, error)
	DebugListCommands(ctx context.Context, callbackQuery *telegram.CallbackQuery) (*string, error)
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
	log := c.container.GetLogger()
	chatID := callbackQuery.Message.Chat.ID
	messageID := callbackQuery.Message.ID
	telegramMessagingInfo := TelegramMessagingInfo{ChatID: chatID, MessageID: messageID}
	callbackData, _ := c.cache.GetTelegramCallbackData(ctx, telegramMessagingInfo)
	if callbackData == nil {
		callbackData = make([]app.TelegramCallbackData, 0)
	}
	item, err := utils.DecodeTelegramCallbackData(callbackQuery.Data)
	if err != nil {
		return err
	}
	callbackData = append(callbackData, *item)
	log.Debug("callback will be saved after push operation", logger.F("callback_data", callbackData))
	return c.cache.SaveTelegramCallbackData(ctx, callbackData, telegramMessagingInfo)
}

func (c *callbackDataStack) Pop(ctx context.Context, callbackQuery *telegram.CallbackQuery) (*app.TelegramCallbackData, error) {
	log := c.container.GetLogger()
	chatID := callbackQuery.Message.Chat.ID
	messageID := callbackQuery.Message.ID
	telegramMessagingInfo := TelegramMessagingInfo{ChatID: chatID, MessageID: messageID}
	callbackData, err := c.cache.GetTelegramCallbackData(ctx, telegramMessagingInfo)
	if err != nil {
		return nil, err
	}
	var lastItem *app.TelegramCallbackData
	if len(callbackData) > 0 {
		lastItem = &callbackData[len(callbackData)-1]
		callbackData = callbackData[:len(callbackData)-1]
		log.Debug("callback will be saved after pop operation", logger.F("callback_data", callbackData))
		if err := c.cache.SaveTelegramCallbackData(ctx, callbackData, telegramMessagingInfo); err != nil {
			return nil, err
		}
	}
	if lastItem == nil {
		return nil, app.NilError
	}
	return lastItem, nil
}

func (c *callbackDataStack) Top(ctx context.Context, callbackQuery *telegram.CallbackQuery) (*app.TelegramCallbackData, error) {
	chatID := callbackQuery.Message.Chat.ID
	messageID := callbackQuery.Message.ID
	telegramMessagingInfo := TelegramMessagingInfo{ChatID: chatID, MessageID: messageID}
	callbackData, err := c.cache.GetTelegramCallbackData(ctx, telegramMessagingInfo)
	if err != nil {
		return nil, err
	}
	if len(callbackData) > 0 {
		topItem := &callbackData[len(callbackData)-1]
		return topItem, err
	}
	return nil, app.NilError
}

func (c *callbackDataStack) DebugListCommands(ctx context.Context, callbackQuery *telegram.CallbackQuery) (*string, error) {
	log := c.container.GetLogger()
	chatID := callbackQuery.Message.Chat.ID
	messageID := callbackQuery.Message.ID
	telegramMessagingInfo := TelegramMessagingInfo{ChatID: chatID, MessageID: messageID}
	callbackData, err := c.cache.GetTelegramCallbackData(ctx, telegramMessagingInfo)
	if err != nil {
		return nil, err
	}
	var names = make([]string, 0, len(callbackData))
	for _, item := range callbackData {
		names = append(names, item.Name)
	}
	log.Debug("list commands in cache", logger.F("list", names))
	var namesText = strings.Join(names, "/")
	return &namesText, nil
}
