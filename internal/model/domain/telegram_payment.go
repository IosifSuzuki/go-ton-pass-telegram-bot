package domain

import "time"

type TelegramPayment struct {
	ID                      int64
	ProfileID               int64
	TelegramPaymentChargeID string
	Currency                string
	Amount                  int64
	CreditAmount            float64
	IsRefunded              bool
	CreatedAt               *time.Time
	UpdatedAt               *time.Time
	DeletedAt               *time.Time
}
