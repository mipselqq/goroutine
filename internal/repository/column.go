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

type PGColumn struct {
	pgPool *pgxpool.Pool
}

func NewPGColumn(pgPool *pgxpool.Pool) *PGColumn {
	return &PGColumn{pgPool: pgPool}
}

const listColumnsByBoardIDQuery = `
	SELECT c.id, c.board_id, c.name, c.description, c.position, c.created_at, c.updated_at
	FROM boards b
	JOIN columns c ON c.board_id = b.id
	WHERE b.id = @board_id
	  AND b.owner_id = @caller_id
	ORDER BY c.position ASC`

func (r *PGColumn) Create(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	name domain.ColumnName,
	description domain.ColumnDescription,
) (domain.Column, error) {
	const (
		beginQuery     = `BEGIN`
		lockBoardQuery = `
		SELECT 1
		FROM boards
		WHERE id = @board_id
		  AND owner_id = @caller_id
		FOR UPDATE`
		insertColumnQuery = `
		WITH created_column AS (
			INSERT INTO columns (board_id, name, description, position)
			SELECT
			    b.id,
			    @name,
			    @description,
			    COALESCE(MAX(c.position), 0) + 1
			FROM boards b
			LEFT JOIN columns c ON c.board_id = b.id
			WHERE b.id = @board_id
			  AND b.owner_id = @caller_id
			GROUP BY b.id
			RETURNING id, board_id, name, description, position, created_at, updated_at
		),
		created_event AS (
			INSERT INTO notification_outbox (recipient_user_id, event_type, payload)
			SELECT
				b.owner_id,
				@event_type,
				jsonb_build_object(
					'callerEmail', u.email,
					'boardName', b.name,
					'columnName', c.name
				)
			FROM created_column c
			JOIN boards b ON b.id = c.board_id
			JOIN users u ON u.id = b.owner_id
			WHERE u.telegram_chat_id IS NOT NULL
		)
		SELECT c.id, c.board_id, c.name, c.description, c.position, c.created_at, c.updated_at
		FROM created_column c`
		commitQuery = `COMMIT`
	)

	args := pgx.NamedArgs{
		"caller_id":   callerID,
		"board_id":    boardID,
		"name":        name,
		"description": description,
		"event_type":  domain.TypeColumnCreated,
	}
	batch := &pgx.Batch{}
	batch.Queue(beginQuery)
	batch.Queue(lockBoardQuery, args)
	batch.Queue(insertColumnQuery, args)
	batch.Queue(commitQuery)

	results := r.pgPool.SendBatch(ctx, batch)
	defer func() {
		_ = results.Close()
	}()

	_, err := results.Exec()
	if err != nil {
		return domain.Column{}, fmt.Errorf("column repo: create begin tx: %v: %w", err, ErrInternal)
	}
	_, err = results.Exec()
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Column{}, ErrRowNotFound
		}
		return domain.Column{}, fmt.Errorf("column repo: create lock board: %v: %w", err, ErrInternal)
	}

	column, err := ScanColumn(results.QueryRow())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Column{}, ErrRowNotFound
		}
		return domain.Column{}, fmt.Errorf("column repo: create insert: %v: %w", err, ErrInternal)
	}
	_, err = results.Exec()
	if err != nil {
		return domain.Column{}, fmt.Errorf("column repo: create commit: %v: %w", err, ErrInternal)
	}
	if err = results.Close(); err != nil {
		return domain.Column{}, fmt.Errorf("column repo: create close batch: %v: %w", err, ErrInternal)
	}

	return column, nil
}

func (r *PGColumn) ListByBoardID(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
) ([]domain.Column, error) {
	const boardQuery = `
		SELECT 1
		FROM boards
		WHERE id = @board_id
		  AND owner_id = @caller_id`

	args := pgx.NamedArgs{
		"caller_id": callerID,
		"board_id":  boardID,
	}
	batch := &pgx.Batch{}
	batch.Queue(boardQuery, args)
	batch.Queue(listColumnsByBoardIDQuery, args)

	results := r.pgPool.SendBatch(ctx, batch)
	defer func() {
		_ = results.Close()
	}()

	var found int
	err := results.QueryRow().Scan(&found)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRowNotFound
		}
		return nil, fmt.Errorf("column repo: list by board id check board: %v: %w", err, ErrInternal)
	}

	rows, err := results.Query()
	if err != nil {
		return nil, fmt.Errorf("column repo: list by board id: %v: %w", err, ErrInternal)
	}
	defer rows.Close()

	var result []domain.Column
	for rows.Next() {
		col, scanErr := ScanColumn(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("column repo: list by board id: scan: %v: %w", scanErr, ErrInternal)
		}
		result = append(result, col)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("column repo: list by board id: rows final error: %v: %w", err, ErrInternal)
	}
	if err = results.Close(); err != nil {
		return nil, fmt.Errorf("column repo: list by board id close batch: %v: %w", err, ErrInternal)
	}

	return result, nil
}

