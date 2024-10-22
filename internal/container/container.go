package container

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go-ton-pass-telegram-bot/internal/config"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/localizer"
	"go-ton-pass-telegram-bot/pkg/logger"
)

type Container interface {
	GetLogger() logger.Logger
	GetConfig() config.Config
	GetLocalizer(langTag string) localizer.Localizer
	GetPreferredServiceCodesOrder() []string
	GetPreferredCountryCodesOrder() []string
	GetFlagEmoji(name string) *string
	GetExtraService(serviceCode string) *app.ExtraService
	PreloadData() error
}

type container struct {
	config                     config.Config
	bundle                     *i18n.Bundle
	logger                     logger.Logger
	preferredServiceCodesOrder []string
	preferredCountryCodesOrder []string
	emojiFlag                  map[string]string
	extraServices              map[string]app.ExtraService
}

func NewContainer(logger logger.Logger, config config.Config, bundle *i18n.Bundle) Container {
	return &container{
		config:                     config,
		bundle:                     bundle,
		logger:                     logger,
		preferredServiceCodesOrder: make([]string, 0),
		preferredCountryCodesOrder: make([]string, 0),
		emojiFlag:                  make(map[string]string),
		extraServices:              make(map[string]app.ExtraService),
	}
}

func (c *container) PreloadData() error {
	if err := utils.UnmarshalFromFile("/jsons/preferred_services_order.json", &c.preferredServiceCodesOrder); err != nil {
		return err
	}
	if err := utils.UnmarshalFromFile("/jsons/preferred_countries_order.json", &c.preferredCountryCodesOrder); err != nil {
		return err
	}
	var flagEmoji = make([]app.FlagEmoji, 0)
	if err := utils.UnmarshalFromFile("/jsons/country_flag_emoji.json", &flagEmoji); err != nil {
		return err
	}
	for _, value := range flagEmoji {
		c.emojiFlag[value.Name] = value.Flag
	}
	var extraServices = make([]app.ExtraService, 0)
	if err := utils.UnmarshalFromFile("/jsons/extra_info_services.json", &extraServices); err != nil {
		return err
	}
	for _, value := range extraServices {
		c.extraServices[value.Code] = value
	}
	return nil
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

func (c *container) GetPreferredServiceCodesOrder() []string {
	return c.preferredServiceCodesOrder
}

func (c *container) GetPreferredCountryCodesOrder() []string {
	return c.preferredCountryCodesOrder
}

func (c *container) GetFlagEmoji(name string) *string {
	flag, ok := c.emojiFlag[name]
	if ok {
		return &flag
	}
	return nil
}

func (c *container) GetExtraService(serviceCode string) *app.ExtraService {
	extraService, ok := c.extraServices[serviceCode]
	if ok {
		return &extraService
	}
	return nil
}
