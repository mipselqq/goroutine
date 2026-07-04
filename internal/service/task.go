package service

import (
	"context"
	"errors"
	"fmt"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
)

type TaskRepository interface {
	Create(ctx context.Context, columnID domain.ColumnID, name domain.TaskName, description domain.TaskDescription) (domain.Task, error)
	ListByColumnID(ctx context.Context, columnID domain.ColumnID) ([]domain.Task, error)
	Get(ctx context.Context, taskID domain.TaskID) (domain.Task, error)
	Update(ctx context.Context, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error)
	Move(ctx context.Context, boardID domain.BoardID, currentColumnID domain.ColumnID, taskID domain.TaskID, targetColumnID domain.ColumnID, targetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error)
	Delete(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID) error
}

type TaskBoardRepository interface {
	Get(ctx context.Context, boardID domain.BoardID) (domain.Board, error)
}

type TaskColumnRepository interface {
	Get(ctx context.Context, columnID domain.ColumnID) (domain.Column, error)
}

type Task struct {
	taskRepo   TaskRepository
	boardRepo  TaskBoardRepository
	columnRepo TaskColumnRepository
}

func NewTask(taskRepo TaskRepository, boardRepo TaskBoardRepository, columnRepo TaskColumnRepository) *Task {
	return &Task{
		taskRepo:   taskRepo,
		boardRepo:  boardRepo,
		columnRepo: columnRepo,
	}
}

func (s *Task) Create(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	name domain.TaskName,
	description domain.TaskDescription,
) (domain.Task, error) {
	board, err := s.boardRepo.Get(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Task{}, ErrColumnNotFound
		}
		return domain.Task{}, fmt.Errorf("task service: create get board: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return domain.Task{}, ErrColumnNotFound
	}

	column, err := s.columnRepo.Get(ctx, columnID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Task{}, ErrColumnNotFound
		}
		return domain.Task{}, fmt.Errorf("task service: create get column: %v: %w", err, ErrInternal)
	}
	if column.BoardID != boardID {
		return domain.Task{}, ErrColumnNotFound
	}

	task, err := s.taskRepo.Create(ctx, columnID, name, description)
	if err != nil {
		return domain.Task{}, fmt.Errorf("task service: create: %v: %w", err, ErrInternal)
	}

	return task, nil
}

func (s *Task) ListByColumnID(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	columnID domain.ColumnID,
) ([]domain.Task, error) {
	board, err := s.boardRepo.Get(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return nil, ErrColumnNotFound
		}
		return nil, fmt.Errorf("task service: list get board: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return nil, ErrColumnNotFound
	}

	column, err := s.columnRepo.Get(ctx, columnID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return nil, ErrColumnNotFound
		}
		return nil, fmt.Errorf("task service: list get column: %v: %w", err, ErrInternal)
	}
	if column.BoardID != boardID {
		return nil, ErrColumnNotFound
	}

	tasks, err := s.taskRepo.ListByColumnID(ctx, columnID)
	if err != nil {
		return nil, fmt.Errorf("task service: list: %v: %w", err, ErrInternal)
	}

	return tasks, nil
}

func (s *Task) Update(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	taskID domain.TaskID,
	name *domain.TaskName,
	description *domain.TaskDescription,
) (domain.Task, error) {
	board, err := s.boardRepo.Get(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Task{}, ErrTaskNotFound
		}
		return domain.Task{}, fmt.Errorf("task service: update get board: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return domain.Task{}, ErrTaskNotFound
	}

	column, err := s.columnRepo.Get(ctx, columnID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Task{}, ErrTaskNotFound
		}
		return domain.Task{}, fmt.Errorf("task service: update get column: %v: %w", err, ErrInternal)
	}
	if column.BoardID != boardID {
		return domain.Task{}, ErrTaskNotFound
	}

	task, err := s.taskRepo.Get(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Task{}, ErrTaskNotFound
		}
		return domain.Task{}, fmt.Errorf("task service: update get task: %v: %w", err, ErrInternal)
	}
	if task.ColumnID != columnID {
		return domain.Task{}, ErrTaskNotFound
	}

	if name == nil && description == nil {
		return task, nil
	}

	updated, err := s.taskRepo.Update(ctx, columnID, taskID, name, description)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Task{}, ErrTaskNotFound
		}
		return domain.Task{}, fmt.Errorf("task service: update by id: %v: %w", err, ErrInternal)
	}

	return updated, nil
}

