package worker

import (
	"fmt"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/sms"
)

type FormatterType uint

const (
	DefaultFormatterType FormatterType = iota
)

type Formatter interface {
	Country(country *sms.Country, formatterType FormatterType) string
	Service(service *sms.Service, formatterType FormatterType) string
}

type formatter struct {
	container container.Container
}

func NewFormatter(container container.Container) Formatter {
	f := formatter{
		container: container,
	}
	return &f
}

func (f *formatter) Country(country *sms.Country, _ FormatterType) string {
	var title string
	flag := f.container.GetFlagEmoji(country.Title)
	if flag != nil {
		title = fmt.Sprintf("%s %s", *flag, country.Title)
	} else {
		title = country.Title
	}
	return title
}

func (f *formatter) Service(service *sms.Service, _ FormatterType) string {
	var (
		name  string
		emoji string
	)
	extraService := f.container.GetExtraService(service.Code)
	if extraService != nil {
		name = extraService.Name
		emoji = extraService.Emoji
	} else {
		name = service.Name
		emoji = "🌐"
	}
	return fmt.Sprintf("%s %s", emoji, name)
}
