package domain

import "time"

type Invoice struct {
	ID        int64
	ProfileID int64
	InvoiceID int64
	ChatID    int64
	Status    string
	UpdatedAt *time.Time
	CreatedAt *time.Time
}
