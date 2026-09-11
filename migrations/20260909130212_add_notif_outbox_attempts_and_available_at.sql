-- +goose Up
ALTER TABLE notification_outbox
    ADD COLUMN attempts INT NOT NULL DEFAULT 0,
    ADD COLUMN available_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC');

-- +goose Down
ALTER TABLE notification_outbox
    DROP COLUMN attempts,
    DROP COLUMN available_at;
