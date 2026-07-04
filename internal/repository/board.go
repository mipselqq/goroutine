package repository

import (
	"context"
	"errors"
	"fmt"

	"goroutine/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGBoard struct {
	pgPool *pgxpool.Pool
}

func NewPGBoard(pgPool *pgxpool.Pool) *PGBoard {
	return &PGBoard{
		pgPool: pgPool,
	}
}

func (r *PGBoard) Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
	const query = `INSERT INTO boards (owner_id, name, description) VALUES ($1, $2, $3) RETURNING id, owner_id, name, description, created_at, updated_at`

	var board domain.Board
	err := r.pgPool.QueryRow(ctx, query, ownerID, name, description).Scan(
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

func (r *PGBoard) Get(ctx context.Context, boardID domain.BoardID) (domain.Board, error) {
	const query = `
		SELECT id, owner_id, name, description, created_at, updated_at
		FROM boards
		WHERE id = $1`

	var board domain.Board
	err := r.pgPool.QueryRow(ctx, query, boardID).Scan(
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

func (r *PGBoard) List(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
	const query = `
		SELECT id, owner_id, name, description, created_at, updated_at
		FROM boards
		WHERE owner_id = $1
		ORDER BY created_at ASC`

	rows, err := r.pgPool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("board repo: list: %v: %w", err, ErrInternal)
	}
	defer rows.Close()

	var boards []domain.Board
	for rows.Next() {
		var board domain.Board
		err = rows.Scan(
			&board.ID,
			&board.OwnerID,
			&board.Name,
			&board.Description,
			&board.CreatedAt,
			&board.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("board repo: list: scan: %v: %w", err, ErrInternal)
		}

		boards = append(boards, board)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("board repo: list: rows final error: %v: %w", err, ErrInternal)
	}

	return boards, nil
}

func (r *PGBoard) Update(ctx context.Context, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
	const query = `
		UPDATE boards
		SET
			name = COALESCE($1, name),
			description = COALESCE($2, description),
			updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
		WHERE id = $3
		RETURNING id, owner_id, name, description, created_at, updated_at`

	var board domain.Board
	err := r.pgPool.QueryRow(ctx, query, name, description, boardID).Scan(
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
		return domain.Board{}, fmt.Errorf("board repo: update by id: %v: %w", err, ErrInternal)
	}

	return board, nil
}

func (r *PGBoard) Delete(ctx context.Context, boardID domain.BoardID) error {
	const query = `DELETE FROM boards WHERE id = $1`

	cmd, err := r.pgPool.Exec(ctx, query, boardID)
	if err != nil {
		return fmt.Errorf("board repo: delete: %v: %w", err, ErrInternal)
	}
	if cmd.RowsAffected() == 0 {
		return ErrRowNotFound
	}

	return nil
}
