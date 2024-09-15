package telegram

import (
	"context"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
)

func (b *botController) balanceCallbackQueryCommandHandler(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	log := b.container.GetLogger()
	log.Debug("execute balanceCallbackQueryCommandHandler", logger.F("callbackQuery", callbackQuery))
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      utils.NewString("Your balance has 0 point"),
		ShowAlert: true,
		CacheTime: 10,
	}
	return b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod)
}

func (b *botController) unsupportedCallbackQueryCommandHandle(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	answerCallbackQuery := telegram.AnswerCallbackQuery{
		ID:        callbackQuery.ID,
		Text:      nil,
		ShowAlert: false,
		CacheTime: 10,
	}
	return b.telegramBotService.SendResponse(answerCallbackQuery, app.AnswerCallbackQueryTelegramMethod)
}
