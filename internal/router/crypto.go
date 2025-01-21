package router

import (
	"encoding/json"
	"errors"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/controller/crypto"
	"go-ton-pass-telegram-bot/internal/model/crypto/bot"
	"go-ton-pass-telegram-bot/pkg/logger"
	"io"
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
	log.Debug("receive message from webhook")
	var update bot.WebhookUpdates
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		if err := filterCryptoErrors(err); err != nil {
			log.Fatal("fail to decode", logger.FError(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	if err := c.controller.Serve(&update); err != nil {
		log.Fatal("controller has failed", logger.FError(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func filterCryptoErrors(err error) error {
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}
