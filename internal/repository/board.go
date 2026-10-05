package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"goroutine/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGBoard struct {
	pgPool *pgxpool.Pool
}

const boardSelectQuery = `
	SELECT id, owner_id, name, description, created_at, updated_at
	FROM boards
	WHERE id = @board_id
	  AND owner_id = @caller_id`

func NewPGBoard(pgPool *pgxpool.Pool) *PGBoard {
	return &PGBoard{
		pgPool: pgPool,
	}
}

func (r *PGBoard) Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
	const query = `
		WITH created_board AS (
			INSERT INTO boards (owner_id, name, description)
			VALUES ($1, $2, $3)
			RETURNING id, owner_id, name, description, created_at, updated_at
		),
		created_event AS (
			INSERT INTO notification_outbox (recipient_user_id, event_type, payload)
			SELECT
				b.owner_id,
				$4,
				jsonb_build_object(
					'callerEmail', u.email,
					'boardName', b.name
				)
			FROM created_board b
			JOIN users u ON u.id = b.owner_id
			WHERE u.telegram_chat_id IS NOT NULL
		)
		SELECT b.id, b.owner_id, b.name, b.description, b.created_at, b.updated_at
		FROM created_board b`

	board, err := ScanBoard(r.pgPool.QueryRow(ctx, query, ownerID, name, description, domain.TypeBoardCreated))
	if err != nil {
		return domain.Board{}, fmt.Errorf("board repo: create: %v: %w", err, ErrInternal)
	}

	return board, nil
}

func (r *PGBoard) Get(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) (domain.Board, error) {
	board, err := ScanBoard(r.pgPool.QueryRow(ctx, boardSelectQuery, pgx.NamedArgs{
		"caller_id": callerID,
		"board_id":  boardID,
	}))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Board{}, ErrRowNotFound
		}
		return domain.Board{}, fmt.Errorf("board repo: get: %v: %w", err, ErrInternal)
	}

	return board, nil
}

func (r *PGBoard) GetAggregate(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
) (domain.Board, []domain.Column, []domain.Task, error) {
	const taskSelectByBoardIDQuery = `
		SELECT t.id, t.column_id, t.name, t.description, t.position, t.created_at, t.updated_at
		FROM boards b
		JOIN columns c ON c.board_id = b.id
		JOIN tasks t ON t.column_id = c.id
		WHERE b.id = @board_id
		  AND b.owner_id = @caller_id
		ORDER BY c.position ASC, t.position ASC`

	args := pgx.NamedArgs{
		"caller_id": callerID,
		"board_id":  boardID,
	}
	batch := &pgx.Batch{}
	batch.Queue(boardSelectQuery, args)
	batch.Queue(listColumnsByBoardIDQuery, args)
	batch.Queue(taskSelectByBoardIDQuery, args)

	results := r.pgPool.SendBatch(ctx, batch)
	defer func() {
		_ = results.Close()
	}()

	board, err := ScanBoard(results.QueryRow())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Board{}, nil, nil, ErrRowNotFound
		}
		return domain.Board{}, nil, nil, fmt.Errorf("board repo: get aggregate board: %v: %w", err, ErrInternal)
	}

	columnRows, err := results.Query()
	if err != nil {
		return domain.Board{}, nil, nil, fmt.Errorf("board repo: get aggregate columns: %v: %w", err, ErrInternal)
	}
	var columns []domain.Column
	for columnRows.Next() {
		column, scanErr := ScanColumn(columnRows)
		if scanErr != nil {
			columnRows.Close()
			return domain.Board{}, nil, nil, fmt.Errorf("board repo: get aggregate columns scan: %v: %w", scanErr, ErrInternal)
		}
		columns = append(columns, column)
	}
	err = columnRows.Err()
	columnRows.Close()
	if err != nil {
		return domain.Board{}, nil, nil, fmt.Errorf("board repo: get aggregate columns rows final error: %v: %w", err, ErrInternal)
	}

	taskRows, err := results.Query()
	if err != nil {
		return domain.Board{}, nil, nil, fmt.Errorf("board repo: get aggregate tasks: %v: %w", err, ErrInternal)
	}
	var tasks []domain.Task
	for taskRows.Next() {
		task, scanErr := ScanTask(taskRows)
		if scanErr != nil {
			taskRows.Close()
			return domain.Board{}, nil, nil, fmt.Errorf("board repo: get aggregate tasks scan: %v: %w", scanErr, ErrInternal)
		}
		tasks = append(tasks, task)
	}
	err = taskRows.Err()
	taskRows.Close()
	if err != nil {
		return domain.Board{}, nil, nil, fmt.Errorf("board repo: get aggregate tasks rows final error: %v: %w", err, ErrInternal)
	}

	err = results.Close()
	if err != nil {
		return domain.Board{}, nil, nil, fmt.Errorf("board repo: get aggregate close batch: %v: %w", err, ErrInternal)
	}

	return board, columns, tasks, nil
}

