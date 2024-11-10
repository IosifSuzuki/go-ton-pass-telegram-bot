CREATE TABLE IF NOT EXISTS temporal_workflow
(
    id SERIAL PRIMARY KEY,
    sms_history_id INT REFERENCES sms_history(id) ON DELETE CASCADE,
    temporal_id VARCHAR(64),
    temporal_run_id VARCHAR(64),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
