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
	Delete(ctx context.Context, boardID domain.BoardID) error
}

type Board struct {
	repository BoardRepository
}

func NewBoard(r BoardRepository) *Board {
	return &Board{repository: r}
}

func (s *Board) Create(ctx context.Context, callerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
	board, err := s.repository.Create(ctx, callerID, name, description)
	if err != nil {
		return domain.Board{}, fmt.Errorf("board service: create: %v: %w", err, ErrInternal)
	}

	return board, nil
}

func (s *Board) GetMany(ctx context.Context, callerID domain.UserID) ([]domain.Board, error) {
	boards, err := s.repository.GetMany(ctx, callerID)
	if err != nil {
		return nil, fmt.Errorf("board service: get many: %v: %w", err, ErrInternal)
	}

	return boards, nil
}

func (s *Board) Get(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) (domain.Board, error) {
	board, err := s.repository.GetByID(ctx, boardID)
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

func (s *Board) Delete(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) error {
	board, err := s.repository.GetByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return ErrBoardNotFound
		}
		return fmt.Errorf("board service: delete get by id: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return ErrBoardNotFound
	}

	err = s.repository.Delete(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return ErrBoardNotFound
		}
		return fmt.Errorf("board service: delete: %v: %w", err, ErrInternal)
	}

	return nil
}
