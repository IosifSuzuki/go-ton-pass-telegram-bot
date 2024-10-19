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
	return e.filterAndConvert(networkExchangeRates), nil
}

func (e *exchangeRate) ConvertFromUSD(amount float64, targetCurrencyCode string) (*float64, error) {
	exchangeRate, err := e.findExchangeRateByCode(targetCurrencyCode)
	if err != nil {
		return nil, err
	}
	convertedAmount := amount / exchangeRate.Rate
	return &convertedAmount, nil
}

func (e *exchangeRate) ConvertFromRUB(amount float64, targetCurrencyCode string) (*float64, error) {
	rubExchangeRate, err := e.findExchangeRateByCode("RUB")
	if err != nil {
		return nil, err
	}
	usdValue := amount * rubExchangeRate.Rate
	return e.ConvertFromUSD(usdValue, targetCurrencyCode)
}

func (e *exchangeRate) PriceWithFee(amount float64) float64 {
	return 1.1 * amount
}

func (e *exchangeRate) ConvertToUSD(amount float64, sourceCurrencyCode string) (*float64, error) {
	exchangeRate, err := e.findExchangeRateByCode(sourceCurrencyCode)
	if err != nil {
		return nil, err
	}
	convertedAmount := exchangeRate.Rate * amount
	return &convertedAmount, nil
}

func (e *exchangeRate) findExchangeRateByCode(currencyCode string) (*app.ExchangeRate, error) {
	log := e.container.GetLogger()
	ctx := context.Background()
	exchangeRates, err := e.GetExchangeRate(ctx)
	if err != nil {
		log.Error("fail to get exchange rate", logger.FError(err))
		return nil, err
	}
	var targetExchangeRate *app.ExchangeRate
	for _, exchangeRate := range exchangeRates {
		if strings.EqualFold(exchangeRate.Currency, currencyCode) {
			var exchangeRate = exchangeRate
			log.Debug("found currency", logger.F("targetCurrencyCode", currencyCode))
			targetExchangeRate = &exchangeRate
		}
	}
	if strings.EqualFold("USD", currencyCode) {
		targetExchangeRate = &app.ExchangeRate{
			Currency: currencyCode,
			Rate:     1,
		}
	}
	if targetExchangeRate == nil {
		return nil, app.NilError
	}
	log.Debug("found exchangeRate", logger.F("targetExchangeRate", targetExchangeRate))
	return targetExchangeRate, nil
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
	if timeFetched.Add(24*time.Hour).Compare(time.Now()) <= 0 {
		return nil
	}
	networkExchangeRates, err := e.cryptoBot.FetchExchangeRate()
	if err != nil {
		log.Debug("fail to fetch exchange rate", logger.FError(err))
		return err
	}
	exchangeRates := e.filterAndConvert(networkExchangeRates)
	return e.cache.SaveExchangeRate(ctx, exchangeRates)
}

func (e *exchangeRate) filterAndConvert(networkExchangeRates []bot.ExchangeRate) []app.ExchangeRate {
	filteredNetworkExchangeRates := utils.Filter(networkExchangeRates, func(exchangeRate bot.ExchangeRate) bool {
		if !exchangeRate.IsValid {
			return false
		}
		return strings.EqualFold(exchangeRate.Target, "USD")
	})
	exchangeRates := make([]app.ExchangeRate, 0, len(filteredNetworkExchangeRates))
	for _, networkExchangeRate := range filteredNetworkExchangeRates {
		rate, err := strconv.ParseFloat(networkExchangeRate.Rate, 64)
		if err != nil {
			continue
		}
		exchangeRates = append(exchangeRates, app.ExchangeRate{
			Currency: networkExchangeRate.Source,
			Rate:     rate,
		})
	}
	return exchangeRates
}
