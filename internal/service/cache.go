package service

import (
	"context"
	"fmt"
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
	SaveSMSServices(ctx context.Context, services []sms.Service) error
	GetSMSServices(ctx context.Context) (*app.CacheResponse[[]sms.Service], error)
	SaveTelegramCallbackData(ctx context.Context, callbackData []app.TelegramCallbackData, userID, chatID, messageID int64) error
	GetTelegramCallbackData(ctx context.Context, userID, chatID, messageID int64) ([]app.TelegramCallbackData, error)
	GetLastTelegramCallbackDataOperation(ctx context.Context, userID, chatID, messageID int64) (string, error)
	SetLastTelegramCallbackDataOperation(ctx context.Context, operation string, userID, chatID, messageID int64) error
}

const (
	exchangeRateCacheKey              = "exchangeRateCacheKey"
	smsCountriesCacheKey              = "smsCountriesCacheKey"
	smsServicesCacheKey               = "smsServicesCacheKey"
	telegramCallbackDataCacheKey      = "telegramCallbackDataCacheKey"
	lastOperationCallbackDataCacheKey = "lastOperationCallbackDataCacheKey"
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
		log.Debug("fail to decode payload from cache", logger.FError(err))
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

func (c *cache) SaveSMSServices(ctx context.Context, services []sms.Service) error {
	log := c.container.GetLogger()
	log.Debug("will save sms services")
	var cacheResponse app.CacheResponse[[]sms.Service]
	cacheResponse.Result = services
	cacheResponse.TimeFetched = time.Now()
	encodedData, err := utils.EncodePayload(&cacheResponse)
	if err != nil {
		log.Debug("fail to encode payload", logger.FError(err))
		return err
	}
	return c.client.Set(ctx, smsServicesCacheKey, encodedData, 0).Err()
}

func (c *cache) GetSMSServices(ctx context.Context) (*app.CacheResponse[[]sms.Service], error) {
	log := c.container.GetLogger()
	log.Debug("will get sms countries")
	encodedText, err := c.client.Get(ctx, smsServicesCacheKey).Result()
	if err != nil {
		log.Debug("fail to get sms services from cache", logger.FError(err))
		return nil, err
	}
	var cacheResponse app.CacheResponse[[]sms.Service]
	if err := utils.DecodePayload(encodedText, &cacheResponse); err != nil {
		return nil, err
	}
	return &cacheResponse, nil
}

func (c *cache) SaveTelegramCallbackData(ctx context.Context, callbackData []app.TelegramCallbackData, userID, chatID, messageID int64) error {
	log := c.container.GetLogger()
	log.Debug("will save telegram callback data")
	var cacheResponse app.CacheResponse[[]string]
	key := keyForTelegramCallbackData(telegramCallbackDataCacheKey, userID, chatID, messageID)
	encodedData := make([]string, 0, len(callbackData))
	for _, item := range callbackData {
		encodedItem, err := utils.EncodeTelegramCallbackData(item)
		if err != nil {
			return err
		}
		encodedData = append(encodedData, *encodedItem)
	}
	cacheResponse.Result = encodedData
	finalEncoding, err := utils.EncodePayload(&cacheResponse)
	if err != nil {
		log.Debug("fail to encode payload", logger.FError(err))
		return err
	}
	return c.client.Set(ctx, key, finalEncoding, 4*time.Hour).Err()
}

func (c *cache) GetTelegramCallbackData(ctx context.Context, userID, chatID, messageID int64) ([]app.TelegramCallbackData, error) {
	log := c.container.GetLogger()
	log.Debug("will get telegram callback data")
	key := keyForTelegramCallbackData(telegramCallbackDataCacheKey, userID, chatID, messageID)
	encodedText, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var cacheResponse app.CacheResponse[[]string]
	if err := utils.DecodePayload(encodedText, &cacheResponse); err != nil {
		return nil, err
	}
	callbackData := make([]app.TelegramCallbackData, 0, len(cacheResponse.Result))
	for _, item := range cacheResponse.Result {
		decodedItem, err := utils.DecodeTelegramCallbackData(item)
		if err != nil {
			return nil, err
		}
		callbackData = append(callbackData, *decodedItem)
	}
	return callbackData, nil
}

func (c *cache) GetLastTelegramCallbackDataOperation(ctx context.Context, userID, chatID, messageID int64) (string, error) {
	log := c.container.GetLogger()
	log.Debug("will get telegram callback data")
	key := keyForTelegramCallbackData(telegramCallbackDataCacheKey, userID, chatID, messageID)
	operation, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return operation, nil
}

func (c *cache) SetLastTelegramCallbackDataOperation(ctx context.Context, operation string, userID, chatID, messageID int64) error {
	log := c.container.GetLogger()
	log.Debug("will get telegram callback data")
	key := keyForTelegramCallbackData(telegramCallbackDataCacheKey, userID, chatID, messageID)
	return c.client.Set(ctx, key, operation, 4*time.Hour).Err()
}

func keyForTelegramCallbackData(key string, userID, chatID, messageID int64) string {
	return fmt.Sprintf("%s/%d_%d_%d", key, userID, chatID, messageID)
}
