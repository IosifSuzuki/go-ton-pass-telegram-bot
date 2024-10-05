package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"time"
)

type Cache interface {
	SaveExchangeRate(ctx context.Context, exchangeRates []app.ExchangeRate) error
	GetExchangeRate(ctx context.Context) (*app.CacheResponse[[]app.ExchangeRate], error)
}

const (
	exchangeRateCacheKey = "exchangeRateCacheKey"
)

type cache struct {
	container container.Container
	client    *redis.Client
}

func NewCache(container container.Container, client *redis.Client) Cache {
	return &cache{
		container: container,
		client:    client,
	}
}

func (c *cache) SaveExchangeRate(ctx context.Context, exchangeRates []app.ExchangeRate) error {
	var cacheResponse app.CacheResponse[[]app.ExchangeRate]
	cacheResponse.TimeFetched = time.Now()
	cacheResponse.Result = exchangeRates
	data, err := json.Marshal(cacheResponse)
	if err != nil {
		return err
	}
	encodedData := base64.StdEncoding.EncodeToString(data)
	return c.client.Set(ctx, exchangeRateCacheKey, encodedData, 0).Err()
}

func (c *cache) GetExchangeRate(ctx context.Context) (*app.CacheResponse[[]app.ExchangeRate], error) {
	encodedText, err := c.client.Get(ctx, exchangeRateCacheKey).Result()
	if err != nil {
		return nil, err
	}
	data, err := base64.StdEncoding.DecodeString(encodedText)
	if err != nil {
		return nil, err
	}
	var exchangeRates app.CacheResponse[[]app.ExchangeRate]
	if err := json.Unmarshal(data, &exchangeRates); err != nil {
		return nil, err
	}
	return &exchangeRates, nil
}
