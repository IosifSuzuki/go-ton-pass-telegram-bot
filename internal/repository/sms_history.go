package repository

import (
	"context"
	"database/sql"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/domain"
	"time"
)

type SMSHistoryRepository interface {
	Create(ctx context.Context, smsHistory *domain.SMSHistory) (*int64, error)
	GetByActivationID(ctx context.Context, activationID int64) (*domain.SMSHistory, error)
	ChangeActivationStatus(ctx context.Context, activationID int64, activationStatus string) error
	ReceiveSMSCode(ctx context.Context, smsHistory *domain.SMSHistory) error
	GetNumberOfRows(ctx context.Context, profileID int64) (*int64, error)
	FetchList(ctx context.Context, profileID int64, offset int, limit int) ([]domain.SMSHistory, error)
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
	query := "INSERT INTO sms_history (profile_id, activation_id, service_code, service_name, country_id, country_name, " +
		"phone_code_number, phone_short_number, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) " +
		"RETURNING id;"
	var id int64
	err := s.conn.QueryRowContext(
		ctx,
		query,
		smsHistory.ProfileID,
		smsHistory.ActivationID,
		smsHistory.ServiceCode,
		smsHistory.ServiceName,
		smsHistory.CountryID,
		smsHistory.CountryName,
		smsHistory.PhoneCodeNumber,
		smsHistory.PhoneShortNumber,
		smsHistory.Status,
		time.Now(),
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &id, err
}

func (s *smsHistoryRepository) ReceiveSMSCode(ctx context.Context, smsHistory *domain.SMSHistory) error {
	query := "UPDATE sms_history SET sms_text = $1, sms_code = $2, status = $3, received_at = $4, updated_at = $5 WHERE activation_id = $6"
	_, err := s.conn.ExecContext(ctx, query, smsHistory.SMSText, smsHistory.SMSCode, app.DoneSMSActivateState, smsHistory.ReceivedAt, time.Now(), smsHistory.ActivationID)
	return err
}

func (s *smsHistoryRepository) GetByActivationID(ctx context.Context, activationID int64) (*domain.SMSHistory, error) {
	query := "SELECT id, profile_id, service_code, service_name, country_id, country_name, phone_code_number, phone_short_number, status, sms_text, sms_code, received_at, " +
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
		&smsHistory.ServiceName,
		&smsHistory.CountryID,
		&smsHistory.CountryName,
		&smsHistory.PhoneCodeNumber,
		&smsHistory.PhoneShortNumber,
		&smsHistory.Status,
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

func (s *smsHistoryRepository) ChangeActivationStatus(ctx context.Context, activationID int64, activationStatus string) error {
	query := "UPDATE sms_history SET status = $1, updated_at = $2 WHERE activation_id = $3"
	_, err := s.conn.ExecContext(ctx, query, activationStatus, time.Now(), activationID)
	return err
}

func (s *smsHistoryRepository) GetNumberOfRows(ctx context.Context, profileID int64) (*int64, error) {
	var numberOfRows int64
	query := "SELECT COUNT(*) FROM sms_history WHERE profile_id = $1"
	err := s.conn.QueryRowContext(ctx, query, profileID).Scan(&numberOfRows)
	if err != nil {
		return nil, err
	}
	return &numberOfRows, nil
}

func (s *smsHistoryRepository) FetchList(ctx context.Context, profileID int64, offset int, limit int) ([]domain.SMSHistory, error) {
	query := "SELECT id, profile_id, activation_id, service_code, service_name, country_id, country_name, phone_code_number, phone_short_number, status, sms_text, sms_code, received_at, created_at, updated_at, deleted_at " +
		"FROM sms_history WHERE profile_id = $1  ORDER BY created_at DESC LIMIT $2 OFFSET $3"
	rows, err := s.conn.QueryContext(ctx, query, profileID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]domain.SMSHistory, 0, limit)
	for rows.Next() {
		smsHistory := domain.SMSHistory{}
		var smsText sql.NullString
		var smsCode sql.NullString
		var receivedAt sql.NullTime
		var createdAt sql.NullTime
		var updatedAt sql.NullTime
		var deletedAt sql.NullTime
		err := rows.Scan(
			&smsHistory.ID,
			&smsHistory.ProfileID,
			&smsHistory.ActivationID,
			&smsHistory.ServiceCode,
			&smsHistory.ServiceName,
			&smsHistory.CountryID,
			&smsHistory.CountryName,
			&smsHistory.PhoneCodeNumber,
			&smsHistory.PhoneShortNumber,
			&smsHistory.Status,
			&smsText,
			&smsCode,
			&receivedAt,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			continue
		}
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
		list = append(list, smsHistory)
	}
	return list, nil
}
