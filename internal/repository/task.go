package repository

import (
	"context"
	"errors"
	"fmt"

	"goroutine/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgTask struct {
	pool *pgxpool.Pool
}

func NewPgTask(pool *pgxpool.Pool) *PgTask {
	return &PgTask{pool: pool}
}

// LockTaskColumns acquires FOR UPDATE row locks on the given columns for boardID.
func LockTaskColumns(
	ctx context.Context,
	tx pgx.Tx,
	boardID domain.BoardID, // To ensure columns are locked in the same board.
	columnIDs ...domain.ColumnID,
) error {
	if len(columnIDs) == 0 {
		return errors.New("BUG: LockTaskColumns called with no columns. Isn't column ID forgotten?")
	}

	seen := make(map[domain.ColumnID]struct{}, len(columnIDs))
	for _, columnID := range columnIDs {
		if _, ok := seen[columnID]; ok {
			return errors.New("BUG: LockTaskColumns called so it locks the same column multiple times")
		}
		seen[columnID] = struct{}{}
	}

	// Deadlock protection in case same ids are passed in a different order:
	// if T1 locks A and then waits for B, while T2 already locked B and waits for A,
	// PostgreSQL will detect a deadlock and abort one transaction. Ordering makes all
	// callers acquire row locks in the same order.
	const lockColumnsQuery = `
		SELECT id
		FROM columns
		WHERE board_id = @board_id
		  AND id = ANY(@column_ids)
		ORDER BY id
		FOR UPDATE`

	rows, err := tx.Query(ctx, lockColumnsQuery, pgx.NamedArgs{
		"board_id":   boardID,
		"column_ids": columnIDs,
	})
	if err != nil {
		return fmt.Errorf("lock task columns: %w", err)
	}
	defer rows.Close()

	locked := 0
	for rows.Next() {
		var columnID domain.ColumnID
		if err = rows.Scan(&columnID); err != nil {
			return fmt.Errorf("failed to scan locked column row: %w", err)
		}
		locked++
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf("lock task columns rows final error: %w", err)
	}
	if locked != len(columnIDs) {
		return ErrRowNotFound
	}

	return nil
}

func (r *PgTask) Create(
	ctx context.Context,
	columnID domain.ColumnID,
	name domain.TaskName,
	description domain.TaskDescription,
) (domain.Task, error) {
	const query = `
		WITH column_lock AS (
			SELECT id
			FROM columns
			WHERE id = $1
			FOR UPDATE
		), next_pos AS (
			SELECT COALESCE(MAX(position), 0) + 1 AS position
			FROM tasks
			WHERE column_id = $1
		), inserted AS (
			INSERT INTO tasks (column_id, name, description, position)
			SELECT $1, $2, $3, next_pos.position
			FROM column_lock, next_pos
			RETURNING id, column_id, name, description, position, created_at, updated_at
		)
		SELECT id, column_id, name, description, position, created_at, updated_at
		FROM inserted`

	var task domain.Task
	err := r.pool.QueryRow(ctx, query, columnID, name, description).Scan(
		&task.ID,
		&task.ColumnID,
		&task.Name,
		&task.Description,
		&task.Position,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Task{}, ErrRowNotFound
		}
		return domain.Task{}, fmt.Errorf("task repo: create: %v: %w", err, ErrInternal)
	}

	return task, nil
}

