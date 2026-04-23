package service

import (
	"context"
	"errors"
	"fmt"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
)

type ColumnRepository interface {
	Create(ctx context.Context, boardID domain.BoardID, name domain.ColumnName, description domain.ColumnDescription) (domain.Column, error)
	ListByBoardID(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error)
	GetByID(ctx context.Context, columnID domain.ColumnID) (domain.Column, error)
	UpdateByID(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error)
	Move(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error)
	Delete(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error
}

type ColumnBoardRepository interface {
	GetByID(ctx context.Context, id domain.BoardID) (domain.Board, error)
}

type Column struct {
	columnRepository ColumnRepository
	boardRepository  ColumnBoardRepository
}

func NewColumn(columnRepo ColumnRepository, boardRepo ColumnBoardRepository) *Column {
	return &Column{columnRepository: columnRepo, boardRepository: boardRepo}
}

func (s *Column) Create(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName, description domain.ColumnDescription) (domain.Column, error) {
	board, err := s.boardRepository.GetByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Column{}, ErrBoardNotFound
		}
		return domain.Column{}, fmt.Errorf("column service: create get board: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return domain.Column{}, ErrBoardNotFound
	}

	column, err := s.columnRepository.Create(ctx, boardID, name, description)
	if err != nil {
		return domain.Column{}, fmt.Errorf("column service: create: %v: %w", err, ErrInternal)
	}

	return column, nil
}

func (s *Column) List(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) ([]domain.Column, error) {
	board, err := s.boardRepository.GetByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return nil, ErrBoardNotFound
		}
		return nil, fmt.Errorf("column service: list get board: %v: %w", err, ErrInternal)
	}

	if board.OwnerID != callerID {
		return nil, ErrBoardNotFound
	}

	columns, err := s.columnRepository.ListByBoardID(ctx, boardID)
	if err != nil {
		return nil, fmt.Errorf("column service: list: %v: %w", err, ErrInternal)
	}

	return columns, nil
}

func (s *Column) UpdateByID(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	name *domain.ColumnName,
	description *domain.ColumnDescription,
) (domain.Column, error) {
	board, err := s.boardRepository.GetByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Column{}, ErrColumnNotFound
		}
		return domain.Column{}, fmt.Errorf("column service: update get board: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return domain.Column{}, ErrColumnNotFound
	}

	column, err := s.columnRepository.GetByID(ctx, columnID)
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

	updated, err := s.columnRepository.UpdateByID(ctx, boardID, columnID, name, description)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Column{}, ErrColumnNotFound
		}
		return domain.Column{}, fmt.Errorf("column service: update by id: %v: %w", err, ErrInternal)
	}

	return updated, nil
}

func (s *Column) Delete(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	columnID domain.ColumnID,
) error {
	board, err := s.boardRepository.GetByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return ErrColumnNotFound
		}
		return fmt.Errorf("column service: delete get board: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return ErrColumnNotFound
	}

	err = s.columnRepository.Delete(ctx, boardID, columnID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return ErrColumnNotFound
		}
		return fmt.Errorf("column service: delete: %v: %w", err, ErrInternal)
	}

	return nil
}

func (s *Column) Move(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	targetPosition domain.ColumnPosition,
) (domain.ColumnPosition, error) {
	board, err := s.boardRepository.GetByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.ColumnPosition{}, ErrColumnNotFound
		}
		return domain.ColumnPosition{}, fmt.Errorf("column service: move get board: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return domain.ColumnPosition{}, ErrColumnNotFound
	}

	column, err := s.columnRepository.GetByID(ctx, columnID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.ColumnPosition{}, ErrColumnNotFound
		}
		return domain.ColumnPosition{}, fmt.Errorf("column service: move get column: %v: %w", err, ErrInternal)
	}
	if column.BoardID != boardID {
		return domain.ColumnPosition{}, ErrColumnNotFound
	}

	position, err := s.columnRepository.Move(ctx, boardID, columnID, targetPosition)
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
