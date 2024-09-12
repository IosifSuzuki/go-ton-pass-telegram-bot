package config

import (
	"net"
	"os"
)

type Config struct {
	ServerAddr       string
	ServerPort       string
	TelegramBotToken string
	SMSServiceToken  string
}

func (c Config) Address() string {
	return net.JoinHostPort(c.ServerAddr, c.ServerPort)
}

func ParseConfig() Config {
	config := Config{}
	config.ServerAddr = os.Getenv("SERVER_HOST")
	config.ServerPort = os.Getenv("SERVER_PORT")
	config.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	config.SMSServiceToken = os.Getenv("SMS_SERVICE_TOKEN")
	return config
}
