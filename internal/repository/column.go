package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"goroutine/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgColumn struct {
	pool *pgxpool.Pool
}

func NewPgColumn(pool *pgxpool.Pool) *PgColumn {
	return &PgColumn{pool: pool}
}

func (r *PgColumn) Create(
	ctx context.Context,
	boardID domain.BoardID,
	name domain.ColumnName,
	createdAt time.Time,
	updatedAt time.Time,
) (domain.Column, error) {
	const query = `
		WITH board_lock AS (
			SELECT id
			FROM boards
			WHERE id = $1
			FOR UPDATE
		), next_pos AS (
			SELECT COALESCE(MAX(position), 0) + 1 AS position
			FROM columns
			WHERE board_id = $1
		), inserted AS (
			INSERT INTO columns (board_id, name, position, created_at, updated_at)
			SELECT $1, $2, next_pos.position, $3, $4
			FROM board_lock, next_pos
			RETURNING id, board_id, name, position, created_at, updated_at
		)
		SELECT id, board_id, name, position, created_at, updated_at
		FROM inserted`

	var column domain.Column
	err := r.pool.QueryRow(ctx, query, boardID, name, createdAt, updatedAt).Scan(
		&column.ID,
		&column.BoardID,
		&column.Name,
		&column.Position,
		&column.CreatedAt,
		&column.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Column{}, ErrRowNotFound
		}
		return domain.Column{}, fmt.Errorf("column repo: create: %v: %w", err, ErrInternal)
	}

	return column, nil
}
