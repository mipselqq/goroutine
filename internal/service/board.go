package service

import (
	"context"
	"fmt"

	"goroutine/internal/domain"
)

type BoardRepository interface {
	Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error)
	GetMany(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error)
}

type Board struct {
	repository BoardRepository
}

func NewBoard(r BoardRepository) *Board {
	return &Board{repository: r}
}

func (s *Board) Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
	board, err := s.repository.Create(ctx, ownerID, name, description)
	if err != nil {
		return domain.Board{}, fmt.Errorf("board service: create: %v: %w", err, ErrInternal)
	}

	return board, nil
}

func (s *Board) GetMany(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
	boards, err := s.repository.GetMany(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("board service: get many: %v: %w", err, ErrInternal)
	}

	return boards, nil
}
