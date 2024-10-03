package router

import (
	"encoding/json"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/controller/crypto"
	"go-ton-pass-telegram-bot/internal/model/crypto/bot"
	"go-ton-pass-telegram-bot/pkg/logger"
	"net/http"
)

type CryptoBotRouter struct {
	container  container.Container
	controller crypto.CryptoController
}

func NewCryptoBotRouter(container container.Container, controller crypto.CryptoController) *CryptoBotRouter {
	return &CryptoBotRouter{
		container:  container,
		controller: controller,
	}
}

func (c *CryptoBotRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		w.WriteHeader(http.StatusOK)
	}()
	log := c.container.GetLogger()
	var update bot.WebhookUpdates
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		log.Fatal("fail to decode", logger.FError(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := c.controller.Serve(&update)
	if err != nil {
		log.Fatal("controller has failed", logger.FError(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
