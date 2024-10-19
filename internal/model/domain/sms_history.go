package domain

import "time"

type SMSHistory struct {
	ID           int64
	ProfileID    int64
	ActivationID int64
	Status       string
	ServiceCode  string
	PhoneNumber  string
	SMSText      *string
	SMSCode      *string
	ReceivedAt   *time.Time
	UpdatedAt    *time.Time
	CreatedAt    *time.Time
	DeletedAt    *time.Time
}
