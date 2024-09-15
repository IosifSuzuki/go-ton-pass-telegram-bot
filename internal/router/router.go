package router

import (
	"github.com/gorilla/mux"
	"go-ton-pass-telegram-bot/internal/container"
	telegramController "go-ton-pass-telegram-bot/internal/controller/telegram"
	"go-ton-pass-telegram-bot/internal/repository"
	"go-ton-pass-telegram-bot/internal/service"
	"net/http"
)

func PrepareAndConfigureRouter(
	container container.Container,
	sessionService service.SessionService,
	profileRepository repository.ProfileRepository,
) http.Handler {
	router := mux.NewRouter()
	telegramBotController := telegramController.NewBotController(container, sessionService, profileRepository)
	telegramRouter := NewTelegramRouter(container, telegramBotController)
	router.Handle("/telegram/handler/webhook", telegramRouter)
	return router
}
