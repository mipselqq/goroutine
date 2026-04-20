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

func (r *PgColumn) ListByBoardID(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error) {
	const query = `
		SELECT id, board_id, name, position, created_at, updated_at
		FROM columns
		WHERE board_id = $1
		ORDER BY position ASC`

	rows, err := r.pool.Query(ctx, query, boardID)
	if err != nil {
		return nil, fmt.Errorf("column repo: list: %v: %w", err, ErrInternal)
	}
	defer rows.Close()

	var result []domain.Column
	for rows.Next() {
		var col domain.Column
		err := rows.Scan(
			&col.ID,
			&col.BoardID,
			&col.Name,
			&col.Position,
			&col.CreatedAt,
			&col.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("column repo: list: scan: %v: %w", err, ErrInternal)
		}
		result = append(result, col)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("column repo: list: rows final error: %v: %w", err, ErrInternal)
	}

	return result, nil
}

func (r *PgColumn) GetByID(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
	const query = `
		SELECT id, board_id, name, position, created_at, updated_at
		FROM columns
		WHERE id = $1`

	var column domain.Column
	err := r.pool.QueryRow(ctx, query, columnID).Scan(
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
		return domain.Column{}, fmt.Errorf("column repo: get by id: %v: %w", err, ErrInternal)
	}

	return column, nil
}

func (r *PgColumn) UpdateByID(
	ctx context.Context,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	name *domain.ColumnName,
	updatedAt time.Time,
) (domain.Column, error) {
	const query = `
		UPDATE columns
		SET
			name = COALESCE($3, name),
			updated_at = $4
		WHERE board_id = $1
		  AND id = $2
		RETURNING id, board_id, name, position, created_at, updated_at`

	var column domain.Column
	err := r.pool.QueryRow(ctx, query, boardID, columnID, name, updatedAt).Scan(
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
		return domain.Column{}, fmt.Errorf("column repo: update by id: %v: %w", err, ErrInternal)
	}

	return column, nil
}

