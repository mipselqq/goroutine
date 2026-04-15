-- +goose Up
CREATE TABLE columns (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    position INT NOT NULL CHECK (position > 0),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE (board_id, position)
);

-- +goose Down
DROP TABLE columns;
