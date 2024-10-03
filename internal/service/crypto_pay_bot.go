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
	CreateInvoice(currency string, amount float64) (*bot.Invoice, error)
}

type cryptoPayBot struct {
	container container.Container
}

func NewCryptoPayBot(container container.Container) CryptoPayBot {
	return &cryptoPayBot{
		container: container,
	}
}

func (c *cryptoPayBot) CreateInvoice(currency string, amount float64) (*bot.Invoice, error) {
	log := c.container.GetLogger()
	amountText := fmt.Sprintf("%.2f", amount)
	queryParams := url.Values{}
	queryParams.Set("currency_type", "crypto")
	queryParams.Set("asset", currency)
	queryParams.Set("amount", amountText)
	req, err := c.prepareRequest(app.CreateInvoiceCryptoBotMethod, queryParams)
	if err != nil {
		log.Debug("fail prepare a request", logger.FError(err))
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Debug("fail to create a http client", logger.FError(err))
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
