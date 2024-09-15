package domain

import "time"

type Profile struct {
	ID                int64
	TelegramID        int64
	Username          *string
	PreferredCurrency *string
	PreferredLanguage *string
	Balance           int64
	UpdatedAt         *time.Time
	CreatedAt         *time.Time
}
