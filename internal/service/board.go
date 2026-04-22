package service

import (
	"context"
	"errors"
	"fmt"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
)

type BoardRepository interface {
	Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error)
	GetByID(ctx context.Context, id domain.BoardID) (domain.Board, error)
	GetMany(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error)
	UpdateByID(ctx context.Context, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error)
	Delete(ctx context.Context, boardID domain.BoardID) error
}

type BoardColumnRepository interface {
	ListByBoardID(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error)
}

type BoardTaskRepository interface {
	ListByBoardID(ctx context.Context, boardID domain.BoardID) ([]domain.Task, error)
}

type AggregateBoard struct {
	Board   domain.Board
	Columns []AggregateColumn
}

type AggregateColumn struct {
	Column domain.Column
	Tasks  []domain.Task
}

type Board struct {
	boardRepository  BoardRepository
	columnRepository ColumnRepository
	taskRepository   BoardTaskRepository
}

func NewBoard(boardRepo BoardRepository, columnRepo ColumnRepository, taskRepo BoardTaskRepository) *Board {
	return &Board{boardRepository: boardRepo, columnRepository: columnRepo, taskRepository: taskRepo}
}

func (s *Board) Create(ctx context.Context, callerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
	board, err := s.boardRepository.Create(ctx, callerID, name, description)
	if err != nil {
		return domain.Board{}, fmt.Errorf("board service: create: %v: %w", err, ErrInternal)
	}

	return board, nil
}

func (s *Board) GetMany(ctx context.Context, callerID domain.UserID) ([]domain.Board, error) {
	boards, err := s.boardRepository.GetMany(ctx, callerID)
	if err != nil {
		return nil, fmt.Errorf("board service: get many: %v: %w", err, ErrInternal)
	}

	return boards, nil
}

func (s *Board) Get(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) (domain.Board, error) {
	board, err := s.boardRepository.GetByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Board{}, ErrBoardNotFound
		}
		return domain.Board{}, fmt.Errorf("board service: get: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return domain.Board{}, ErrBoardNotFound
	}

	return board, nil
}

func (s *Board) GetAggregate(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) (AggregateBoard, error) {
	panic("Get aggregate not implemented")
}

func (s *Board) UpdateByID(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	name *domain.BoardName,
	description *domain.BoardDescription,
) (domain.Board, error) {
	board, err := s.boardRepository.GetByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Board{}, ErrBoardNotFound
		}
		return domain.Board{}, fmt.Errorf("board service: update get by id: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return domain.Board{}, ErrBoardNotFound
	}

	if name == nil && description == nil {
		return board, nil
	}

	updated, err := s.boardRepository.UpdateByID(ctx, boardID, name, description)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Board{}, ErrBoardNotFound
		}
		return domain.Board{}, fmt.Errorf("board service: update: %v: %w", err, ErrInternal)
	}

	return updated, nil
}

func (s *Board) Delete(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) error {
	board, err := s.boardRepository.GetByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return ErrBoardNotFound
		}
		return fmt.Errorf("board service: delete get by id: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return ErrBoardNotFound
	}

	err = s.boardRepository.Delete(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return ErrBoardNotFound
		}
		return fmt.Errorf("board service: delete: %v: %w", err, ErrInternal)
	}

	return nil
}
