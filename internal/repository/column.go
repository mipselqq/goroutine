package repository

import (
	"context"
	"errors"
	"fmt"

	"goroutine/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGColumn struct {
	pgPool *pgxpool.Pool
}

func NewPGColumn(pgPool *pgxpool.Pool) *PGColumn {
	return &PGColumn{pgPool: pgPool}
}

func (r *PGColumn) Create(
	ctx context.Context,
	boardID domain.BoardID,
	name domain.ColumnName,
	description domain.ColumnDescription,
) (domain.Column, error) {
	const (
		lockBoardQuery = `
		SELECT 1
		FROM boards
		WHERE id = @board_id
		FOR UPDATE`
		nextPositionQuery = `
		SELECT COALESCE(MAX(position), 0) + 1
		FROM columns
		WHERE board_id = @board_id`
		insertColumnQuery = `
		INSERT INTO columns (board_id, name, description, position)
		VALUES (@board_id, @name, @description, @position)
		RETURNING id, board_id, name, description, position, created_at, updated_at`
	)

	tx, err := r.pgPool.Begin(ctx)
	if err != nil {
		return domain.Column{}, fmt.Errorf("column repo: create begin tx: %v: %w", err, ErrInternal)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var locked int
	err = tx.QueryRow(ctx, lockBoardQuery, pgx.NamedArgs{
		"board_id": boardID,
	}).Scan(&locked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Column{}, ErrRowNotFound
		}
		return domain.Column{}, fmt.Errorf("column repo: create lock board: %v: %w", err, ErrInternal)
	}

	var nextPosition int64
	err = tx.QueryRow(ctx, nextPositionQuery, pgx.NamedArgs{
		"board_id": boardID,
	}).Scan(&nextPosition)
	if err != nil {
		return domain.Column{}, fmt.Errorf("column repo: create next position: %v: %w", err, ErrInternal)
	}

	var column domain.Column
	err = tx.QueryRow(ctx, insertColumnQuery, pgx.NamedArgs{
		"board_id":    boardID,
		"name":        name,
		"description": description,
		"position":    nextPosition,
	}).Scan(
		&column.ID,
		&column.BoardID,
		&column.Name,
		&column.Description,
		&column.Position,
		&column.CreatedAt,
		&column.UpdatedAt,
	)
	if err != nil {
		return domain.Column{}, fmt.Errorf("column repo: create insert: %v: %w", err, ErrInternal)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return domain.Column{}, fmt.Errorf("column repo: create commit: %v: %w", err, ErrInternal)
	}

	return column, nil
}

func (r *PGColumn) List(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error) {
	const query = `
		SELECT id, board_id, name, description, position, created_at, updated_at
		FROM columns
		WHERE board_id = $1
		ORDER BY position ASC`

	rows, err := r.pgPool.Query(ctx, query, boardID)
	if err != nil {
		return nil, fmt.Errorf("column repo: list: %v: %w", err, ErrInternal)
	}
	defer rows.Close()

	var result []domain.Column
	for rows.Next() {
		var col domain.Column
		err = rows.Scan(
			&col.ID,
			&col.BoardID,
			&col.Name,
			&col.Description,
			&col.Position,
			&col.CreatedAt,
			&col.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("column repo: list: scan: %v: %w", err, ErrInternal)
		}
		result = append(result, col)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("column repo: list: rows final error: %v: %w", err, ErrInternal)
	}

	return result, nil
}

func (r *PGColumn) Get(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
	const query = `
		SELECT id, board_id, name, description, position, created_at, updated_at
		FROM columns
		WHERE id = $1`

	var column domain.Column
	err := r.pgPool.QueryRow(ctx, query, columnID).Scan(
		&column.ID,
		&column.BoardID,
		&column.Name,
		&column.Description,
		&column.Position,
		&column.CreatedAt,
		&column.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Column{}, ErrRowNotFound
		}
		return domain.Column{}, fmt.Errorf("column repo: get by id: %v: %w", err, ErrInternal)
	}

	return column, nil
}

func (r *PGColumn) Update(
	ctx context.Context,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	name *domain.ColumnName,
	description *domain.ColumnDescription,
) (domain.Column, error) {
	const query = `
		UPDATE columns
		SET
			name = COALESCE($1, name),
			description = COALESCE($2, description),
			updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
		WHERE board_id = $3
		  AND id = $4
		RETURNING id, board_id, name, description, position, created_at, updated_at`

	var column domain.Column
	err := r.pgPool.QueryRow(ctx, query, name, description, boardID, columnID).Scan(
		&column.ID,
		&column.BoardID,
		&column.Name,
		&column.Description,
		&column.Position,
		&column.CreatedAt,
		&column.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Column{}, ErrRowNotFound
		}
		return domain.Column{}, fmt.Errorf("column repo: update by id: %v: %w", err, ErrInternal)
	}

	return column, nil
}

func (r *PGColumn) Move(
	ctx context.Context,
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
		"board_id": boardID,
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
		"board_id":  boardID,
		"column_id": columnID,
	}).Scan(&currentPosition)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ColumnPosition{}, ErrRowNotFound
		}
		return domain.ColumnPosition{}, fmt.Errorf("column repo: move get current position: %v: %w", err, ErrInternal)
	}

	var columnsCount int64
	err = tx.QueryRow(ctx, countColumnsQuery, pgx.NamedArgs{
		"board_id": boardID,
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

	err = tx.Commit(ctx)
	if err != nil {
		return domain.ColumnPosition{}, fmt.Errorf("column repo: move commit: %v: %w", err, ErrInternal)
	}

	return targetPosition, nil
}

func (r *PGColumn) Delete(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error {
	const (
		// 1. Lock the board row so no concurrent operation can reorder columns in the same board.
		lockBoardQuery = `
		SELECT 1
		FROM boards
		WHERE id = @board_id
		FOR UPDATE`

		// 2. Defer the unique constraint until COMMIT for this transaction only.
		deferPositionConstraintQuery = `
		SET CONSTRAINTS columns_board_id_position_key DEFERRED`

		// 3. Delete the target column and remember its position.
		deleteColumnQuery = `
		DELETE FROM columns
		WHERE board_id = @board_id
		  AND id = @column_id
		RETURNING position`

		// 4. Close the gap left by the deleted column.
		compactTrailingColumnsQuery = `
		UPDATE columns
		SET position = position - 1
		WHERE board_id = @board_id
		  AND position > @deleted_position`
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
		"board_id": boardID,
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

	var deletedPosition int64
	err = tx.QueryRow(ctx, deleteColumnQuery, pgx.NamedArgs{
		"board_id":  boardID,
		"column_id": columnID,
	}).Scan(&deletedPosition)
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

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("column repo: delete commit: %v: %w", err, ErrInternal)
	}

	return nil
}
