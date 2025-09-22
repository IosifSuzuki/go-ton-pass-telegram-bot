CREATE TABLE IF NOT EXISTS telegram_payment (
    id SERIAL PRIMARY KEY,
    profile_id INT REFERENCES profile(id) ON DELETE CASCADE,
    telegram_payment_charge_id TEXT NOT NULL,
    currency VARCHAR(16) NOT NULL,
    amount BIGINT NOT NULL,
    credit_amount DOUBLE PRECISION NOT NULL,
    is_refunded BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);