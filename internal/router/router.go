package router

import (
	"github.com/gorilla/mux"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/service"
	"net/http"
)

func PrepareAndConfigureRouter(container container.Container, sessionService service.SessionService) http.Handler {
	router := mux.NewRouter()
	telegramRouter := NewTelegramRouter(container, sessionService)
	router.Handle("/telegram/handler/webhook", telegramRouter)
	return router
}
