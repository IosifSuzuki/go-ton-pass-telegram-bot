package config

import (
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/utils"
	"net"
	"os"
	"strconv"
	"strings"
)

type Config interface {
	Address() string
	TelegramBotToken() string
	CryptoBotToken() string
	SMSKey() string
	Redis() Redis
	DB() DB
	AvailablePreferredCurrencies() []app.Currency
	AvailablePayCurrencies() []app.Currency
	CurrencyByAbbr(abbr string) *app.Currency
	AvailableLanguages() []app.Language
	AllLanguages() []app.Language
	LanguageByCode(code string) *app.Language
	LanguageByName(name string) *app.Language
}

type Redis struct {
	Host     string
	Port     string
	Password string
	DataBase int
}

type DB struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	Mode     string
}

func (r *Redis) Address() string {
	return net.JoinHostPort(r.Host, r.Port)
}

type config struct {
	serverAddr            string
	serverPort            string
	telegramBotToken      string
	cryptoBotToken        string
	smsServiceToken       string
	allLanguages          []app.Language
	localizedLanguageTags []string
	allCurrencies         []app.Currency
	redis                 Redis
	db                    DB
}

func (c *config) Address() string {
	return net.JoinHostPort(c.serverAddr, c.serverPort)
}

func (c *config) TelegramBotToken() string {
	return c.telegramBotToken
}

func (c *config) CryptoBotToken() string {
	return c.cryptoBotToken
}

func (c *config) SMSKey() string {
	return c.smsServiceToken
}

func (c *config) AvailablePreferredCurrencies() []app.Currency {
	allPreferredCurrencyABBRs := []string{
		"USD",
		"EUR",
		"UAH",
	}
	return utils.Filter(c.allCurrencies, func(currency app.Currency) bool {
		return utils.Contains(allPreferredCurrencyABBRs, func(abbr string) bool {
			return strings.EqualFold(abbr, currency.ABBR)
		})
	})
}

func (c *config) AvailablePayCurrencies() []app.Currency {
	allPayCurrenciesABBRs := []string{
		"USDT",
		"ETH",
		"BTC",
		"TON",
	}
	return utils.Filter(c.allCurrencies, func(currency app.Currency) bool {
		return utils.Contains(allPayCurrenciesABBRs, func(abbr string) bool {
			return strings.EqualFold(abbr, currency.ABBR)
		})
	})
}

func (c *config) CurrencyByAbbr(abbr string) *app.Currency {
	for _, currency := range c.allCurrencies {
		if currency.ABBR == abbr {
			return &currency
		}
	}
	return nil
}

func (c *config) LanguageByCode(code string) *app.Language {
	for _, language := range c.AvailableLanguages() {
		if language.Code == code {
			return &language
		}
	}
	return nil
}

func (c *config) LanguageByName(name string) *app.Language {
	for _, language := range c.allLanguages {
		if language.Name == name {
			return &language
		}
	}
	return nil
}

func (c *config) AllLanguages() []app.Language {
	return c.allLanguages
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

func (c *config) Redis() Redis {
	return c.redis
}

func (c *config) DB() DB {
	return c.db
}

func ParseConfig() (Config, error) {
	config := config{
		serverAddr:       os.Getenv("SERVER_HOST"),
		serverPort:       os.Getenv("SERVER_PORT"),
		telegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		cryptoBotToken:   os.Getenv("CRYPTO_BOT_TOKEN"),
		smsServiceToken:  os.Getenv("SMS_SERVICE_API_KEY"),
	}
	allLanguages, err := fetchAllLanguages()
	if err != nil {
		return nil, err
	}
	localizedLanguageTags, err := fetchLocalizedLanguageTags()
	if err != nil {
		return nil, err
	}
	allCurrencies, err := fetchAllCurrencies()
	if err != nil {
		return nil, err
	}

	config.allLanguages = allLanguages
	config.localizedLanguageTags = localizedLanguageTags
	config.allCurrencies = allCurrencies
	config.redis = ParseRedisConfig()
	config.db = ParseDBConfig()

	return &config, nil
}

func ParseRedisConfig() Redis {
	redis := Redis{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
	}
	dataBase, _ := strconv.Atoi(os.Getenv("REDIS_DATABASE"))
	redis.DataBase = dataBase
	return redis
}

func ParseDBConfig() DB {
	return DB{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Name:     os.Getenv("POSTGRES_DB"),
		Mode:     os.Getenv("POSTGRES_MODE"),
	}
}

func fetchAllLanguages() ([]app.Language, error) {
	var allLanguages = make([]app.Language, 0)
	if err := utils.MarshalFromFile("/jsons/languages.json", &allLanguages); err != nil {
		return nil, err
	}
	return allLanguages, nil
}

func fetchAllCurrencies() ([]app.Currency, error) {
	var allCurrencies = make([]app.Currency, 0)
	if err := utils.MarshalFromFile("/jsons/currencies.json", &allCurrencies); err != nil {
		return nil, err
	}
	return allCurrencies, nil
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
