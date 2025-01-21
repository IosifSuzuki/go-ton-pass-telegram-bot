package localizer

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type Localizer interface {
	LocalizedString(key string) string
	LocalizedStringWithTemplateData(key string, templateData map[string]any) string
	GetISOLang() string
}

type localizer struct {
	localizer *i18n.Localizer
	lang      string
}

func NewLocalizer(bundle *i18n.Bundle, lang string) Localizer {
	return &localizer{
		localizer: i18n.NewLocalizer(bundle, lang),
		lang:      lang,
	}
}

func (l *localizer) LocalizedString(key string) string {
	localizedString, err := l.localizer.LocalizeMessage(&i18n.Message{
		ID: key,
	})
	if err != nil {
		// fallback with localized key
		return key
	}
	return localizedString
}

func (l *localizer) LocalizedStringWithTemplateData(key string, templateData map[string]any) string {
	localizedString := l.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: templateData,
	})
	return localizedString
}

func (l *localizer) GetISOLang() string {
	return l.lang
}
