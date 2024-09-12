package router

import (
	"go-ton-pass-telegram-bot/internal/container"
	"net/http"
)

type TelegramRouter struct {
	container container.Container
}

func NewTelegramRouter(container container.Container) *TelegramRouter {
	return &TelegramRouter{
		container: container,
	}
}

func (t *TelegramRouter) WebHookHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
