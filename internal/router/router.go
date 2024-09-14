package router

import (
	"github.com/gorilla/mux"
	"go-ton-pass-telegram-bot/internal/container"
	"net/http"
)

func PrepareAndConfigureRouter(container container.Container) http.Handler {
	router := mux.NewRouter()
	telegramRouter := NewTelegramRouter(container)
	router.Handle("/telegram/handler/webhook", telegramRouter)
	return router
}
