package worker

import (
	"context"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/crypto/bot"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
	"strconv"
	"strings"
	"time"
)

type ExchangeRate interface {
	GetExchangeRate(ctx context.Context) ([]app.ExchangeRate, error)
	Convert(amount float64, sourceCurrencyCode, targetCurrencyCode string) (*float64, error)
	ConvertFromUSD(amount float64, targetCurrencyCode string) (*float64, error)
	ConvertToUSD(amount float64, sourceCurrencyCode string) (*float64, error)
	ConvertFromRUB(amount float64, targetCurrencyCode string) (*float64, error)
	PriceWithFee(amount float64) float64
}

const feePriceRate = 1.15

type exchangeRate struct {
	container container.Container
	cryptoBot service.CryptoPayBot
	cache     service.Cache
}

func NewExchangeRate(container container.Container, cache service.Cache, cryptoBot service.CryptoPayBot) ExchangeRate {
	return &exchangeRate{
		container: container,
		cryptoBot: cryptoBot,
		cache:     cache,
	}
}

func (e *exchangeRate) GetExchangeRate(ctx context.Context) ([]app.ExchangeRate, error) {
	log := e.container.GetLogger()
	response, err := e.cache.GetExchangeRate(ctx)
	if err == nil && !e.shouldUpToDateExchangeRate(response) {
		log.Debug("get exchange rates from cache")
		return response.Result, nil
	}
	networkExchangeRates, err := e.cryptoBot.FetchExchangeRate()
	if err != nil {
		log.Debug("fail to fetch exchange rate", logger.FError(err))
		return nil, err
	}
	exchangeRates := e.mapNetworkResponse(networkExchangeRates)
	if exchangeRates == nil {
		log.Debug("exchange rates must contains not nil value")
		return nil, app.NilError
	}
	// skip error checking
	go func() {
		_ = e.cache.SaveExchangeRate(ctx, exchangeRates)
	}()
	return exchangeRates, nil
}

func (e *exchangeRate) ConvertFromUSD(amount float64, targetCurrencyCode string) (*float64, error) {
	return e.Convert(amount, "USD", targetCurrencyCode)
}

func (e *exchangeRate) ConvertFromRUB(amount float64, targetCurrencyCode string) (*float64, error) {
	return e.Convert(amount, "RUB", targetCurrencyCode)
}

func (e *exchangeRate) ConvertToUSD(amount float64, sourceCurrencyCode string) (*float64, error) {
	return e.Convert(amount, sourceCurrencyCode, "USD")
}

func (e *exchangeRate) PriceWithFee(amount float64) float64 {
	return feePriceRate * amount
}

func (e *exchangeRate) Convert(amount float64, sourceCurrencyCode, targetCurrencyCode string) (*float64, error) {
	log := e.container.GetLogger()
	rate, err := e.GetRate(sourceCurrencyCode, targetCurrencyCode)
	if err != nil {
		log.Debug("fail to get rate", logger.FError(err))
		return nil, err
	}
	targetAmount := amount * (*rate)
	log.Debug("get rate",
		logger.F("source_currency_code", sourceCurrencyCode),
		logger.F("target_currency_code", targetCurrencyCode),
		logger.F("rate", rate),
		logger.F("source_amount", amount),
		logger.F("target_amount", targetAmount),
	)
	return &targetAmount, nil
}

func (e *exchangeRate) GetRate(sourceCurrencyCode, targetCurrencyCode string) (*float64, error) {
	log := e.container.GetLogger()
	exchangeRate, err := e.findExchangeRate(sourceCurrencyCode, targetCurrencyCode)
	var rate float64
	if err == nil {
		if strings.EqualFold(exchangeRate.SourceCurrency, sourceCurrencyCode) {
			rate = exchangeRate.Rate
		} else {
			rate = 1 / exchangeRate.Rate
		}
	} else {
		sourceCurrencyRateInUSD, err := e.GetRateToUSD(sourceCurrencyCode)
		if err != nil {
			log.Debug("fail to get rate to usd", logger.F("currency_code", sourceCurrencyRateInUSD))
			return nil, err
		}
		targetCurrencyRateInUSD, err := e.GetRateToUSD(targetCurrencyCode)
		if err != nil {
			log.Debug("fail to get rate to usd", logger.F("currency_code", targetCurrencyRateInUSD))
			return nil, err
		}
		rate = (*sourceCurrencyRateInUSD) / (*targetCurrencyRateInUSD)
	}
	log.Debug(
		"exchange rate",
		logger.F("source_currency_code", sourceCurrencyCode),
		logger.F("target_currency_code", targetCurrencyCode),
		logger.F("rate", rate),
	)
	return &rate, nil
}