func (r *PGBoard) ListByOwnerID(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
	const query = `
		SELECT id, owner_id, name, description, created_at, updated_at
		FROM boards
		WHERE owner_id = $1
		ORDER BY created_at ASC`

	rows, err := r.pgPool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("board repo: list by owner id: %v: %w", err, ErrInternal)
	}
	defer rows.Close()

	var boards []domain.Board
	for rows.Next() {
		board, scanErr := ScanBoard(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("board repo: list by owner id: scan: %v: %w", scanErr, ErrInternal)
		}

		boards = append(boards, board)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("board repo: list by owner id: rows final error: %v: %w", err, ErrInternal)
	}

	return boards, nil
}

func (r *PGBoard) Update(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	name *domain.BoardName,
	description *domain.BoardDescription,
) (domain.Board, error) {
	const query = `
		WITH updated_board AS (
			UPDATE boards
			SET
				name = COALESCE($1, name),
				description = COALESCE($2, description),
				updated_at = CASE
					WHEN $1 IS NULL
					 AND $2 IS NULL
					THEN updated_at
					ELSE CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
				END
			WHERE id = $3
			  AND owner_id = $4
			RETURNING id, owner_id, name, description, created_at, updated_at
		),
		created_event AS (
			INSERT INTO notification_outbox (recipient_user_id, event_type, payload)
			SELECT
				b.owner_id,
				$5,
				jsonb_build_object(
					'callerEmail', u.email,
					'boardName', b.name
				)
			FROM updated_board b
			JOIN users u ON u.id = b.owner_id
			WHERE ($1 IS NOT NULL OR $2 IS NOT NULL)
			  AND u.telegram_chat_id IS NOT NULL
		)
		SELECT b.id, b.owner_id, b.name, b.description, b.created_at, b.updated_at
		FROM updated_board b`

	board, err := ScanBoard(r.pgPool.QueryRow(ctx, query, name, description, boardID, callerID, domain.TypeBoardUpdated))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Board{}, ErrRowNotFound
		}
		return domain.Board{}, fmt.Errorf("board repo: update: %v: %w", err, ErrInternal)
	}

	return board, nil
}

func (r *PGBoard) Delete(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) error {
	const query = `
		WITH deleted_board AS (
			DELETE FROM boards
			WHERE id = $1
			  AND owner_id = $2
			RETURNING id, owner_id, name
		),
		created_event AS (
			INSERT INTO notification_outbox (recipient_user_id, event_type, payload)
			SELECT
				b.owner_id,
				$3,
				jsonb_build_object(
					'callerEmail', u.email,
					'boardName', b.name
				)
			FROM deleted_board b
			JOIN users u ON u.id = b.owner_id
			WHERE u.telegram_chat_id IS NOT NULL
		)
		SELECT b.id
		FROM deleted_board b`

	var deletedBoardID uuid.UUID
	err := r.pgPool.QueryRow(ctx, query, boardID, callerID, domain.TypeBoardDeleted).Scan(&deletedBoardID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRowNotFound
		}
		return fmt.Errorf("board repo: delete: %v: %w", err, ErrInternal)
	}

	return nil
}

func ScanBoard(row interface{ Scan(...any) error }) (domain.Board, error) {
	var (
		rawID      uuid.UUID
		rawOwnerID uuid.UUID
		rawName    string
		rawDesc    string
		createdAt  time.Time
		updatedAt  time.Time
	)
	err := row.Scan(&rawID, &rawOwnerID, &rawName, &rawDesc, &createdAt, &updatedAt)
	if err != nil {
		return domain.Board{}, fmt.Errorf("scan board: %w", err)
	}
	name, err := domain.NewBoardName(rawName)
	if err != nil {
		return domain.Board{}, fmt.Errorf("scan board: name: %v: %w", err, errDataCorrupted)
	}
	desc, err := domain.NewBoardDescription(rawDesc)
	if err != nil {
		return domain.Board{}, fmt.Errorf("scan board: description: %v: %w", err, errDataCorrupted)
	}
	id, err := domain.NewBoardIDFromUUID(rawID)
	if err != nil {
		return domain.Board{}, fmt.Errorf("scan board: id: %v: %w", err, errDataCorrupted)
	}
	ownerID, err := domain.NewUserIDFromUUID(rawOwnerID)
	if err != nil {
		return domain.Board{}, fmt.Errorf("scan board: owner id: %v: %w", err, errDataCorrupted)
	}
	return domain.Board{
		ID:          id,
		OwnerID:     ownerID,
		Name:        name,
		Description: desc,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}
