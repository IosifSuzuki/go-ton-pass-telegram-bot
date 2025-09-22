package repository

import (
	"context"
	"database/sql"
	"go-ton-pass-telegram-bot/internal/model/domain"
	"time"
)

type TelegramPaymentRepository interface {
	Create(ctx context.Context, telegramPayment *domain.TelegramPayment) (*int64, error)
	MarkRefunded(ctx context.Context, telegramPaymentChargeID string) error
	FetchByTelegramPaymentChargeID(ctx context.Context, telegramPaymentChargeID string) (*domain.TelegramPayment, error)
}

type telegramPaymentRepository struct {
	conn *sql.DB
}

func NewTelegramPaymentRepository(conn *sql.DB) TelegramPaymentRepository {
	return &telegramPaymentRepository{
		conn: conn,
	}
}

func (t *telegramPaymentRepository) Create(ctx context.Context, telegramPayment *domain.TelegramPayment) (*int64, error) {
	query := "INSERT INTO telegram_payment (profile_id, telegram_payment_charge_id, currency, amount, credit_amount) " +
		"VALUES ($1, $2, $3, $4, $5) " +
		"RETURNING id;"
	var id int64
	err := t.conn.QueryRowContext(
		ctx,
		query,
		telegramPayment.ProfileID,
		telegramPayment.TelegramPaymentChargeID,
		telegramPayment.Currency,
		telegramPayment.Amount,
		telegramPayment.CreditAmount,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &id, err
}

func (t *telegramPaymentRepository) MarkRefunded(ctx context.Context, telegramPaymentChargeID string) error {
	query := "UPDATE telegram_payment SET is_refunded = $1, updated_at = $2 WHERE telegram_payment_charge_id = $3"
	_, err := t.conn.ExecContext(ctx, query, true, time.Now(), telegramPaymentChargeID)
	return err
}

func (t *telegramPaymentRepository) FetchByTelegramPaymentChargeID(ctx context.Context, telegramPaymentChargeID string) (*domain.TelegramPayment, error) {
	query := "SELECT profile_id, currency, amount, credit_amount, is_refunded, created_at, updated_at " +
		"FROM telegram_payment WHERE telegram_payment_charge_id = $1;"
	var telegramPaymentPayload = domain.TelegramPayment{
		TelegramPaymentChargeID: telegramPaymentChargeID,
		CreatedAt:               new(time.Time),
		UpdatedAt:               new(time.Time),
	}
	row := t.conn.QueryRowContext(ctx, query, telegramPaymentChargeID)
	err := row.Scan(
		&telegramPaymentPayload.ProfileID,
		&telegramPaymentPayload.Currency,
		&telegramPaymentPayload.Amount,
		&telegramPaymentPayload.CreditAmount,
		&telegramPaymentPayload.IsRefunded,
		&telegramPaymentPayload.CreatedAt,
		&telegramPaymentPayload.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &telegramPaymentPayload, nil
}