func (r *PGColumn) Get(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	columnID domain.ColumnID,
) (domain.Column, error) {
	const query = `
		SELECT c.id, c.board_id, c.name, c.description, c.position, c.created_at, c.updated_at
		FROM boards b
		JOIN columns c ON c.board_id = b.id
		WHERE b.id = $1
		  AND b.owner_id = $2
		  AND c.id = $3`

	column, err := ScanColumn(r.pgPool.QueryRow(ctx, query, boardID, callerID, columnID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Column{}, ErrRowNotFound
		}
		return domain.Column{}, fmt.Errorf("column repo: get: %v: %w", err, ErrInternal)
	}

	return column, nil
}

func (r *PGColumn) Update(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	name *domain.ColumnName,
	description *domain.ColumnDescription,
) (domain.Column, error) {
	const query = `
		WITH updated_column AS (
			UPDATE columns c
			SET
				name = COALESCE($1, c.name),
				description = COALESCE($2, c.description),
				updated_at = CASE
					WHEN $1 IS NULL
					 AND $2 IS NULL
					THEN c.updated_at
					ELSE CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
				END
			FROM boards b
			WHERE c.board_id = $3
			  AND c.id = $4
			  AND b.id = c.board_id
			  AND b.owner_id = $5
			RETURNING c.id, c.board_id, c.name, c.description, c.position, c.created_at, c.updated_at
		),
		created_event AS (
			INSERT INTO notification_outbox (recipient_user_id, event_type, payload)
			SELECT
				b.owner_id,
				$6,
				jsonb_build_object(
					'callerEmail', u.email,
					'boardName', b.name,
					'columnName', c.name
				)
			FROM updated_column c
			JOIN boards b ON b.id = c.board_id
			JOIN users u ON u.id = b.owner_id
			WHERE ($1 IS NOT NULL OR $2 IS NOT NULL)
			  AND u.telegram_chat_id IS NOT NULL
		)
		SELECT c.id, c.board_id, c.name, c.description, c.position, c.created_at, c.updated_at
		FROM updated_column c`

	column, err := ScanColumn(r.pgPool.QueryRow(ctx, query, name, description, boardID, columnID, callerID, domain.TypeColumnUpdated))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Column{}, ErrRowNotFound
		}
		return domain.Column{}, fmt.Errorf("column repo: update: %v: %w", err, ErrInternal)
	}

	return column, nil
}

