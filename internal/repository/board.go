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
	return &PgBoard{
		pool: pool,
	}
}

func (r *PgBoard) Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
	const query = `INSERT INTO boards (owner_id, name, description) VALUES ($1, $2, $3) RETURNING id, owner_id, name, description, created_at, updated_at`

	var board domain.Board
	err := r.pool.QueryRow(ctx, query, ownerID, name, description).Scan(
		&board.ID,
		&board.OwnerID,
		&board.Name,
		&board.Description,
		&board.CreatedAt,
		&board.UpdatedAt,
	)
	if err != nil {
		return domain.Board{}, fmt.Errorf("board repo: create: %v: %w", err, ErrInternal)
	}

	return board, nil
}
