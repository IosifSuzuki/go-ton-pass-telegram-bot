package service

import (
	"encoding/json"
	"fmt"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/crypto/bot"
	"go-ton-pass-telegram-bot/pkg/logger"
	"net/http"
	"net/url"
)

const (
	baseCryptoBotURL = "https://testnet-pay.crypt.bot/api"
)

type CryptoPayBot interface {
	CreateInvoice(currency string, amount float64, payloadData string) (*bot.Invoice, error)
	RemoveInvoice(invoiceID int64) error
	FetchExchangeRate() ([]bot.ExchangeRate, error)
}

type cryptoPayBot struct {
	container container.Container
}

func NewCryptoPayBot(container container.Container) CryptoPayBot {
	return &cryptoPayBot{
		container: container,
	}
}

func (c *cryptoPayBot) CreateInvoice(currency string, amount float64, payload string) (*bot.Invoice, error) {
	log := c.container.GetLogger()
	amountText := fmt.Sprintf("%.2f", amount)
	queryParams := url.Values{}
	queryParams.Set("currency_type", "crypto")
	queryParams.Set("asset", currency)
	queryParams.Set("amount", amountText)
	queryParams.Set("payload", payload)
	req, err := c.prepareRequest(app.CreateInvoiceCryptoBotMethod, queryParams)
	if err != nil {
		log.Error("fail prepare a request", logger.FError(err))
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("fail to create a http client", logger.FError(err))
		return nil, err
	}
	defer resp.Body.Close()
	var result bot.Result[bot.Invoice]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Debug("fail to decode", logger.FError(err))
		return nil, err
	}
	return &result.Result, nil
}

func (c *cryptoPayBot) RemoveInvoice(invoiceID int64) error {
	log := c.container.GetLogger()
	queryParams := url.Values{}
	queryParams.Set("invoice_id", fmt.Sprintf("%d", invoiceID))
	req, err := c.prepareRequest(app.DeleteInvoiceCryptoBotMethod, queryParams)
	if err != nil {
		log.Error("fail to prepare request", logger.FError(err))
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("fail to create a http client", logger.FError(err))
		return err
	}
	defer resp.Body.Close()
	var result bot.Result[bool]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Error("fail to parse response", logger.FError(err))
		return err
	}
	if !result.Result {
		log.Debug("fail to delete invoice")
		return app.DeleteInvoiceError
	}
	return nil
}

func (c *cryptoPayBot) FetchExchangeRate() ([]bot.ExchangeRate, error) {
	log := c.container.GetLogger()
	req, err := c.prepareRequest(app.ExchangeRateCryptoBotMethod, url.Values{})
	if err != nil {
		log.Error("fail prepare a request", logger.FError(err))
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("fail to create a http client", logger.FError(err))
		return nil, err
	}
	defer resp.Body.Close()
	var result bot.Result[[]bot.ExchangeRate]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Debug("fail to decode", logger.FError(err))
		return nil, err
	}
	return result.Result, nil
}

func (c *cryptoPayBot) prepareRequest(method app.CryptoBotMethod, queryParams url.Values) (*http.Request, error) {
	log := c.container.GetLogger()
	token := c.container.GetConfig().CryptoBotToken()
	urlPath, err := url.Parse(baseCryptoBotURL)
	if err != nil {
		log.Error("fail to parse url", logger.FError(err))
		return nil, err
	}
	urlPath = urlPath.JoinPath(string(method))
	urlPath.RawQuery = queryParams.Encode()
	req, err := http.NewRequest("GET", urlPath.String(), nil)
	if err != nil {
		log.Error("fail to create a request", logger.FError(err))
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Crypto-Pay-API-Token", token)
	log.Debug("prepare request", logger.F("url", req.URL))
	return req, nil
}
