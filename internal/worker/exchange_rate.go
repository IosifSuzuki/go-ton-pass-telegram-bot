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
	UpToDateExchangeRate(ctx context.Context) error
	Convert(amount float64, sourceCurrencyCode, targetCurrencyCode string) (*float64, error)
	ConvertFromUSD(amount float64, targetCurrencyCode string) (*float64, error)
	ConvertToUSD(amount float64, sourceCurrencyCode string) (*float64, error)
	ConvertFromRUB(amount float64, targetCurrencyCode string) (*float64, error)
	PriceWithFee(amount float64) float64
}

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
	go func() {
		ctx := context.Background()
		_ = e.UpToDateExchangeRate(ctx)
	}()

	log := e.container.GetLogger()
	response, err := e.cache.GetExchangeRate(ctx)
	if err == nil {
		return response.Result, nil
	}
	networkExchangeRates, err := e.cryptoBot.FetchExchangeRate()
	if err != nil {
		log.Error("fail to fetch exchange rate", logger.FError(err))
		return nil, err
	}
	return e.mapNetworkResponse(networkExchangeRates), nil
}

func (e *exchangeRate) ConvertFromUSD(amount float64, targetCurrencyCode string) (*float64, error) {
	return e.Convert(amount, "USD", targetCurrencyCode)
}

func (e *exchangeRate) ConvertFromRUB(amount float64, targetCurrencyCode string) (*float64, error) {
	return e.Convert(amount, "RUB", targetCurrencyCode)
}

func (e *exchangeRate) PriceWithFee(amount float64) float64 {
	return 1.15 * amount
}

func (e *exchangeRate) ConvertToUSD(amount float64, sourceCurrencyCode string) (*float64, error) {
	return e.Convert(amount, sourceCurrencyCode, "USD")
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
	if err != nil {
		log.Debug("fail to find exchange rate", logger.FError(err))
		return nil, err
	}
	var rate float64
	if strings.EqualFold(exchangeRate.SourceCurrency, sourceCurrencyCode) {
		rate = exchangeRate.Rate
	} else {
		rate = 1 / exchangeRate.Rate
	}
	return &rate, nil
}

func (e *exchangeRate) findExchangeRate(currencyCode1, currencyCode2 string) (*app.ExchangeRate, error) {
	log := e.container.GetLogger()
	ctx := context.Background()
	exchangeRates, err := e.GetExchangeRate(ctx)
	if err != nil {
		log.Error("fail to get exchange rate", logger.FError(err))
		return nil, err
	}
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
		log.Debug("fail to find exchange rate")
		return nil, app.NilError
	}
	return foundExchangeRate, nil
}

func (e *exchangeRate) UpToDateExchangeRate(ctx context.Context) error {
	log := e.container.GetLogger()
	response, err := e.cache.GetExchangeRate(ctx)
	var timeFetched time.Time
	if response != nil {
		timeFetched = response.TimeFetched
	} else {
		timeFetched = time.Now()
	}
	if timeFetched.Add(24*time.Hour).Compare(time.Now()) >= 0 {
		return nil
	}
	networkExchangeRates, err := e.cryptoBot.FetchExchangeRate()
	if err != nil {
		log.Debug("fail to fetch exchange rate", logger.FError(err))
		return err
	}
	exchangeRates := e.mapNetworkResponse(networkExchangeRates)
	return e.cache.SaveExchangeRate(ctx, exchangeRates)
}

func (e *exchangeRate) mapNetworkResponse(networkExchangeRates []bot.ExchangeRate) []app.ExchangeRate {
	filteredNetworkExchangeRates := utils.Filter(networkExchangeRates, func(exchangeRate bot.ExchangeRate) bool {
		return exchangeRate.IsValid
	})
	exchangeRates := make([]app.ExchangeRate, 0, len(filteredNetworkExchangeRates))
	for _, networkExchangeRate := range filteredNetworkExchangeRates {
		rate, err := strconv.ParseFloat(networkExchangeRate.Rate, 64)
		if err != nil {
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
