package middleware

import (
	"context"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/domain"
	"go-ton-pass-telegram-bot/internal/model/telegram"
	"go-ton-pass-telegram-bot/internal/repository"
	"go-ton-pass-telegram-bot/pkg/logger"
	"net/http"
)

type Authentication struct {
	container         container.Container
	profileRepository repository.ProfileRepository
}

func NewAuthentication(container container.Container, profileRepository repository.ProfileRepository) Authentication {
	return Authentication{
		container:         container,
		profileRepository: profileRepository,
	}
}

func (a *Authentication) Handler(next http.Handler) http.Handler {
	log := a.container.GetLogger()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		update := r.Context().Value(app.UpdateContextKey).(*telegram.Update)
		telegramUser, err := getTelegramUser(update)
		if err != nil {
			log.Error("fail to get telegram account from telegram's update", logger.FError(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		profileExist, err := a.profileRepository.ExistsWithTelegramID(ctx, telegramUser.ID)
		if err != nil {
			log.Error(
				"fail to check an existing telegram account in db with provided telegram id",
				logger.F("telegram_id", telegramUser.ID),
				logger.FError(err),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !profileExist && update.PreCheckoutQuery != nil {
			err = app.UserNotFoundError
			log.Error(
				"profile is not exist in db with pre checkout query",
				logger.F("telegram_id", update.PreCheckoutQuery.From.ID),
				logger.FError(err),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if !profileExist {
			log.Debug(
				"record the profile to db",
				logger.F("telegram_id", telegramUser.ID),
			)
			profile := &domain.Profile{
				TelegramID:     telegramUser.ID,
				TelegramChatID: update.GetChatID(),
				Username:       telegramUser.Username,
			}
			_, err := a.profileRepository.Create(ctx, profile)
			if err != nil {
				log.Error(
					"fail to record the profile to db",
					logger.F("telegram_id", telegramUser.ID),
				)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		profile, err := a.profileRepository.FetchByTelegramID(ctx, telegramUser.ID)
		if err != nil {
			log.Error(
				"fetch a profile by telegram_id",
				logger.F("telegram_id", telegramUser.ID),
			)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		newCtx := context.WithValue(r.Context(), app.ProfileContextKey, profile)
		next.ServeHTTP(w, r.WithContext(newCtx))
	})
}

func getTelegramUser(update *telegram.Update) (*telegram.User, error) {
	if update.Message != nil {
		return update.Message.From, nil
	} else if update.CallbackQuery != nil {
		return &update.CallbackQuery.From, nil
	} else if update.PreCheckoutQuery != nil {
		return &update.PreCheckoutQuery.From, nil
	}
	return nil, app.NilError
}