func (r *PGColumn) Move(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	targetPosition domain.ColumnPosition,
) (domain.ColumnPosition, error) {
	const (
		// SET position order is not guaranteed, so we disable uniqueness constraint for this transaction.

		// 1. Lock the board row so no concurrent operation can reorder columns in the same board.
		lockBoardQuery = `
		SELECT 1
		FROM boards
		WHERE id = @board_id
		  AND owner_id = @caller_id
		FOR UPDATE`

		// 2. Defer the unique constraint until COMMIT for this transaction only.
		deferPositionConstraintQuery = `
		SET CONSTRAINTS columns_board_id_position_key DEFERRED`

		// 3. Read the current position of the column we are moving.
		getCurrentPositionQuery = `
		SELECT position
		FROM columns
		WHERE board_id = @board_id
		  AND id = @column_id`

		// 4. Read how many columns the board currently has to validate targetPosition.
		countColumnsQuery = `
		SELECT COUNT(*)
		FROM columns
		WHERE board_id = @board_id`

		// 5. If the moved column goes down, shift neighbors from (current, target] one slot up.
		//    Example: moving 2 -> 5 means 3,4,5 become 2,3,4.
		moveNeighborsDownQuery = `
		UPDATE columns
		SET position = position - 1
		WHERE board_id = @board_id
		  AND position > @current_position
		  AND position <= @target_position`

		// 5. If the moved column goes up, shift neighbors from [target, current) one slot down.
		//    Example: moving 5 -> 2 means 2,3,4 become 3,4,5.
		moveNeighborsUpQuery = `
		UPDATE columns
		SET position = position + 1
		WHERE board_id = @board_id
		  AND position >= @target_position
		  AND position < @current_position`

		// 6. Put the moved column into targetPosition after neighbors have been shifted.
		moveColumnIntoTargetQuery = `
		UPDATE columns
		SET position = @target_position
		WHERE board_id = @board_id
		  AND id = @column_id`

		insertMovedEventQuery = `
		INSERT INTO notification_outbox (recipient_user_id, event_type, payload)
		SELECT
			b.owner_id,
			@event_type,
			jsonb_build_object(
				'callerEmail', u.email,
				'boardName', b.name,
				'columnName', c.name,
				'sourcePosition', @current_position::integer,
				'targetPosition', c.position
			)
		FROM columns c
		JOIN boards b ON b.id = c.board_id
		JOIN users u ON u.id = b.owner_id
		WHERE c.board_id = @board_id
		  AND c.id = @column_id
		  AND u.telegram_chat_id IS NOT NULL`
	)

	tx, err := r.pgPool.Begin(ctx)
	if err != nil {
		return domain.ColumnPosition{}, fmt.Errorf("column repo: move begin tx: %v: %w", err, ErrInternal)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var locked int
	err = tx.QueryRow(ctx, lockBoardQuery, pgx.NamedArgs{
		"caller_id": callerID,
		"board_id":  boardID,
	}).Scan(&locked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ColumnPosition{}, ErrRowNotFound
		}
		return domain.ColumnPosition{}, fmt.Errorf("column repo: move lock board: %v: %w", err, ErrInternal)
	}

	_, err = tx.Exec(ctx, deferPositionConstraintQuery)
	if err != nil {
		return domain.ColumnPosition{}, fmt.Errorf("column repo: move defer position constraint: %v: %w", err, ErrInternal)
	}

	var currentPosition int64
	err = tx.QueryRow(ctx, getCurrentPositionQuery, pgx.NamedArgs{
		"board_id":  boardID.UUID(),
		"column_id": columnID.UUID(),
	}).Scan(&currentPosition)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ColumnPosition{}, ErrRowNotFound
		}
		return domain.ColumnPosition{}, fmt.Errorf("column repo: move get current position: %v: %w", err, ErrInternal)
	}

	var columnsCount int64
	err = tx.QueryRow(ctx, countColumnsQuery, pgx.NamedArgs{
		"board_id": boardID.UUID(),
	}).Scan(&columnsCount)
	if err != nil {
		return domain.ColumnPosition{}, fmt.Errorf("column repo: move count columns: %v: %w", err, ErrInternal)
	}

	targetPositionInt := targetPosition.Int64()
	if targetPositionInt > columnsCount {
		return domain.ColumnPosition{}, ErrIndexOutOfBounds
	}
	if targetPositionInt == currentPosition {
		return targetPosition, nil
	}

	moveNeighborsArgs := pgx.NamedArgs{
		"board_id":         boardID,
		"current_position": currentPosition,
		"target_position":  targetPositionInt,
	}
	if currentPosition < targetPositionInt {
		_, err = tx.Exec(ctx, moveNeighborsDownQuery, moveNeighborsArgs)
		if err != nil {
			return domain.ColumnPosition{}, fmt.Errorf("column repo: move neighbors down: %v: %w", err, ErrInternal)
		}
	} else {
		_, err = tx.Exec(ctx, moveNeighborsUpQuery, moveNeighborsArgs)
		if err != nil {
			return domain.ColumnPosition{}, fmt.Errorf("column repo: move neighbors up: %v: %w", err, ErrInternal)
		}
	}

	_, err = tx.Exec(ctx, moveColumnIntoTargetQuery, pgx.NamedArgs{
		"board_id":        boardID,
		"column_id":       columnID,
		"target_position": targetPosition,
	})
	if err != nil {
		return domain.ColumnPosition{}, fmt.Errorf("column repo: move column into target: %v: %w", err, ErrInternal)
	}

	_, err = tx.Exec(ctx, insertMovedEventQuery, pgx.NamedArgs{
		"board_id":         boardID,
		"column_id":        columnID,
		"current_position": currentPosition,
		"event_type":       domain.TypeColumnMoved,
	})
	if err != nil {
		return domain.ColumnPosition{}, fmt.Errorf("column repo: move insert outbox event: %v: %w", err, ErrInternal)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return domain.ColumnPosition{}, fmt.Errorf("column repo: move commit: %v: %w", err, ErrInternal)
	}

	return targetPosition, nil
}

