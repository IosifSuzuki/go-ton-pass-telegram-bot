package router

import (
	"github.com/gorilla/mux"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/service/telegramBot"
	"net/http"
)

func PrepareAndConfigureRouter(container container.Container, telegramBotService telegramBot.Bot) http.Handler {
	router := mux.NewRouter()
	telegramRouter := NewTelegramRouter(container, telegramBotService)
	router.Handle("/telegram/handler/webhook", telegramRouter)
	return router
}
