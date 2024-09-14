package config

import (
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/utils"
	"net"
	"os"
	"strings"
)

type Config interface {
	Address() string
	TelegramBotToken() string
	AvailableCurrencies() []app.Currency
	AvailableLanguages() []app.Language
}

type config struct {
	serverAddr            string
	serverPort            string
	telegramBotToken      string
	smsServiceToken       string
	allLanguages          []app.Language
	localizedLanguageTags []string
	availableCurrencies   []app.Currency
}

func (c *config) Address() string {
	return net.JoinHostPort(c.serverAddr, c.serverPort)
}

func (c *config) TelegramBotToken() string {
	return c.telegramBotToken
}

func (c *config) AvailableCurrencies() []app.Currency {
	return c.availableCurrencies
}

func (c *config) AvailableLanguages() []app.Language {
	return utils.Filter(c.allLanguages, func(language app.Language) bool {
		for _, localizedLanguageTag := range c.localizedLanguageTags {
			if language.Code == localizedLanguageTag {
				return true
			}
		}
		return false
	})
}

func ParseConfig() (Config, error) {
	config := config{
		serverAddr:       os.Getenv("SERVER_HOST"),
		serverPort:       os.Getenv("SERVER_PORT"),
		telegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		smsServiceToken:  os.Getenv("SMS_SERVICE_TOKEN"),
	}
	allLanguages, err := fetchAllLanguages()
	if err != nil {
		return nil, err
	}
	localizedLanguageTags, err := fetchLocalizedLanguageTags()
	if err != nil {
		return nil, err
	}
	availableCurrencies, err := fetchAvailableCurrencies()
	if err != nil {
		return nil, err
	}

	config.allLanguages = allLanguages
	config.localizedLanguageTags = localizedLanguageTags
	config.availableCurrencies = availableCurrencies

	return &config, nil
}

func fetchAllLanguages() ([]app.Language, error) {
	var allLanguages = make([]app.Language, 0)
	if err := utils.MarshalFromFile("/jsons/languages.json", allLanguages); err != nil {
		return nil, err
	}
	return allLanguages, nil
}

func fetchAvailableCurrencies() ([]app.Currency, error) {
	var availableCurrencies = make([]app.Currency, 0)
	if err := utils.MarshalFromFile("/jsons/currencies.json", availableCurrencies); err != nil {
		return nil, err
	}
	return availableCurrencies, nil
}

func fetchLocalizedLanguageTags() ([]string, error) {
	localizedFilePaths, err := utils.FilePaths("/locales")
	if err != nil {
		return nil, err
	}
	var localizedLanguageTags = make([]string, 0, len(localizedFilePaths))
	for _, localizedFilePath := range localizedFilePaths {
		languageTag := strings.TrimSuffix(localizedFilePath, ".json")
		languageTag = strings.TrimPrefix(languageTag, "/locales/")
		localizedLanguageTags = append(localizedLanguageTags, languageTag)
	}
	return localizedLanguageTags, nil
}
