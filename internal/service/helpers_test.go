package service_test

import (
	"context"
	"fmt"
	"time"

	"goroutine/internal/domain"
)

func AssertFuncNotNil(funcName string, fn any) {
	if fn == nil {
		panic(fmt.Sprintf("%s = nil, want configured mock", funcName))
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
	CreateFunc        func(ctx context.Context, boardID domain.BoardID, name domain.ColumnName, createdAt time.Time, updatedAt time.Time) (domain.Column, error)
	ListByBoardIDFunc func(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error)
	GetByIDFunc       func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error)
	UpdateByIDFunc    func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, updatedAt time.Time) (domain.Column, error)
	MoveFunc          func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error)
	DeleteFunc        func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error
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

func (m *MockColumnRepository) ListByBoardID(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error) {
	AssertFuncNotNil("ColumnRepository.ListByBoardIDFunc", m.ListByBoardIDFunc)
	return m.ListByBoardIDFunc(ctx, boardID)
}

func (m *MockColumnRepository) GetByID(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
	AssertFuncNotNil("ColumnRepository.GetByIDFunc", m.GetByIDFunc)
	return m.GetByIDFunc(ctx, columnID)
}

func (m *MockColumnRepository) UpdateByID(
	ctx context.Context,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	name *domain.ColumnName,
	updatedAt time.Time,
) (domain.Column, error) {
	AssertFuncNotNil("ColumnRepository.UpdateByIDFunc", m.UpdateByIDFunc)
	return m.UpdateByIDFunc(ctx, boardID, columnID, name, updatedAt)
}

func (m *MockColumnRepository) Move(
	ctx context.Context,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	targetPosition domain.ColumnPosition,
) (domain.ColumnPosition, error) {
	AssertFuncNotNil("ColumnRepository.MoveFunc", m.MoveFunc)
	return m.MoveFunc(ctx, boardID, columnID, targetPosition)
}

func (m *MockColumnRepository) Delete(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error {
	AssertFuncNotNil("ColumnRepository.DeleteFunc", m.DeleteFunc)
	return m.DeleteFunc(ctx, boardID, columnID)
}
