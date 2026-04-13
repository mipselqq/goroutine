package service_test

import (
	"context"
	"errors"
	"time"

	"goroutine/internal/domain"
)

type MockUserRepository struct {
	InsertFunc     func(ctx context.Context, email domain.Email, hash string) error
	GetByEmailFunc func(ctx context.Context, email domain.Email) (id domain.UserID, hash string, err error)
}

func (m *MockUserRepository) Insert(ctx context.Context, email domain.Email, hash string) error {
	return m.InsertFunc(ctx, email, hash)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email domain.Email) (id domain.UserID, hash string, err error) {
	return m.GetByEmailFunc(ctx, email)
}

type MockBoardRepository struct {
	CreateFunc     func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error)
	GetByIDFunc    func(ctx context.Context, id domain.BoardID) (domain.Board, error)
	GetManyFunc    func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error)
	UpdateByIDFunc func(ctx context.Context, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription, updatedAt time.Time) (domain.Board, error)
	DeleteFunc     func(ctx context.Context, boardID domain.BoardID) error
}

func (m *MockBoardRepository) Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
	// TODO(refactor-1): create a function assertFuncNotNil
	return m.CreateFunc(ctx, ownerID, name, description)
}

func (m *MockBoardRepository) GetByID(ctx context.Context, id domain.BoardID) (domain.Board, error) {
	if m.GetByIDFunc == nil {
		return domain.Board{}, errors.New("BUG: GetByIDFunc is not set")
	}
	return m.GetByIDFunc(ctx, id)
}

func (m *MockBoardRepository) GetMany(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
	if m.GetManyFunc == nil {
		return nil, errors.New("BUG: GetManyFunc is not set")
	}
	return m.GetManyFunc(ctx, ownerID)
}

func (m *MockBoardRepository) UpdateByID(ctx context.Context, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription, updatedAt time.Time) (domain.Board, error) {
	if m.UpdateByIDFunc == nil {
		return domain.Board{}, errors.New("BUG: UpdateByIDFunc is not set")
	}
	return m.UpdateByIDFunc(ctx, boardID, name, description, updatedAt)
}

func (m *MockBoardRepository) Delete(ctx context.Context, boardID domain.BoardID) error {
	if m.DeleteFunc == nil {
		return errors.New("BUG: DeleteFunc is not set")
	}
	return m.DeleteFunc(ctx, boardID)
}
