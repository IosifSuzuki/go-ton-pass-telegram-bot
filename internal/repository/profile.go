package repository

import (
	"context"
	"database/sql"
	"go-ton-pass-telegram-bot/internal/model/domain"
	"time"
)

type ProfileRepository interface {
	Create(ctx context.Context, profile *domain.Profile) (*int64, error)
	ExistsWithTelegramID(ctx context.Context, telegramID int64) (bool, error)
	FetchByTelegramID(ctx context.Context, telegramID int64) (*domain.Profile, error)
	FetchByID(ctx context.Context, id int64) (*domain.Profile, error)
	SetPreferredCurrency(ctx context.Context, telegramID int64, preferredCurrency string) error
	SetPreferredLanguage(ctx context.Context, telegramID int64, preferredLanguage string) error
	TopUpBalance(ctx context.Context, telegramID int64, amount float64) error
}
type profileRepository struct {
	conn *sql.DB
}

func NewProfileRepository(conn *sql.DB) ProfileRepository {
	return &profileRepository{
		conn: conn,
	}
}

func (p *profileRepository) Create(ctx context.Context, profile *domain.Profile) (*int64, error) {
	query := "INSERT INTO profile (telegram_id, telegram_chat_id, username, balance, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id;"
	var id int64
	err := p.conn.QueryRowContext(
		ctx,
		query,
		profile.TelegramID,
		profile.TelegramChatID,
		profile.Username,
		profile.Balance,
		time.Now(),
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &id, err
}

func (p *profileRepository) ExistsWithTelegramID(ctx context.Context, telegramID int64) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM profile WHERE telegram_id=$1);"
	var exists bool
	err := p.conn.QueryRowContext(ctx, query, telegramID).Scan(&exists)
	return exists, err
}

func (p *profileRepository) FetchByTelegramID(ctx context.Context, telegramID int64) (*domain.Profile, error) {
	query := "SELECT id, telegram_chat_id, username, preferred_currency, preferred_language, balance, created_at, updated_at FROM profile WHERE telegram_id = $1"
	row := p.conn.QueryRowContext(ctx, query, telegramID)
	profile := domain.Profile{
		TelegramID: telegramID,
		CreatedAt:  new(time.Time),
	}
	var preferredCurrency sql.NullString
	var preferredLanguage sql.NullString
	var updatedAt sql.NullTime

	err := row.Scan(
		&profile.ID,
		&profile.TelegramChatID,
		&profile.Username,
		&preferredCurrency,
		&preferredLanguage,
		&profile.Balance,
		&profile.CreatedAt,
		&updatedAt,
	)
	if preferredCurrency.Valid {
		profile.PreferredCurrency = &preferredCurrency.String
	}
	if preferredLanguage.Valid {
		profile.PreferredLanguage = &preferredLanguage.String
	}
	if updatedAt.Valid {
		profile.UpdatedAt = &updatedAt.Time
	}
	return &profile, err
}

func (p *profileRepository) FetchByID(ctx context.Context, id int64) (*domain.Profile, error) {
	query := "SELECT telegram_id, telegram_chat_id, username, preferred_currency, preferred_language, balance, created_at, updated_at FROM profile WHERE id = $1"
	row := p.conn.QueryRowContext(ctx, query, id)
	profile := domain.Profile{
		ID:        id,
		CreatedAt: new(time.Time),
	}
	var preferredCurrency sql.NullString
	var preferredLanguage sql.NullString
	var updatedAt sql.NullTime

	err := row.Scan(
		&profile.TelegramID,
		&profile.TelegramChatID,
		&profile.Username,
		&preferredCurrency,
		&preferredLanguage,
		&profile.Balance,
		&profile.CreatedAt,
		&updatedAt,
	)
	if preferredCurrency.Valid {
		profile.PreferredCurrency = &preferredCurrency.String
	}
	if preferredLanguage.Valid {
		profile.PreferredLanguage = &preferredLanguage.String
	}
	if updatedAt.Valid {
		profile.UpdatedAt = &updatedAt.Time
	}
	return &profile, err
}

func (p *profileRepository) SetPreferredCurrency(ctx context.Context, telegramID int64, preferredCurrency string) error {
	query := "UPDATE profile SET preferred_currency = $1, updated_at = $2 WHERE telegram_id = $3"
	_, err := p.conn.ExecContext(ctx, query, preferredCurrency, time.Now(), telegramID)
	return err
}

func (p *profileRepository) SetPreferredLanguage(ctx context.Context, telegramID int64, preferredLanguage string) error {
	query := "UPDATE profile SET preferred_language = $1, updated_at = $2 WHERE telegram_id = $3"
	_, err := p.conn.ExecContext(ctx, query, preferredLanguage, time.Now(), telegramID)
	return err
}

func (p *profileRepository) TopUpBalance(ctx context.Context, telegramID int64, amount float64) error {
	query := "UPDATE profile SET balance = balance + $1, updated_at = $2 WHERE telegram_id = $3"
	_, err := p.conn.ExecContext(ctx, query, amount, time.Now(), telegramID)
	return err
}
