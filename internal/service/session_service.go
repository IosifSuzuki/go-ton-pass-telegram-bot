package service

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/pkg/logger"
	"time"
)

type SessionService interface {
	ClearBotStateForUser(ctx context.Context, userID int64) error
	GetBotStateForUser(ctx context.Context, userID int64) app.BotState
	SaveBotStateForUser(ctx context.Context, botState app.BotState, userID int64) error
}

type sessionService struct {
	container container.Container
	client    *redis.Client
}

func NewSessionService(container container.Container, client *redis.Client) SessionService {
	return &sessionService{
		container: container,
		client:    client,
	}
}

func (s *sessionService) GetBotStateForUser(ctx context.Context, userID int64) app.BotState {
	log := s.container.GetLogger()
	key := keyForBotState(userID)
	value, err := s.client.Get(ctx, key).Int64()
	if err != nil {
		log.Debug("get BotState failed", logger.F("key", key))
		return app.IDLEState
	}

	log.Debug("get BotState", logger.F("bot state", app.BotState(value)))
	return app.BotState(value)
}

func (s *sessionService) SaveBotStateForUser(ctx context.Context, botState app.BotState, userID int64) error {
	log := s.container.GetLogger()
	key := keyForBotState(userID)
	ttl := 5 * time.Minute
	log.Debug("will save bot state", logger.F("userID", userID), logger.F("bot state", botState))
	return s.client.Set(ctx, key, int64(botState), ttl).Err()
}

func (s *sessionService) ClearBotStateForUser(ctx context.Context, userID int64) error {
	log := s.container.GetLogger()
	log.Debug("will clear bot state", logger.F("userID", userID))
	key := keyForBotState(userID)
	return s.client.Del(ctx, key).Err()
}

func keyForBotState(userID int64) string {
	return fmt.Sprintf("bot_state/%d", userID)
}