func (r *PGColumn) Delete(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	columnID domain.ColumnID,
) error {
	const (
		// 1. Lock the board row so no concurrent operation can reorder columns in the same board.
		lockBoardQuery = `
		SELECT 1
		FROM boards
		WHERE id = @board_id
		  AND owner_id = @caller_id
		FOR UPDATE`

		// 2. Defer the unique constraint until COMMIT for this transaction only.
		deferPositionConstraintQuery = `
		SET CONSTRAINTS columns_board_id_position_key DEFERRED`

		// 3. Delete the target column and remember its position.
		deleteColumnQuery = `
		DELETE FROM columns
		WHERE board_id = @board_id
		  AND id = @column_id
		RETURNING name, position`

		// 4. Close the gap left by the deleted column.
		compactTrailingColumnsQuery = `
		UPDATE columns
		SET position = position - 1
		WHERE board_id = @board_id
		  AND position > @deleted_position`

		insertDeletedEventQuery = `
		INSERT INTO notification_outbox (recipient_user_id, event_type, payload)
		SELECT
			b.owner_id,
			@event_type,
			jsonb_build_object(
				'callerEmail', u.email,
				'boardName', b.name,
				'columnName', @column_name::text
			)
		FROM boards b
		JOIN users u ON u.id = b.owner_id
		WHERE b.id = @board_id
		  AND u.telegram_chat_id IS NOT NULL`
	)

	tx, err := r.pgPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("column repo: delete begin tx: %v: %w", err, ErrInternal)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var locked int
	err = tx.QueryRow(ctx, lockBoardQuery, pgx.NamedArgs{
		"caller_id": callerID,
		"board_id":  boardID,
	}).Scan(&locked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRowNotFound
		}
		return fmt.Errorf("column repo: delete lock board: %v: %w", err, ErrInternal)
	}

	_, err = tx.Exec(ctx, deferPositionConstraintQuery)
	if err != nil {
		return fmt.Errorf("column repo: delete defer position constraint: %v: %w", err, ErrInternal)
	}

	var (
		deletedName     string
		deletedPosition int64
	)
	err = tx.QueryRow(ctx, deleteColumnQuery, pgx.NamedArgs{
		"board_id":  boardID.UUID(),
		"column_id": columnID.UUID(),
	}).Scan(&deletedName, &deletedPosition)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRowNotFound
		}
		return fmt.Errorf("column repo: delete column: %v: %w", err, ErrInternal)
	}

	_, err = tx.Exec(ctx, compactTrailingColumnsQuery, pgx.NamedArgs{
		"board_id":         boardID,
		"deleted_position": deletedPosition,
	})
	if err != nil {
		return fmt.Errorf("column repo: delete compact trailing columns: %v: %w", err, ErrInternal)
	}

	_, err = tx.Exec(ctx, insertDeletedEventQuery, pgx.NamedArgs{
		"board_id":    boardID,
		"column_name": deletedName,
		"event_type":  domain.TypeColumnDeleted,
	})
	if err != nil {
		return fmt.Errorf("column repo: delete insert outbox event: %v: %w", err, ErrInternal)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("column repo: delete commit: %v: %w", err, ErrInternal)
	}

	return nil
}

func ScanColumn(row interface{ Scan(...any) error }) (domain.Column, error) {
	var (
		rawID      uuid.UUID
		rawBoardID uuid.UUID
		rawName    string
		rawDesc    string
		rawPos     int64
		createdAt  time.Time
		updatedAt  time.Time
	)
	err := row.Scan(&rawID, &rawBoardID, &rawName, &rawDesc, &rawPos, &createdAt, &updatedAt)
	if err != nil {
		return domain.Column{}, fmt.Errorf("scan column: %w", err)
	}
	name, err := domain.NewColumnName(rawName)
	if err != nil {
		return domain.Column{}, fmt.Errorf("scan column: name: %v: %w", err, errDataCorrupted)
	}
	desc, err := domain.NewColumnDescription(rawDesc)
	if err != nil {
		return domain.Column{}, fmt.Errorf("scan column: description: %v: %w", err, errDataCorrupted)
	}
	pos, err := domain.NewColumnPosition(rawPos)
	if err != nil {
		return domain.Column{}, fmt.Errorf("scan column: position: %v: %w", err, errDataCorrupted)
	}
	id, err := domain.NewColumnIDFromUUID(rawID)
	if err != nil {
		return domain.Column{}, fmt.Errorf("scan column: id: %v: %w", err, errDataCorrupted)
	}
	boardID, err := domain.NewBoardIDFromUUID(rawBoardID)
	if err != nil {
		return domain.Column{}, fmt.Errorf("scan column: board id: %v: %w", err, errDataCorrupted)
	}
	return domain.Column{
		ID:          id,
		BoardID:     boardID,
		Name:        name,
		Description: desc,
		Position:    pos,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}
