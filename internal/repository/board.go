package repository

import (
	"context"
	"errors"
	"fmt"

	"goroutine/internal/domain"

	"github.com/jackc/pgx/v5"
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

func (r *PgBoard) GetByID(ctx context.Context, id domain.BoardID) (domain.Board, error) {
	const query = `
		SELECT id, owner_id, name, description, created_at, updated_at
		FROM boards
		WHERE id = $1`

	var board domain.Board
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&board.ID,
		&board.OwnerID,
		&board.Name,
		&board.Description,
		&board.CreatedAt,
		&board.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Board{}, ErrRowNotFound
		}
		return domain.Board{}, fmt.Errorf("board repo: get by id: %v: %w", err, ErrInternal)
	}

	return board, nil
}

func (r *PgBoard) GetMany(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
	const query = `
		SELECT id, owner_id, name, description, created_at, updated_at
		FROM boards
		WHERE owner_id = $1
		ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("board repo: get many: %v: %w", err, ErrInternal)
	}
	defer rows.Close()

	var boards []domain.Board
	for rows.Next() {
		var board domain.Board
		err := rows.Scan(
			&board.ID,
			&board.OwnerID,
			&board.Name,
			&board.Description,
			&board.CreatedAt,
			&board.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("board repo: get many scan: %v: %w", err, ErrInternal)
		}

		boards = append(boards, board)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("board repo: get many rows: %v: %w", err, ErrInternal)
	}

	return boards, nil
}

func (r *PgBoard) Delete(ctx context.Context, boardID domain.BoardID) error {
	const query = `DELETE FROM boards WHERE id = $1`

	cmd, err := r.pool.Exec(ctx, query, boardID)
	if err != nil {
		return fmt.Errorf("board repo: delete: %v: %w", err, ErrInternal)
	}
	if cmd.RowsAffected() == 0 {
		return ErrRowNotFound
	}

	return nil
}
