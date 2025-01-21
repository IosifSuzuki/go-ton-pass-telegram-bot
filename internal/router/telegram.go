package router

import (
	"go-ton-pass-telegram-bot/internal/container"
	telegramController "go-ton-pass-telegram-bot/internal/controller/telegram"
	"go-ton-pass-telegram-bot/internal/manager"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/domain"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/worker"
	"go-ton-pass-telegram-bot/pkg/logger"
	"net/http"
)

type TelegramRouter struct {
	container          container.Container
	exchangeRateWorker worker.ExchangeRate
	controller         telegramController.BotController
}

func NewTelegramRouter(
	container container.Container,
	telegramBotController telegramController.BotController,
	exchangeRateWorker worker.ExchangeRate,
) *TelegramRouter {
	return &TelegramRouter{
		container:          container,
		exchangeRateWorker: exchangeRateWorker,
		controller:         telegramBotController,
	}
}

func (t *TelegramRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := t.container.GetLogger()
	defer func() {
		w.WriteHeader(http.StatusOK)
	}()

	update := r.Context().Value(app.UpdateContextKey).(*telegram.Update)
	profile := r.Context().Value(app.ProfileContextKey).(*domain.Profile)
	isMemberSubscription := r.Context().Value(app.IsProfileSubscription).(bool)

	var languageTag = "en"
	if profile.PreferredLanguage != nil {
		languageTag = *profile.PreferredLanguage
	}
	telegramInlineKeyboardManager := manager.NewTelegramInlineKeyboardManager(
		t.container,
		t.exchangeRateWorker,
	)
	telegramInlineKeyboardManager.Set(languageTag)
	ctxOptions := telegramController.ContextOptions{
		TelegramInlineKeyboardManager: telegramInlineKeyboardManager,
		Update:                        update,
		Profile:                       profile,
		IsMemberSubscription:          isMemberSubscription,
	}
	err := t.controller.Serve(&ctxOptions)
	if err != nil {
		log.Fatal("fail to processing message from bot", logger.FError(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
