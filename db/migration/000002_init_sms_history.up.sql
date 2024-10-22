CREATE TABLE IF NOT EXISTS sms_history
(
    id SERIAL PRIMARY KEY,
    profile_id INT REFERENCES profile(id) ON DELETE CASCADE,
    activation_id BIGINT,
    service_code VARCHAR(64),
    phone_number VARCHAR(64),
    status  VARCHAR(64),
    sms_text TEXT,
    sms_code VARCHAR(64),
    received_at TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);