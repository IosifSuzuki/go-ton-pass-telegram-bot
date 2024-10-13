package repository

import (
	"context"
	"database/sql"
	"go-ton-pass-telegram-bot/internal/model/domain"
	"time"
)

type SMSHistoryRepository interface {
	Create(ctx context.Context, smsHistory *domain.SMSHistory) (*int64, error)
	GetByActivationID(ctx context.Context, activationID int64) (*domain.SMSHistory, error)
	ReceiveSMSCode(ctx context.Context, smsHistory *domain.SMSHistory) error
}

type smsHistoryRepository struct {
	conn *sql.DB
}

func NewSMSHistoryRepository(conn *sql.DB) SMSHistoryRepository {
	return &smsHistoryRepository{
		conn: conn,
	}
}

func (s *smsHistoryRepository) Create(ctx context.Context, smsHistory *domain.SMSHistory) (*int64, error) {
	query := "INSERT INTO sms_history (profile_id, activation_id, service_code, phone_number, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id;"
	var id int64
	err := s.conn.QueryRowContext(
		ctx,
		query,
		smsHistory.ProfileID,
		smsHistory.ActivationID,
		smsHistory.ServiceCode,
		smsHistory.PhoneNumber,
		time.Now(),
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &id, err
}

func (s *smsHistoryRepository) ReceiveSMSCode(ctx context.Context, smsHistory *domain.SMSHistory) error {
	query := "UPDATE sms_history SET sms_text = $1, sms_code = $2, received_at = $3, updated_at = $4 WHERE profile_id = $5"
	_, err := s.conn.ExecContext(ctx, query, smsHistory.SMSText, smsHistory.SMSCode, smsHistory.ReceivedAt, time.Now(), smsHistory.ProfileID)
	return err
}

func (s *smsHistoryRepository) GetByActivationID(ctx context.Context, activationID int64) (*domain.SMSHistory, error) {
	query := "SELECT id, profile_id, service_code, phone_number, sms_text, sms_code, received_at, " +
		"created_at, updated_at, deleted_at FROM sms_history WHERE activation_id = $1"
	row := s.conn.QueryRowContext(ctx, query, activationID)
	smsHistory := domain.SMSHistory{
		ActivationID: activationID,
	}
	var smsText sql.NullString
	var smsCode sql.NullString
	var receivedAt sql.NullTime
	var createdAt sql.NullTime
	var updatedAt sql.NullTime
	var deletedAt sql.NullTime

	err := row.Scan(
		&smsHistory.ID,
		&smsHistory.ProfileID,
		&smsHistory.ServiceCode,
		&smsHistory.PhoneNumber,
		&smsText,
		&smsCode,
		&receivedAt,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)
	if smsText.Valid {
		smsHistory.SMSText = &smsText.String
	}
	if smsCode.Valid {
		smsHistory.SMSCode = &smsCode.String
	}
	if receivedAt.Valid {
		smsHistory.ReceivedAt = &receivedAt.Time
	}
	if createdAt.Valid {
		smsHistory.CreatedAt = &createdAt.Time
	}
	if updatedAt.Valid {
		smsHistory.UpdatedAt = &updatedAt.Time
	}
	if deletedAt.Valid {
		smsHistory.DeletedAt = &deletedAt.Time
	}
	return &smsHistory, err
}
