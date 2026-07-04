-- +goose Up
ALTER TABLE users
    ADD COLUMN telegram_chat_id BIGINT,
    ADD COLUMN telegram_username TEXT DEFAULT NULL;

-- +goose Down
ALTER TABLE users
    DROP COLUMN telegram_chat_id,
    DROP COLUMN telegram_username;
