package service

import (
	"context"
	"github.com/redis/go-redis/v9"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/sms"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
	"time"
)

type Cache interface {
	SaveExchangeRate(ctx context.Context, exchangeRates []app.ExchangeRate) error
	GetExchangeRate(ctx context.Context) (*app.CacheResponse[[]app.ExchangeRate], error)
	SaveSMSCountries(ctx context.Context, services []sms.Country) error
	GetSMSCountries(ctx context.Context) (*app.CacheResponse[[]sms.Country], error)
}

const (
	exchangeRateCacheKey = "exchangeRateCacheKey"
	smsCountriesCacheKey = "smsCountriesCacheKey"
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
	log := c.container.GetLogger()
	log.Debug("will to save exchange rate")
	var cacheResponse app.CacheResponse[[]app.ExchangeRate]
	cacheResponse.TimeFetched = time.Now()
	cacheResponse.Result = exchangeRates
	encodedData, err := utils.EncodePayload(&cacheResponse)
	if err != nil {
		log.Debug("fail to encode payload", logger.FError(err))
		return err
	}
	return c.client.Set(ctx, exchangeRateCacheKey, encodedData, 0).Err()
}

func (c *cache) GetExchangeRate(ctx context.Context) (*app.CacheResponse[[]app.ExchangeRate], error) {
	log := c.container.GetLogger()
	log.Debug("will get exchange rate")
	encodedText, err := c.client.Get(ctx, exchangeRateCacheKey).Result()
	if err != nil {
		log.Debug("fail to get exchangeRate from cache", logger.FError(err))
		return nil, err
	}
	var exchangeRates app.CacheResponse[[]app.ExchangeRate]
	if err := utils.DecodePayload(encodedText, &exchangeRates); err != nil {
		return nil, err
	}
	return &exchangeRates, nil
}

func (c *cache) SaveSMSCountries(ctx context.Context, countries []sms.Country) error {
	log := c.container.GetLogger()
	log.Debug("will to save sms countries rate")
	var cacheResponse app.CacheResponse[[]sms.Country]
	cacheResponse.Result = countries
	cacheResponse.TimeFetched = time.Now()
	encodedData, err := utils.EncodePayload(&cacheResponse)
	if err != nil {
		log.Debug("fail to encode payload", logger.FError(err))
		return err
	}
	return c.client.Set(ctx, smsCountriesCacheKey, encodedData, 0).Err()
}

func (c *cache) GetSMSCountries(ctx context.Context) (*app.CacheResponse[[]sms.Country], error) {
	log := c.container.GetLogger()
	log.Debug("will get sms countries")
	encodedText, err := c.client.Get(ctx, smsCountriesCacheKey).Result()
	if err != nil {
		log.Debug("fail to get sms countries from cache", logger.FError(err))
		return nil, err
	}
	var cacheResponse app.CacheResponse[[]sms.Country]
	if err := utils.DecodePayload(encodedText, &cacheResponse); err != nil {
		return nil, err
	}
	return &cacheResponse, nil
}
