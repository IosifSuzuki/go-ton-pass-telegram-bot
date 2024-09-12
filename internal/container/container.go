package container

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go-ton-pass-telegram-bot/internal/config"
	"go-ton-pass-telegram-bot/pkg/logger"
)

type Container interface {
	GetTelegramBotToken() string
	GetServerAddress() string
	LocalizedString(key string) string
	SetLocalizedLanguage(langTag string)
	GetLogger() logger.Logger
}

type container struct {
	config    config.Config
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
	logger    logger.Logger
}

func NewContainer(logger logger.Logger, config config.Config, bundle *i18n.Bundle) Container {
	return &container{
		config:    config,
		bundle:    bundle,
		localizer: i18n.NewLocalizer(bundle),
		logger:    logger,
	}
}

func (c *container) GetTelegramBotToken() string {
	return c.config.TelegramBotToken
}

func (c *container) GetServerAddress() string {
	return c.config.Address()
}

func (c *container) LocalizedString(key string) string {
	localizedString, err := c.localizer.LocalizeMessage(&i18n.Message{
		ID: key,
	})
	if err != nil {
		c.logger.Error("fail to extract localizedString", logger.FError(err))
		// fallback with localized key
		return key
	}
	return localizedString
}

func (c *container) SetLocalizedLanguage(langTag string) {
	c.localizer = i18n.NewLocalizer(c.bundle, langTag)
}

func (c *container) GetLogger() logger.Logger {
	return c.logger
}
