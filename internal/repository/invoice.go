package repository

import (
	"context"
	"database/sql"
	"go-ton-pass-telegram-bot/internal/model/domain"
	"time"
)

type InvoiceRepository interface {
	Create(ctx context.Context, invoice *domain.Invoice) (*int64, error)
	GetInvoiceByInvoiceID(ctx context.Context, invoiceID int64) (*domain.Invoice, error)
	UpdateStatus(ctx context.Context, invoiceID int64, status string) error
}

type invoiceRepository struct {
	conn *sql.DB
}

func NewInvoiceRepository(conn *sql.DB) InvoiceRepository {
	return &invoiceRepository{
		conn: conn,
	}
}

func (i *invoiceRepository) Create(ctx context.Context, invoice *domain.Invoice) (*int64, error) {
	query := "INSERT INTO invoice (profile_id, chat_id, invoice_id, status, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id;"
	var id int64
	err := i.conn.QueryRowContext(
		ctx,
		query,
		invoice.ProfileID,
		invoice.ChatID,
		invoice.InvoiceID,
		invoice.Status,
		time.Now(),
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &id, err
}

func (i *invoiceRepository) GetInvoiceByInvoiceID(ctx context.Context, invoiceID int64) (*domain.Invoice, error) {
	query := "SELECT id, profile_id, chat_id, status, created_at, updated_at FROM invoice WHERE invoice_id = $1"
	row := i.conn.QueryRowContext(ctx, query, invoiceID)
	profile := domain.Invoice{
		InvoiceID: invoiceID,
	}
	var updatedAt sql.NullTime

	err := row.Scan(
		&profile.ID,
		&profile.ProfileID,
		&profile.ChatID,
		&profile.Status,
		&profile.CreatedAt,
		&updatedAt,
	)
	if updatedAt.Valid {
		profile.UpdatedAt = &updatedAt.Time
	}
	return &profile, err
}

func (i *invoiceRepository) UpdateStatus(ctx context.Context, invoiceID int64, status string) error {
	query := "UPDATE invoice SET status = $1, updated_at = $2 WHERE invoice_id = $3"
	_, err := i.conn.ExecContext(ctx, query, status, time.Now(), invoiceID)
	return err
}