func (r *PgTask) ListByBoardID(ctx context.Context, boardID domain.BoardID) ([]domain.Task, error) {
	const query = `
	SELECT t.id, t.column_id, t.name, t.description, t.position, t.created_at, t.updated_at
	FROM tasks t JOIN columns c ON t.column_id = c.id
	WHERE c.board_id = $1
	ORDER BY c.position ASC, t.position ASC
	`

	rows, err := r.pool.Query(ctx, query, boardID)
	if err != nil {
		return nil, fmt.Errorf("task repo: list by board id: %v: %w", err, ErrInternal)
	}
	defer rows.Close()

	var result []domain.Task
	for rows.Next() {
		var task domain.Task
		err := rows.Scan(
			&task.ID,
			&task.ColumnID,
			&task.Name,
			&task.Description,
			&task.Position,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("task repo: list by board id: scan: %v: %w", err, ErrInternal)
		}
		result = append(result, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("task repo: list by board id: rows final error: %v: %w", err, ErrInternal)
	}

	return result, nil
}

func (r *PgTask) ListByColumnID(ctx context.Context, columnID domain.ColumnID) ([]domain.Task, error) {
	const query = `
		SELECT id, column_id, name, description, position, created_at, updated_at
		FROM tasks
		WHERE column_id = $1
		ORDER BY position ASC`

	rows, err := r.pool.Query(ctx, query, columnID)
	if err != nil {
		return nil, fmt.Errorf("task repo: list by column id: %v: %w", err, ErrInternal)
	}
	defer rows.Close()

	var result []domain.Task
	for rows.Next() {
		var task domain.Task
		err := rows.Scan(
			&task.ID,
			&task.ColumnID,
			&task.Name,
			&task.Description,
			&task.Position,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("task repo: list by column id: scan: %v: %w", err, ErrInternal)
		}
		result = append(result, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("task repo: list: rows final error: %v: %w", err, ErrInternal)
	}

	return result, nil
}

func (r *PgTask) GetByID(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
	const query = `
		SELECT id, column_id, name, description, position, created_at, updated_at
		FROM tasks
		WHERE id = $1`

	var task domain.Task
	err := r.pool.QueryRow(ctx, query, taskID).Scan(
		&task.ID,
		&task.ColumnID,
		&task.Name,
		&task.Description,
		&task.Position,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Task{}, ErrRowNotFound
		}
		return domain.Task{}, fmt.Errorf("task repo: get by id: %v: %w", err, ErrInternal)
	}

	return task, nil
}

func (r *PgTask) UpdateByID(
	ctx context.Context,
	columnID domain.ColumnID,
	taskID domain.TaskID,
	name *domain.TaskName,
	description *domain.TaskDescription,
) (domain.Task, error) {
	const query = `
		UPDATE tasks
		SET
			name = COALESCE($3, name),
			description = COALESCE($4, description),
			updated_at = CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
		WHERE column_id = $1
		  AND id = $2
		RETURNING id, column_id, name, description, position, created_at, updated_at`

	var task domain.Task
	err := r.pool.QueryRow(ctx, query, columnID, taskID, name, description).Scan(
		&task.ID,
		&task.ColumnID,
		&task.Name,
		&task.Description,
		&task.Position,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Task{}, ErrRowNotFound
		}
		return domain.Task{}, fmt.Errorf("task repo: update by id: %v: %w", err, ErrInternal)
	}

	return task, nil
}

