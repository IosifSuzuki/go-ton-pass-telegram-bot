CREATE TABLE IF NOT EXISTS invoice (
    id SERIAL PRIMARY KEY,
    profile_id INT,
    chat_id INT,
    invoice_id INT,
    status VARCHAR(64),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);