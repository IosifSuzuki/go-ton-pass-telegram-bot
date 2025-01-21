package middleware

import (
	"context"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/domain"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/pkg/logger"
	"net/http"
)

type Subscription struct {
	container          container.Container
	telegramBotService service.TelegramBotService
}

func NewSubscription(
	container container.Container,
	telegramBotService service.TelegramBotService,
) *Subscription {
	return &Subscription{
		container:          container,
		telegramBotService: telegramBotService,
	}
}

func (s *Subscription) Handler(next http.Handler) http.Handler {
	log := s.container.GetLogger()
	channelLink := "@tonpassnews"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		profile := r.Context().Value(app.ProfileContextKey).(*domain.Profile)
		isChatMember, err := s.telegramBotService.UserIsChatMember(channelLink, profile.TelegramID)
		if err != nil {
			log.Error("fail to check is user member of chat", logger.FError(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		newCtx := context.WithValue(r.Context(), app.IsProfileSubscription, isChatMember)
		next.ServeHTTP(w, r.WithContext(newCtx))
	})
}
