package service_test

import (
	"context"
	"fmt"
	"time"

	"goroutine/internal/domain"
)

func AssertFuncNotNil(funcName string, fn any) {
	if fn == nil {
		panic(fmt.Sprintf("BUG: Mock: %s is not set", funcName))
	}
}

type MockUserRepository struct {
	InsertFunc     func(ctx context.Context, email domain.Email, hash string) error
	GetByEmailFunc func(ctx context.Context, email domain.Email) (id domain.UserID, hash string, err error)
}

func (m *MockUserRepository) Insert(ctx context.Context, email domain.Email, hash string) error {
	AssertFuncNotNil("UserRepository.InsertFunc", m.InsertFunc)
	return m.InsertFunc(ctx, email, hash)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email domain.Email) (id domain.UserID, hash string, err error) {
	AssertFuncNotNil("UserRepository.GetByEmailFunc", m.GetByEmailFunc)
	return m.GetByEmailFunc(ctx, email)
}

type MockBoardRepository struct {
	CreateFunc     func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error)
	GetByIDFunc    func(ctx context.Context, id domain.BoardID) (domain.Board, error)
	GetManyFunc    func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error)
	UpdateByIDFunc func(ctx context.Context, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription, updatedAt time.Time) (domain.Board, error)
	DeleteFunc     func(ctx context.Context, boardID domain.BoardID) error
}

type MockColumnRepository struct {
	CreateFunc func(ctx context.Context, boardID domain.BoardID, name domain.ColumnName, createdAt time.Time, updatedAt time.Time) (domain.Column, error)
}

func (m *MockBoardRepository) Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
	AssertFuncNotNil("BoardRepository.CreateFunc", m.CreateFunc)
	return m.CreateFunc(ctx, ownerID, name, description)
}

func (m *MockBoardRepository) GetByID(ctx context.Context, id domain.BoardID) (domain.Board, error) {
	AssertFuncNotNil("BoardRepository.GetByIDFunc", m.GetByIDFunc)
	return m.GetByIDFunc(ctx, id)
}

func (m *MockBoardRepository) GetMany(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
	AssertFuncNotNil("BoardRepository.GetManyFunc", m.GetManyFunc)
	return m.GetManyFunc(ctx, ownerID)
}

func (m *MockBoardRepository) UpdateByID(ctx context.Context, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription, updatedAt time.Time) (domain.Board, error) {
	AssertFuncNotNil("BoardRepository.UpdateByIDFunc", m.UpdateByIDFunc)
	return m.UpdateByIDFunc(ctx, boardID, name, description, updatedAt)
}

func (m *MockBoardRepository) Delete(ctx context.Context, boardID domain.BoardID) error {
	AssertFuncNotNil("BoardRepository.DeleteFunc", m.DeleteFunc)
	return m.DeleteFunc(ctx, boardID)
}

func (m *MockColumnRepository) Create(
	ctx context.Context,
	boardID domain.BoardID,
	name domain.ColumnName,
	createdAt time.Time,
	updatedAt time.Time,
) (domain.Column, error) {
	AssertFuncNotNil("ColumnRepository.CreateFunc", m.CreateFunc)
	return m.CreateFunc(ctx, boardID, name, createdAt, updatedAt)
}