func (r *PgColumn) Move(
	ctx context.Context,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	targetPosition domain.ColumnPosition,
) (domain.ColumnPosition, error) {
	const (
		// We cannot move columns in place with a simple "position = position +/- 1":
		// UNIQUE(board_id, position) may be violated temporarily, because PostgreSQL does not guarantee
		// the order in which rows are updated inside one UPDATE statement.
		//
		// So the move works in explicit steps:
		// 1. Lock the board row, so concurrent create/delete/move operations cannot reorder the same board.
		// 2. Read the current position of the moved column.
		// 3. Read the current column count and max position to validate targetPosition and compute delta.
		// 4. Move the whole affected interval out of the working range by +delta.
		//    After that, every working position inside [min(current, target), max(current, target)] is free.
		// 5. Move only the neighboring columns back into the working range, already shifted by one slot.
		// 6. Put the moved column into the target position, which is still free at that point.

		// 1. Lock the board row so no concurrent operation can reorder columns in the same board.
		lockBoardQuery = `
		SELECT 1
		FROM boards
		WHERE id = @board_id
		FOR UPDATE`

		// 2. Read the current position of the column we are moving.
		getCurrentPositionQuery = `
		SELECT position
		FROM columns
		WHERE board_id = @board_id
		  AND id = @column_id`

		// 3. Read how many columns the board currently has and what the current max position is.
		//    count(columns) is used for bounds validation, max(position) is used to compute delta.
		countAndMaxPositionQuery = `
		SELECT COUNT(*), COALESCE(MAX(position), 0)
		FROM columns
		WHERE board_id = @board_id`

		// 4. Move the whole affected interval out of the working range.
		//    Example: if positions 2..4 are affected and delta is 5, they become 7..9.
		//    This frees positions 2..4, so the final slots can be filled safely afterwards.
		bumpAffectedRangeQuery = `
		UPDATE columns
		SET position = position + @position_offset
		WHERE board_id = @board_id
		  AND position >= @affected_left
		  AND position <= @affected_right`

		// 5. If the moved column goes down, return neighbors from (current, target]
		//    one slot up: 3,4,5 -> 2,3,4. The target slot stays free for the moved column.
		moveNeighborsDownQuery = `
		UPDATE columns
		SET position = position - @position_offset - 1
		WHERE board_id = @board_id
		  AND position > @current_position_offset
		  AND position <= @target_position_offset`

		// 5. If the moved column goes up, return neighbors from [target, current)
		//    one slot down: 2,3,4 -> 3,4,5. The target slot stays free for the moved column.
		moveNeighborsUpQuery = `
		UPDATE columns
		SET position = position - @position_offset + 1
		WHERE board_id = @board_id
		  AND position >= @target_position_offset
		  AND position < @current_position_offset`

		// 6. Put the moved column into the target position after neighbors have been restored.
		moveColumnIntoTargetQuery = `
		UPDATE columns
		SET position = @target_position
		WHERE board_id = @board_id
		  AND id = @column_id`
	)

	tx, err := r.pool.Begin(ctx)
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
	var maxPosition int64
	err = tx.QueryRow(ctx, countAndMaxPositionQuery, pgx.NamedArgs{
		"board_id": boardID,
	}).Scan(&columnsCount, &maxPosition)
	if err != nil {
		return domain.ColumnPosition{}, fmt.Errorf("column repo: move count and max position: %v: %w", err, ErrInternal)
	}

	targetPositionInt := targetPosition.Int64()
	if targetPositionInt > columnsCount {
		return domain.ColumnPosition{}, ErrIndexOutOfBounds
	}
	if targetPositionInt == currentPosition {
		return targetPosition, nil
	}

	positionOffset := maxPosition + 1
	affectedLeft := min(currentPosition, targetPositionInt)
	affectedRight := max(currentPosition, targetPositionInt)

	_, err = tx.Exec(ctx, bumpAffectedRangeQuery, pgx.NamedArgs{
		"board_id":        boardID,
		"position_offset": positionOffset,
		"affected_left":   affectedLeft,
		"affected_right":  affectedRight,
	})
	if err != nil {
		return domain.ColumnPosition{}, fmt.Errorf("column repo: move bump affected range: %v: %w", err, ErrInternal)
	}

	currentPositionOffset := currentPosition + positionOffset
	targetPositionOffset := targetPositionInt + positionOffset
	moveNeighborsArgs := pgx.NamedArgs{
		"board_id":                boardID,
		"position_offset":         positionOffset,
		"current_position_offset": currentPositionOffset,
		"target_position_offset":  targetPositionOffset,
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

func (r *PgColumn) Delete(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error {
	const (
		lockBoardQuery = `
		SELECT 1
		FROM boards
		WHERE id = @board_id
		FOR UPDATE`

		deleteColumnQuery = `
		DELETE FROM columns
		WHERE board_id = @board_id
		  AND id = @column_id
		RETURNING position`

		// Workaround:
		// UNIQUE(board_id, position) may be violated during simple offsetting since crawling order is not guaranteed
		// So we shift all the columns out of working scope and then move them back in the correct order
		positionOffsetQuery = `
		SELECT COALESCE(MAX(position), 0) + 1
		FROM columns
		WHERE board_id = @board_id`

		bumpTrailingColumnsQuery = `
		UPDATE columns
		SET position = position + @position_offset
		WHERE board_id = @board_id
		  AND position > @deleted_position`

		compactTrailingColumnsQuery = `
		UPDATE columns
		SET position = position - @position_offset - 1
		WHERE board_id = @board_id
		  AND position > @position_offset`
	)

	tx, err := r.pool.Begin(ctx)
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

	var positionOffset int64
	err = tx.QueryRow(ctx, positionOffsetQuery, pgx.NamedArgs{
		"board_id": boardID,
	}).Scan(&positionOffset)
	if err != nil {
		return fmt.Errorf("column repo: delete position offset: %v: %w", err, ErrInternal)
	}

	_, err = tx.Exec(ctx, bumpTrailingColumnsQuery, pgx.NamedArgs{
		"board_id":         boardID,
		"position_offset":  positionOffset,
		"deleted_position": deletedPosition,
	})
	if err != nil {
		return fmt.Errorf("column repo: delete bump trailing columns: %v: %w", err, ErrInternal)
	}

	_, err = tx.Exec(ctx, compactTrailingColumnsQuery, pgx.NamedArgs{
		"board_id":        boardID,
		"position_offset": positionOffset,
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
