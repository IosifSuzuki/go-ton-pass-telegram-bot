CREATE TABLE IF NOT EXISTS profile (
    id SERIAL PRIMARY KEY,
    telegram_id INT UNIQUE,
    username TEXT,
    preferred_currency VARCHAR(127),
    preferred_language VARCHAR(127),
    balance INT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);