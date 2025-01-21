package service

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
	"time"
)

type SessionService interface {
	ClearBotStateForUser(ctx context.Context, userID int64) error
	GetBotStateForUser(ctx context.Context, userID int64) app.BotState
	SaveBotStateForUser(ctx context.Context, botState app.BotState, userID int64) error
	SaveString(ctx context.Context, key string, value string, userID int64) error
	GetString(ctx context.Context, key string, userID int64) (*string, error)
	ClearString(ctx context.Context, key string, userID int64) error
}

const (
	SelectedPayCurrencyAbbrSessionKey = "selected_pay_currency"
)

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
	key := keyForBotState("bot_state", userID)
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
	key := keyForBotState("bot_state", userID)
	ttl := 5 * time.Minute
	log.Debug("will save bot state", logger.F("userID", userID), logger.F("bot state", botState))
	return s.client.Set(ctx, key, int64(botState), ttl).Err()
}

func (s *sessionService) ClearBotStateForUser(ctx context.Context, userID int64) error {
	log := s.container.GetLogger()
	log.Debug("will clear bot state", logger.F("userID", userID))
	key := keyForBotState("bot_state", userID)
	return s.client.Del(ctx, key).Err()
}

func (s *sessionService) SaveString(ctx context.Context, key string, value string, userID int64) error {
	log := s.container.GetLogger()
	log.Debug("will save string", logger.F("userID", userID), logger.F("value", value), logger.F("key", key))
	transformedKey := keyForBotState(key, userID)
	ttl := 60 * time.Minute
	return s.client.Set(ctx, transformedKey, value, ttl).Err()
}

func (s *sessionService) GetString(ctx context.Context, key string, userID int64) (*string, error) {
	log := s.container.GetLogger()
	transformedKey := keyForBotState(key, userID)
	value, err := s.client.Get(ctx, transformedKey).Result()
	log.Debug("get string", logger.F("userID", userID), logger.F("value", value), logger.F("key", key))
	if err != nil {
		return nil, err
	} else if len(value) == 0 {
		return nil, app.NilError
	}
	return utils.NewString(value), nil
}

func (s *sessionService) ClearString(ctx context.Context, key string, userID int64) error {
	log := s.container.GetLogger()
	log.Debug("will clear bot state", logger.F("userID", userID))
	transformedKey := keyForBotState(key, userID)
	return s.client.Del(ctx, transformedKey).Err()
}

func keyForBotState(key string, userID int64) string {
	return fmt.Sprintf("%s/%d", key, userID)
}
