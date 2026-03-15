package repository

import (
	"context"
	"fmt"
	"time"

	"goroutine/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgBoard struct {
	pool *pgxpool.Pool
}

func NewPgBoard(pool *pgxpool.Pool) *PgBoard {
	return &PgBoard{pool: pool}
}

func (r *PgBoard) Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
	const query = `INSERT INTO boards (owner_id, name, description) VALUES ($1, $2, $3) RETURNING id, owner_id, name, description, created_at, updated_at`

	var idStr, ownerIDStr, nameStr, descStr string
	var createdAt, updatedAt time.Time

	err := r.pool.QueryRow(ctx, query, ownerID.String(), name.String(), description.String()).Scan(&idStr, &ownerIDStr, &nameStr, &descStr, &createdAt, &updatedAt)
	if err != nil {
		return domain.Board{}, fmt.Errorf("board repo: create: %v: %w", err, ErrInternal)
	}

	boardID, _ := domain.ParseBoardID(idStr)
	parsedOwnerID, _ := domain.ParseUserID(ownerIDStr)
	boardName, _ := domain.NewBoardName(nameStr)
	boardDesc, _ := domain.NewBoardDescription(descStr)

	return domain.Board{
		ID:          boardID,
		OwnerID:     parsedOwnerID,
		Name:        boardName,
		Description: boardDesc,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}
