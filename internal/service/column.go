package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
)

type ColumnRepository interface {
	Create(
		ctx context.Context,
		boardID domain.BoardID,
		name domain.ColumnName,
		createdAt time.Time,
		updatedAt time.Time,
	) (domain.Column, error)
	ListByBoardID(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error)
}

type ColumnBoardRepository interface {
	GetByID(ctx context.Context, id domain.BoardID) (domain.Board, error)
}

type Column struct {
	columnRepository ColumnRepository
	boardRepository  ColumnBoardRepository
	timeFunc         func() time.Time
}

func NewColumn(columnRepo ColumnRepository, boardRepo ColumnBoardRepository, timeFunc TimeFunc) *Column {
	if timeFunc == nil {
		panic("BUG: timeFunc is nil")
	}

	return &Column{columnRepository: columnRepo, boardRepository: boardRepo, timeFunc: timeFunc}
}

func (s *Column) Create(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName) (domain.Column, error) {
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

	now := s.timeFunc()
	column, err := s.columnRepository.Create(ctx, boardID, name, now, now)
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
