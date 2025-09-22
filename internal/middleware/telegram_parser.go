package middleware

import (
	"context"
	"encoding/json"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/pkg/logger"
	"net/http"
)

type TelegramParser struct {
	container container.Container
}

func NewTelegramParser(container container.Container) *TelegramParser {
	return &TelegramParser{
		container: container,
	}
}

func (t *TelegramParser) Handler(next http.Handler) http.Handler {
	log := t.container.GetLogger()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var update telegram.Update
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			log.Fatal("fail to decode", logger.FError(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if isEmpty(&update) {
			log.Error("update is empty")
			w.WriteHeader(http.StatusOK)
			return
		}
		newCtx := context.WithValue(r.Context(), app.UpdateContextKey, &update)
		next.ServeHTTP(w, r.WithContext(newCtx))
	})
}

func isEmpty(update *telegram.Update) bool {
	return update.Message == nil && update.CallbackQuery == nil && update.PreCheckoutQuery == nil
}
