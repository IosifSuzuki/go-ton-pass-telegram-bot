package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/service/telegramBot"
	"go-ton-pass-telegram-bot/pkg/logger"
	"io"
	"net/http"
)

type TelegramRouter struct {
	container   container.Container
	telegramBot telegramBot.Bot
}

func NewTelegramRouter(container container.Container, telegramBot telegramBot.Bot) *TelegramRouter {
	return &TelegramRouter{
		container:   container,
		telegramBot: telegramBot,
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
	resp, err := t.telegramBot.Processing(&update)
	if err != nil {
		log.Fatal("fail to processing message from bot", logger.FError(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := t.sendResponseToTelegramServer(resp); err != nil {
		log.Fatal("fail to send message from bot", logger.FError(err))
	}
}

func (t *TelegramRouter) sendResponseToTelegramServer(model *telegram.SendResponse) error {
	log := t.container.GetLogger()
	const baseTelegramAPI = "https://api.telegram.org/bot"
	path := fmt.Sprintf("%s%s/%s", baseTelegramAPI, t.container.GetTelegramBotToken(), "sendMessage")
	body, err := json.Marshal(model)
	if err != nil {
		log.Error("fail to encode telegram message", logger.FError(err))
		return err
	}
	bodyBuffer := bytes.NewBuffer(body)
	resp, err := http.Post(path, "application/json", bodyBuffer)
	if err != nil {
		log.Error("fail to send data to telegram server", logger.FError(err))
		return err
	}
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Error("fail to read body from telegram server", logger.FError(err))
		return err
	}
	var result *telegram.Result
	if err := json.Unmarshal(body, &result); err != nil {
		log.Error("fail to decode body from telegram server", logger.FError(err))
		return err
	}
	if !result.OK {
		log.Error("telegram server return without status code ok", logger.F("description", result.Description))
		return app.TelegramResponseBotError
	}
	return nil
}
