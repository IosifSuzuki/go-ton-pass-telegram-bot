package router

import (
	"github.com/gorilla/mux"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/controller/crypto"
	telegramController "go-ton-pass-telegram-bot/internal/controller/telegram"
	"go-ton-pass-telegram-bot/internal/repository"
	"go-ton-pass-telegram-bot/internal/service"
	"net/http"
)

func PrepareAndConfigureRouter(
	container container.Container,
	sessionService service.SessionService,
	smsService service.SMSService,
	profileRepository repository.ProfileRepository,
	invoiceRepository repository.InvoiceRepository,
) http.Handler {
	router := mux.NewRouter()
	telegramBotController := telegramController.NewBotController(
		container,
		sessionService,
		smsService,
		profileRepository,
		invoiceRepository,
	)
	cryptoController := crypto.NewCryptoController(
		container,
		sessionService,
		profileRepository,
		invoiceRepository,
	)
	telegramRouter := NewTelegramRouter(container, telegramBotController)
	router.Handle("/telegram/handler/webhook", telegramRouter)
	cryptoRouter := NewCryptoBotRouter(container, cryptoController)
	router.Handle("/telegram/crypto_bot/webhook", cryptoRouter)

	return router
}
