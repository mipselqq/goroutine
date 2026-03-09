package service

import (
	"context"
	"fmt"

	"goroutine/internal/domain"
)

type BoardRepository interface {
	Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) error
}

type Board struct {
	repository BoardRepository
}

func NewBoard(r BoardRepository) *Board {
	return &Board{repository: r}
}

func (s *Board) Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) error {
	err := s.repository.Create(ctx, ownerID, name, description)
	if err != nil {
		return fmt.Errorf("board service: create: %v: %w", err, ErrInternal)
	}

	return nil
}