func (e *exchangeRate) GetRateToUSD(currencyCode string) (*float64, error) {
	rate, err := e.findExchangeRate("USD", currencyCode)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(rate.TargetCurrency, "USD") {
		return &rate.Rate, nil
	}
	return nil, app.NilError
}

func (e *exchangeRate) findExchangeRate(currencyCode1, currencyCode2 string) (*app.ExchangeRate, error) {
	log := e.container.GetLogger()
	ctx := context.Background()
	exchangeRates, err := e.GetExchangeRate(ctx)
	if err != nil {
		log.Debug("fail to get exchange rate", logger.FError(err))
		return nil, err
	}
	log.Debug(
		"parameters for find exchange rate",
		logger.F("currencyCode1", currencyCode1),
		logger.F("currencyCode2", currencyCode2),
	)
	var foundExchangeRate *app.ExchangeRate
	for _, exchangeRate := range exchangeRates {
		conditionVariant1 := strings.EqualFold(exchangeRate.SourceCurrency, currencyCode1) && strings.EqualFold(exchangeRate.TargetCurrency, currencyCode2)
		conditionVariant2 := strings.EqualFold(exchangeRate.TargetCurrency, currencyCode1) && strings.EqualFold(exchangeRate.SourceCurrency, currencyCode2)
		if conditionVariant1 || conditionVariant2 {
			var exchangeRate = exchangeRate
			log.Debug("found exchange rate", logger.F("exchange_rate", exchangeRate))
			foundExchangeRate = &exchangeRate
		}
	}
	if foundExchangeRate == nil {
		log.Debug("foundExchangeRate has nil value")
		return nil, app.NilError
	}
	return foundExchangeRate, nil
}

func (e *exchangeRate) mapNetworkResponse(networkExchangeRates []bot.ExchangeRate) []app.ExchangeRate {
	log := e.container.GetLogger()
	filteredNetworkExchangeRates := utils.Filter(networkExchangeRates, func(exchangeRate bot.ExchangeRate) bool {
		return exchangeRate.IsValid
	})
	log.Debug("valid exchange rates from response", logger.F("len", len(filteredNetworkExchangeRates)))
	exchangeRates := make([]app.ExchangeRate, 0, len(filteredNetworkExchangeRates))
	for _, networkExchangeRate := range filteredNetworkExchangeRates {
		rate, err := strconv.ParseFloat(networkExchangeRate.Rate, 64)
		if err != nil {
			log.Debug("fail to parse float from string", logger.F("rate_string", networkExchangeRate.Rate), logger.FError(err))
			continue
		}
		exchangeRates = append(exchangeRates, app.ExchangeRate{
			SourceCurrency: networkExchangeRate.Source,
			TargetCurrency: networkExchangeRate.Target,
			Rate:           rate,
		})
	}
	exchangeRates = append(exchangeRates, app.ExchangeRate{
		SourceCurrency: "USD",
		TargetCurrency: "USD",
		Rate:           1,
	})
	return exchangeRates
}

func (e *exchangeRate) shouldUpToDateExchangeRate(cacheResponse *app.CacheResponse[[]app.ExchangeRate]) bool {
	log := e.container.GetLogger()
	var timeFetched time.Time
	if cacheResponse != nil {
		log.Debug("response from cache hasn't nil value")
		timeFetched = cacheResponse.TimeFetched
	} else {
		log.Debug("response from cache has nil value")
		timeFetched = time.Now()
	}
	shouldUpToDateExchangeRate := timeFetched.Add(24*time.Hour).Compare(time.Now()) < 0
	if shouldUpToDateExchangeRate {
		log.Debug("should update exchange rate")
	}
	return shouldUpToDateExchangeRate
}
