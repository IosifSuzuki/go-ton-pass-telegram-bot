CREATE TABLE IF NOT EXISTS profile (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE,
    telegram_chat_id BIGINT,
    username TEXT,
    preferred_currency VARCHAR(127),
    preferred_language VARCHAR(127),
    balance DOUBLE PRECISION,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);