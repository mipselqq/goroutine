package repository

import (
	"context"
	"fmt"

	"goroutine/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgBoard struct {
	pool *pgxpool.Pool
}

func NewPgBoard(pool *pgxpool.Pool) *PgBoard {
	return &PgBoard{pool: pool}
}

func (r *PgBoard) Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) error {
	const query = `INSERT INTO boards (owner_id, name, description) VALUES ($1, $2, $3)`

	_, err := r.pool.Exec(ctx, query, ownerID.String(), name.String(), description.String())
	if err != nil {
		return fmt.Errorf("board repo: create: %v: %w", err, ErrInternal)
	}
	return nil
}
