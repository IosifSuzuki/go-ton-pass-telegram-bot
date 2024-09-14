package router

import (
	"encoding/json"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/controller"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/pkg/logger"
	"net/http"
)

type TelegramRouter struct {
	container  container.Container
	controller controller.TelegramBotController
}

func NewTelegramRouter(container container.Container, sessionService service.SessionService) *TelegramRouter {
	return &TelegramRouter{
		container:  container,
		controller: controller.NewTelegramBotController(container, sessionService),
	}
}

func (t *TelegramRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := t.container.GetLogger()
	var update telegram.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Fatal("fail to decode", logger.FError(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := t.controller.Serve(&update)
	if err != nil {
		log.Fatal("fail to processing message from bot", logger.FError(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
