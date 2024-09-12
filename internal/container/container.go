package container

import (
	"go-ton-pass-telegram-bot/internal/config"
)

type Container interface {
	GetTelegramBotToken() string
	GetServerAddress() string
}

type container struct {
	config config.Config
}

func NewContainer(config config.Config) Container {
	return &container{
		config: config,
	}
}

func (c *container) GetTelegramBotToken() string {
	return c.config.TelegramBotToken
}

func (c *container) GetServerAddress() string {
	return c.config.Address()
}
