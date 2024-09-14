package container

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go-ton-pass-telegram-bot/internal/config"
	"go-ton-pass-telegram-bot/pkg/localizer"
	"go-ton-pass-telegram-bot/pkg/logger"
)

type Container interface {
	GetLogger() logger.Logger
	GetConfig() config.Config
	GetLocalizer(langTag string) localizer.Localizer
}

type container struct {
	config config.Config
	bundle *i18n.Bundle
	logger logger.Logger
}

func NewContainer(logger logger.Logger, config config.Config, bundle *i18n.Bundle) Container {
	return &container{
		config: config,
		bundle: bundle,
		logger: logger,
	}
}

func (c *container) GetLocalizer(langTag string) localizer.Localizer {
	return localizer.NewLocalizer(c.bundle, langTag)
}

func (c *container) GetLogger() logger.Logger {
	return c.logger
}

func (c *container) GetConfig() config.Config {
	return c.config
}
