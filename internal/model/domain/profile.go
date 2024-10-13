package domain

import "time"

type Profile struct {
	ID                int64
	TelegramID        int64
	TelegramChatID    int64
	Username          *string
	PreferredCurrency *string
	PreferredLanguage *string
	Balance           float64
	UpdatedAt         *time.Time
	CreatedAt         *time.Time
}
