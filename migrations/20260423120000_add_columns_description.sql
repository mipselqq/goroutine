-- +goose Up
ALTER TABLE columns
    ADD COLUMN description TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE columns
    DROP COLUMN description;
