package router

import (
	"github.com/gorilla/mux"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/controller/crypto"
	"go-ton-pass-telegram-bot/internal/controller/sms"
	telegramController "go-ton-pass-telegram-bot/internal/controller/telegram"
	"go-ton-pass-telegram-bot/internal/repository"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/service/postpone"
	"net/http"
)

func PrepareAndConfigureRouter(
	container container.Container,
	sessionService service.SessionService,
	cacheService service.Cache,
	smsService service.SMSService,
	postponeService postpone.Postpone,
	profileRepository repository.ProfileRepository,
	smsHistoryRepository repository.SMSHistoryRepository,
) http.Handler {
	router := mux.NewRouter()
	telegramBotController := telegramController.NewBotController(
		container,
		sessionService,
		cacheService,
		smsService,
		postponeService,
		profileRepository,
		smsHistoryRepository,
	)
	cryptoController := crypto.NewCryptoController(
		container,
		sessionService,
		profileRepository,
	)
	router.HandleFunc("/ping", PingServe)
	smsActivateController := sms.NewSMSActivateController(container, profileRepository, smsHistoryRepository)
	telegramRouter := NewTelegramRouter(container, telegramBotController)
	router.Handle("/telegram/handler/webhook", telegramRouter)
	cryptoRouter := NewCryptoBotRouter(container, cryptoController)
	router.Handle("/telegram/crypto_bot/webhook", cryptoRouter)
	smsActivateRouter := NewSMSActivateRouter(container, smsActivateController)
	router.Handle("/sms_activate/webhook", smsActivateRouter)

	return router
}

func PingServe(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("OK"))
	w.WriteHeader(http.StatusOK)
}
