package service

import (
	"context"
	"errors"
	"fmt"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
)

type columnRepository interface {
	Create(ctx context.Context, boardID domain.BoardID, name domain.ColumnName, description domain.ColumnDescription) (domain.Column, error)
	ListByBoardID(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error)
	Get(ctx context.Context, columnID domain.ColumnID) (domain.Column, error)
	Update(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error)
	Move(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error)
	Delete(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error
}

type columnBoardRepository interface {
	Get(ctx context.Context, boardID domain.BoardID) (domain.Board, error)
}

type column struct {
	columnRepo columnRepository
	boardRepo  columnBoardRepository
}

func NewColumn(columnRepo columnRepository, boardRepo columnBoardRepository) *column {
	return &column{columnRepo: columnRepo, boardRepo: boardRepo}
}

func (s *column) Create(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName, description domain.ColumnDescription) (domain.Column, error) {
	board, err := s.boardRepo.Get(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Column{}, ErrBoardNotFound
		}
		return domain.Column{}, fmt.Errorf("column service: create get board: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return domain.Column{}, ErrBoardNotFound
	}

	column, err := s.columnRepo.Create(ctx, boardID, name, description)
	if err != nil {
		return domain.Column{}, fmt.Errorf("column service: create: %v: %w", err, ErrInternal)
	}

	return column, nil
}

func (s *column) ListByBoardID(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) ([]domain.Column, error) {
	board, err := s.boardRepo.Get(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return nil, ErrBoardNotFound
		}
		return nil, fmt.Errorf("column service: list by board id get board: %v: %w", err, ErrInternal)
	}

	if board.OwnerID != callerID {
		return nil, ErrBoardNotFound
	}

	columns, err := s.columnRepo.ListByBoardID(ctx, boardID)
	if err != nil {
		return nil, fmt.Errorf("column service: list by board id: %v: %w", err, ErrInternal)
	}

	return columns, nil
}

func (s *column) Update(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	name *domain.ColumnName,
	description *domain.ColumnDescription,
) (domain.Column, error) {
	board, err := s.boardRepo.Get(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Column{}, ErrColumnNotFound
		}
		return domain.Column{}, fmt.Errorf("column service: update get board: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return domain.Column{}, ErrColumnNotFound
	}

	column, err := s.columnRepo.Get(ctx, columnID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Column{}, ErrColumnNotFound
		}
		return domain.Column{}, fmt.Errorf("column service: update get column: %v: %w", err, ErrInternal)
	}
	if column.BoardID != boardID {
		return domain.Column{}, ErrColumnNotFound
	}

	if name == nil && description == nil {
		return column, nil
	}

	updated, err := s.columnRepo.Update(ctx, boardID, columnID, name, description)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Column{}, ErrColumnNotFound
		}
		return domain.Column{}, fmt.Errorf("column service: update: %v: %w", err, ErrInternal)
	}

	return updated, nil
}

func (s *column) Delete(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	columnID domain.ColumnID,
) error {
	board, err := s.boardRepo.Get(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return ErrColumnNotFound
		}
		return fmt.Errorf("column service: delete get board: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return ErrColumnNotFound
	}

	err = s.columnRepo.Delete(ctx, boardID, columnID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return ErrColumnNotFound
		}
		return fmt.Errorf("column service: delete: %v: %w", err, ErrInternal)
	}

	return nil
}

func (s *column) Move(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	targetPosition domain.ColumnPosition,
) (domain.ColumnPosition, error) {
	board, err := s.boardRepo.Get(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.ColumnPosition{}, ErrColumnNotFound
		}
		return domain.ColumnPosition{}, fmt.Errorf("column service: move get board: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return domain.ColumnPosition{}, ErrColumnNotFound
	}

	column, err := s.columnRepo.Get(ctx, columnID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.ColumnPosition{}, ErrColumnNotFound
		}
		return domain.ColumnPosition{}, fmt.Errorf("column service: move get column: %v: %w", err, ErrInternal)
	}
	if column.BoardID != boardID {
		return domain.ColumnPosition{}, ErrColumnNotFound
	}

	position, err := s.columnRepo.Move(ctx, boardID, columnID, targetPosition)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.ColumnPosition{}, ErrColumnNotFound
		}
		if errors.Is(err, repository.ErrIndexOutOfBounds) {
			return domain.ColumnPosition{}, ErrIndexOutOfBounds
		}
		return domain.ColumnPosition{}, fmt.Errorf("column service: move: %v: %w", err, ErrInternal)
	}

	return position, nil
}
