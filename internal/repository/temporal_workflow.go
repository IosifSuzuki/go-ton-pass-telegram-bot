package repository

import (
	"context"
	"database/sql"
	"go-ton-pass-telegram-bot/internal/model/domain"
	"time"
)

type TemporalWorkflowRepository interface {
	Create(ctx context.Context, temporalWorkflow *domain.TemporalWorkflow) (*int64, error)
	GetBySMSHistoryID(ctx context.Context, smsHistoryID int64) (*domain.TemporalWorkflow, error)
}

type temporalWorkflowRepository struct {
	conn *sql.DB
}

func NewTemporalWorkflowRepository(conn *sql.DB) TemporalWorkflowRepository {
	return &temporalWorkflowRepository{
		conn: conn,
	}
}

func (t *temporalWorkflowRepository) Create(ctx context.Context, temporalWorkflow *domain.TemporalWorkflow) (*int64, error) {
	query := "INSERT INTO temporal_workflow (sms_history_id, temporal_id, temporal_run_id, created_at) VALUES ($1, $2, $3, $4) " +
		"RETURNING id;"
	var id int64
	err := t.conn.QueryRowContext(
		ctx,
		query,
		temporalWorkflow.SMSHistoryID,
		temporalWorkflow.TemporalID,
		temporalWorkflow.TemporalRunID,
		time.Now(),
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (t *temporalWorkflowRepository) GetBySMSHistoryID(ctx context.Context, smsHistoryID int64) (*domain.TemporalWorkflow, error) {
	query := "SELECT id, temporal_id, temporal_run_id, created_at, updated_at, deleted_at FROM temporal_workflow WHERE sms_history_id = $1"
	row := t.conn.QueryRowContext(ctx, query, smsHistoryID)
	smsHistory := domain.TemporalWorkflow{
		SMSHistoryID: smsHistoryID,
	}
	var createdAt sql.NullTime
	var updatedAt sql.NullTime
	var deletedAt sql.NullTime
	err := row.Scan(
		&smsHistory.ID,
		&smsHistory.TemporalID,
		&smsHistory.TemporalRunID,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		return nil, err
	}
	if createdAt.Valid {
		smsHistory.CreatedAt = &createdAt.Time
	}
	if updatedAt.Valid {
		smsHistory.CreatedAt = &createdAt.Time
	}
	if deletedAt.Valid {
		smsHistory.CreatedAt = &createdAt.Time
	}
	return &smsHistory, nil
}
