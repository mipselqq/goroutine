-- +goose Up
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    column_id UUID NOT NULL REFERENCES columns(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    position INT NOT NULL CHECK (position > 0),
    created_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    CONSTRAINT tasks_column_id_position_key UNIQUE (column_id, position) DEFERRABLE INITIALLY IMMEDIATE
);

-- +goose Down
DROP TABLE tasks;
