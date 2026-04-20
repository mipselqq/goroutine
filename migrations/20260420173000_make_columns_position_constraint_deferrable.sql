-- +goose Up
ALTER TABLE columns
DROP CONSTRAINT columns_board_id_position_key;

ALTER TABLE columns
ADD CONSTRAINT columns_board_id_position_key
UNIQUE (board_id, position)
DEFERRABLE INITIALLY IMMEDIATE;

-- +goose Down
ALTER TABLE columns
DROP CONSTRAINT columns_board_id_position_key;

ALTER TABLE columns
ADD CONSTRAINT columns_board_id_position_key
UNIQUE (board_id, position);