func (r *PgTask) Move(
	ctx context.Context,
	boardID domain.BoardID,
	currentColumnID domain.ColumnID,
	taskID domain.TaskID,
	targetColumnID domain.ColumnID,
	targetPosition domain.TaskPosition,
) (domain.ColumnID, domain.TaskPosition, error) {
	const (
		// 2. SET position order is not guaranteed, so we disable uniqueness constraint for this transaction.
		deferPositionConstraintQuery = `
		SET CONSTRAINTS tasks_column_id_position_key DEFERRED`

		// 3. Read the current position of the task we are moving in its source column.
		getCurrentPositionQuery = `
		SELECT position
		FROM tasks
		WHERE column_id = @current_column_id
		  AND id = @task_id`

		// 4. Read how many tasks the target column currently has to validate targetPosition.
		countTargetTasksQuery = `
		SELECT COUNT(*)
		FROM tasks
		WHERE column_id = @target_column_id`

		// 5a. Same column, moving down: shift neighbors from (current, target] one slot up.
		//     Example: moving 2 -> 5 means 3,4,5 become 2,3,4.
		moveNeighborsDownQuery = `
		UPDATE tasks
		SET position = position - 1
		WHERE column_id = @current_column_id
		  AND position > @current_position
		  AND position <= @target_position`

		// 5b. Same column, moving up: shift neighbors from [target, current) one slot down.
		//     Example: moving 5 -> 2 means 2,3,4 become 3,4,5.
		moveNeighborsUpQuery = `
		UPDATE tasks
		SET position = position + 1
		WHERE column_id = @current_column_id
		  AND position >= @target_position
		  AND position < @current_position`

		// 5c. Cross-column compaction of the source column: shift everything below the
		//     moved task one slot up to close the gap.
		compactSourceQuery = `
		UPDATE tasks
		SET position = position - 1
		WHERE column_id = @current_column_id
		  AND position > @current_position`

		// 5d. Cross-column slot opening in the target column: shift positions >= target
		//     one slot down to make room.
		openTargetSlotQuery = `
		UPDATE tasks
		SET position = position + 1
		WHERE column_id = @target_column_id
		  AND position >= @target_position`

		// 6a. Same-column move: place the task at the target position.
		moveTaskWithinColumnQuery = `
		UPDATE tasks
		SET position = @target_position
		WHERE id = @task_id`

		// 6b. Cross-column move: switch column_id and place the task at the target position.
		moveTaskAcrossColumnsQuery = `
		UPDATE tasks
		SET column_id = @target_column_id,
		    position = @target_position
		WHERE id = @task_id`
	)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.ColumnID{}, domain.TaskPosition{}, fmt.Errorf("task repo: move begin tx: %v: %w", err, ErrInternal)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	sameColumn := currentColumnID == targetColumnID

	// 1. Lock affected columns so concurrent operations can't interrupt the move.
	if sameColumn {
		err = LockTaskColumns(ctx, tx, boardID, currentColumnID)
	} else {
		err = LockTaskColumns(ctx, tx, boardID, currentColumnID, targetColumnID)
	}
	if err != nil {
		if errors.Is(err, ErrRowNotFound) {
			return domain.ColumnID{}, domain.TaskPosition{}, ErrRowNotFound
		}
		return domain.ColumnID{}, domain.TaskPosition{}, fmt.Errorf("task repo: move lock columns: %v: %w", err, ErrInternal)
	}

	_, err = tx.Exec(ctx, deferPositionConstraintQuery)
	if err != nil {
		return domain.ColumnID{}, domain.TaskPosition{}, fmt.Errorf("task repo: move defer position constraint: %v: %w", err, ErrInternal)
	}

	var currentPosition int64
	err = tx.QueryRow(ctx, getCurrentPositionQuery, pgx.NamedArgs{
		"current_column_id": currentColumnID,
		"task_id":           taskID,
	}).Scan(&currentPosition)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ColumnID{}, domain.TaskPosition{}, ErrRowNotFound
		}
		return domain.ColumnID{}, domain.TaskPosition{}, fmt.Errorf("task repo: move get current position: %v: %w", err, ErrInternal)
	}

	targetPositionInt := targetPosition.Int64()

	var targetTasksCount int64
	err = tx.QueryRow(ctx, countTargetTasksQuery, pgx.NamedArgs{
		"target_column_id": targetColumnID,
	}).Scan(&targetTasksCount)
	if err != nil {
		return domain.ColumnID{}, domain.TaskPosition{}, fmt.Errorf("task repo: move count target tasks: %v: %w", err, ErrInternal)
	}

	if sameColumn {
		// In the same column, moving the task does not grow the column, so the
		// upper bound is the current task count.
		if targetPositionInt > targetTasksCount {
			return domain.ColumnID{}, domain.TaskPosition{}, ErrIndexOutOfBounds
		}
		if targetPositionInt == currentPosition {
			return currentColumnID, targetPosition, nil
		}

		moveNeighborsArgs := pgx.NamedArgs{
			"current_column_id": currentColumnID,
			"current_position":  currentPosition,
			"target_position":   targetPositionInt,
		}
		if currentPosition < targetPositionInt {
			_, err = tx.Exec(ctx, moveNeighborsDownQuery, moveNeighborsArgs)
			if err != nil {
				return domain.ColumnID{}, domain.TaskPosition{}, fmt.Errorf("task repo: move neighbors down: %v: %w", err, ErrInternal)
			}
		} else {
			_, err = tx.Exec(ctx, moveNeighborsUpQuery, moveNeighborsArgs)
			if err != nil {
				return domain.ColumnID{}, domain.TaskPosition{}, fmt.Errorf("task repo: move neighbors up: %v: %w", err, ErrInternal)
			}
		}

		_, err = tx.Exec(ctx, moveTaskWithinColumnQuery, pgx.NamedArgs{
			"task_id":         taskID,
			"target_position": targetPositionInt,
		})
		if err != nil {
			return domain.ColumnID{}, domain.TaskPosition{}, fmt.Errorf("task repo: move task within column: %v: %w", err, ErrInternal)
		}
	} else {
		// Across columns the target column grows by one, so an append at
		// targetTasksCount+1 is valid.
		if targetPositionInt > targetTasksCount+1 {
			return domain.ColumnID{}, domain.TaskPosition{}, ErrIndexOutOfBounds
		}

		_, err = tx.Exec(ctx, compactSourceQuery, pgx.NamedArgs{
			"current_column_id": currentColumnID,
			"current_position":  currentPosition,
		})
		if err != nil {
			return domain.ColumnID{}, domain.TaskPosition{}, fmt.Errorf("task repo: move compact source: %v: %w", err, ErrInternal)
		}

		_, err = tx.Exec(ctx, openTargetSlotQuery, pgx.NamedArgs{
			"target_column_id": targetColumnID,
			"target_position":  targetPositionInt,
		})
		if err != nil {
			return domain.ColumnID{}, domain.TaskPosition{}, fmt.Errorf("task repo: move open target slot: %v: %w", err, ErrInternal)
		}

		_, err = tx.Exec(ctx, moveTaskAcrossColumnsQuery, pgx.NamedArgs{
			"task_id":          taskID,
			"target_column_id": targetColumnID,
			"target_position":  targetPositionInt,
		})
		if err != nil {
			return domain.ColumnID{}, domain.TaskPosition{}, fmt.Errorf("task repo: move task across columns: %v: %w", err, ErrInternal)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return domain.ColumnID{}, domain.TaskPosition{}, fmt.Errorf("task repo: move commit: %v: %w", err, ErrInternal)
	}

	return targetColumnID, targetPosition, nil
}