func (s *Task) Delete(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	taskID domain.TaskID,
) error {
	board, err := s.boardRepo.Get(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("task service: delete get board: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return ErrTaskNotFound
	}

	column, err := s.columnRepo.Get(ctx, columnID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("task service: delete get column: %v: %w", err, ErrInternal)
	}
	if column.BoardID != boardID {
		return ErrTaskNotFound
	}

	task, err := s.taskRepo.Get(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("task service: delete get task: %v: %w", err, ErrInternal)
	}
	if task.ColumnID != columnID {
		return ErrTaskNotFound
	}

	err = s.taskRepo.Delete(ctx, boardID, columnID, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("task service: delete: %v: %w", err, ErrInternal)
	}

	return nil
}

func (s *Task) Move(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	taskID domain.TaskID,
	targetColumnID domain.ColumnID,
	targetPosition domain.TaskPosition,
) (domain.ColumnID, domain.TaskPosition, error) {
	board, err := s.boardRepo.Get(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.ColumnID{}, domain.TaskPosition{}, ErrTaskNotFound
		}
		return domain.ColumnID{}, domain.TaskPosition{}, fmt.Errorf("task service: move get board: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return domain.ColumnID{}, domain.TaskPosition{}, ErrTaskNotFound
	}

	column, err := s.columnRepo.Get(ctx, columnID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.ColumnID{}, domain.TaskPosition{}, ErrTaskNotFound
		}
		return domain.ColumnID{}, domain.TaskPosition{}, fmt.Errorf("task service: move get column: %v: %w", err, ErrInternal)
	}
	if column.BoardID != boardID {
		return domain.ColumnID{}, domain.TaskPosition{}, ErrTaskNotFound
	}

	task, err := s.taskRepo.Get(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.ColumnID{}, domain.TaskPosition{}, ErrTaskNotFound
		}
		return domain.ColumnID{}, domain.TaskPosition{}, fmt.Errorf("task service: move get task: %v: %w", err, ErrInternal)
	}
	if task.ColumnID != columnID {
		return domain.ColumnID{}, domain.TaskPosition{}, ErrTaskNotFound
	}

	if targetColumnID != columnID {
		var targetColumn domain.Column
		targetColumn, err = s.columnRepo.Get(ctx, targetColumnID)
		if err != nil {
			if errors.Is(err, repository.ErrRowNotFound) {
				return domain.ColumnID{}, domain.TaskPosition{}, ErrColumnNotFound
			}
			return domain.ColumnID{}, domain.TaskPosition{}, fmt.Errorf("task service: move get target column: %v: %w", err, ErrInternal)
		}
		if targetColumn.BoardID != boardID {
			return domain.ColumnID{}, domain.TaskPosition{}, ErrColumnNotFound
		}
	}

	newColumnID, newPosition, err := s.taskRepo.Move(ctx, boardID, columnID, taskID, targetColumnID, targetPosition)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.ColumnID{}, domain.TaskPosition{}, ErrTaskNotFound
		}
		if errors.Is(err, repository.ErrIndexOutOfBounds) {
			return domain.ColumnID{}, domain.TaskPosition{}, ErrIndexOutOfBounds
		}
		return domain.ColumnID{}, domain.TaskPosition{}, fmt.Errorf("task service: move: %v: %w", err, ErrInternal)
	}

	return newColumnID, newPosition, nil
}
