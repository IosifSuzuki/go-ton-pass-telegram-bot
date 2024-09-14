package localizer

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type Localizer interface {
	LocalizedString(key string) string
}

type localizer struct {
	localizer *i18n.Localizer
}

func NewLocalizer(bundle *i18n.Bundle, lang string) Localizer {
	return &localizer{
		localizer: i18n.NewLocalizer(bundle, lang),
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