func (r *PgTask) Delete(
	ctx context.Context,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	taskID domain.TaskID,
) error {
	const (
		// 2. Defer the unique constraint until COMMIT for this transaction only.
		deferPositionConstraintQuery = `
		SET CONSTRAINTS tasks_column_id_position_key DEFERRED`

		// 3. Delete the target task and remember its position.
		deleteTaskQuery = `
		DELETE FROM tasks
		WHERE column_id = @column_id
		  AND id = @task_id
		RETURNING position`

		// 4. Close the gap left by the deleted task.
		compactTrailingTasksQuery = `
		UPDATE tasks
		SET position = position - 1
		WHERE column_id = @column_id
		  AND position > @deleted_position`
	)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("task repo: delete begin tx: %v: %w", err, ErrInternal)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// 1. Lock affected columns so concurrent operations can't interrupt the delete.
	err = LockTaskColumns(ctx, tx, boardID, columnID)
	if err != nil {
		if errors.Is(err, ErrRowNotFound) {
			return ErrRowNotFound
		}
		return fmt.Errorf("task repo: delete lock column: %v: %w", err, ErrInternal)
	}

	_, err = tx.Exec(ctx, deferPositionConstraintQuery)
	if err != nil {
		return fmt.Errorf("task repo: delete defer position constraint: %v: %w", err, ErrInternal)
	}

	var deletedPosition int64
	err = tx.QueryRow(ctx, deleteTaskQuery, pgx.NamedArgs{
		"column_id": columnID,
		"task_id":   taskID,
	}).Scan(&deletedPosition)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRowNotFound
		}
		return fmt.Errorf("task repo: delete task: %v: %w", err, ErrInternal)
	}

	_, err = tx.Exec(ctx, compactTrailingTasksQuery, pgx.NamedArgs{
		"column_id":        columnID,
		"deleted_position": deletedPosition,
	})
	if err != nil {
		return fmt.Errorf("task repo: delete compact trailing tasks: %v: %w", err, ErrInternal)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("task repo: delete commit: %v: %w", err, ErrInternal)
	}

	return nil
}
