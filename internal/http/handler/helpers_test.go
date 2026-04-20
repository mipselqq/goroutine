package handler_test

import (
	"context"
	"fmt"

	"goroutine/internal/domain"
	"goroutine/internal/testutil"
)

type MockAuth struct {
	RegisterFunc func(ctx context.Context, email domain.Email, password domain.UserPassword) error
	LoginFunc    func(ctx context.Context, email domain.Email, password domain.UserPassword) (string, error)
}

func AssertFuncNotNil(funcName string, fn any) {
	if fn == nil {
		panic(fmt.Sprintf("BUG: Mock: %s is not set", funcName))
	}
}

func (m *MockAuth) Register(ctx context.Context, email domain.Email, password domain.UserPassword) error {
	AssertFuncNotNil("AuthService.RegisterFunc", m.RegisterFunc)
	return m.RegisterFunc(ctx, email, password)
}

func (m *MockAuth) Login(ctx context.Context, email domain.Email, password domain.UserPassword) (string, error) {
	AssertFuncNotNil("AuthService.LoginFunc", m.LoginFunc)
	return m.LoginFunc(ctx, email, password)
}

type MockBoards struct {
	CreateFunc     func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error)
	GetFunc        func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (domain.Board, error)
	GetManyFunc    func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error)
	UpdateByIDFunc func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error)
	DeleteFunc     func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) error
}

type MockColumns struct {
	CreateFunc     func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName) (domain.Column, error)
	ListFunc       func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) ([]domain.Column, error)
	UpdateByIDFunc func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName) (domain.Column, error)
	MoveFunc       func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error)
	DeleteFunc     func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) error
}

func (m *MockBoards) Get(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (domain.Board, error) {
	AssertFuncNotNil("BoardsService.GetFunc", m.GetFunc)
	return m.GetFunc(ctx, ownerID, boardID)
}

func (m *MockBoards) GetMany(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
	AssertFuncNotNil("BoardsService.GetManyFunc", m.GetManyFunc)
	return m.GetManyFunc(ctx, ownerID)
}

func (m *MockBoards) Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
	AssertFuncNotNil("BoardsService.CreateFunc", m.CreateFunc)
	return m.CreateFunc(ctx, ownerID, name, description)
}

func (m *MockBoards) UpdateByID(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
	AssertFuncNotNil("BoardsService.UpdateByIDFunc", m.UpdateByIDFunc)
	return m.UpdateByIDFunc(ctx, ownerID, boardID, name, description)
}

func (m *MockBoards) Delete(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) error {
	AssertFuncNotNil("BoardsService.DeleteFunc", m.DeleteFunc)
	return m.DeleteFunc(ctx, ownerID, boardID)
}

func (m *MockColumns) Create(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName) (domain.Column, error) {
	AssertFuncNotNil("ColumnsService.CreateFunc", m.CreateFunc)
	return m.CreateFunc(ctx, callerID, boardID, name)
}

func (m *MockColumns) List(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) ([]domain.Column, error) {
	AssertFuncNotNil("ColumnsService.ListFunc", m.ListFunc)
	return m.ListFunc(ctx, callerID, boardID)
}

func (m *MockColumns) UpdateByID(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName) (domain.Column, error) {
	AssertFuncNotNil("ColumnsService.UpdateByIDFunc", m.UpdateByIDFunc)
	return m.UpdateByIDFunc(ctx, callerID, boardID, columnID, name)
}

func (m *MockColumns) Move(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error) {
	AssertFuncNotNil("ColumnsService.MoveFunc", m.MoveFunc)
	return m.MoveFunc(ctx, callerID, boardID, columnID, targetPosition)
}

func (m *MockColumns) Delete(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) error {
	AssertFuncNotNil("ColumnsService.DeleteFunc", m.DeleteFunc)
	return m.DeleteFunc(ctx, callerID, boardID, columnID)
}

func columnNotFoundErrorBody() map[string]any {
	return map[string]any{
		"code":      "COLUMN_NOT_FOUND",
		"message":   "Column not found",
		"timestamp": testutil.FixedTimeNowStr(),
		"details": []any{
			map[string]any{"field": "columnId", "issues": []string{"Column not found"}},
		},
	}
}

func invalidJsonBody() map[string]any {
	return map[string]any{
		"code":      "VALIDATION_ERROR",
		"message":   "Some fields are invalid",
		"timestamp": testutil.FixedTimeNowStr(),
		"details": []any{
			map[string]any{"field": "body", "issues": []string{"Invalid JSON body"}},
		},
	}
}

func internalErrorBody() map[string]any {
	return map[string]any{
		"code":      "INTERNAL_SERVER_ERROR",
		"message":   "Internal server error",
		"timestamp": testutil.FixedTimeNowStr(),
	}
}

func notFoundErrorBody() map[string]any {
	return map[string]any{
		"code":      "NOT_FOUND",
		"message":   "Resource not found",
		"timestamp": testutil.FixedTimeNowStr(),
	}
}

func boardNotFoundErrorBody() map[string]any {
	return map[string]any{
		"code":      "BOARD_NOT_FOUND",
		"message":   "Board not found",
		"timestamp": testutil.FixedTimeNowStr(),
		"details": []any{
			map[string]any{"field": "boardId", "issues": []string{"Board not found"}},
		},
	}
}

func unauthorizedTokenBody() map[string]any {
	return map[string]any{
		"code":      "INVALID_TOKEN",
		"message":   "Invalid token",
		"timestamp": testutil.FixedTimeNowStr(),
		"details": []any{
			map[string]any{"field": "Authorization", "issues": []string{"Invalid token"}},
		},
	}
}

func validationErrorBody(field string, issues []string) map[string]any {
	return map[string]any{
		"code":      "VALIDATION_ERROR",
		"message":   "Some fields are invalid",
		"timestamp": testutil.FixedTimeNowStr(),
		"details": []any{
			map[string]any{"field": field, "issues": issues},
		},
	}
}
