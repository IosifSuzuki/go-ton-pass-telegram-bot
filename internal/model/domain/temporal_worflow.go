package domain

import "time"

type TemporalWorkflow struct {
	ID            int64
	SMSHistoryID  int64
	TemporalID    string
	TemporalRunID string
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
	DeletedAt     *time.Time
}
