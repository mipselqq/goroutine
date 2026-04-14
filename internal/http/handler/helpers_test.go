package handler_test

import (
	"context"

	"goroutine/internal/domain"
	"goroutine/internal/testutil"
)

type MockAuth struct {
	RegisterFunc func(ctx context.Context, email domain.Email, password domain.UserPassword) error
	LoginFunc    func(ctx context.Context, email domain.Email, password domain.UserPassword) (string, error)
}

func (m *MockAuth) Register(ctx context.Context, email domain.Email, password domain.UserPassword) error {
	if m.RegisterFunc == nil {
		panic("BUG: Mock: AuthService.RegisterFunc is not set")
	}
	return m.RegisterFunc(ctx, email, password)
}

func (m *MockAuth) Login(ctx context.Context, email domain.Email, password domain.UserPassword) (string, error) {
	if m.LoginFunc == nil {
		panic("BUG: Mock: AuthService.LoginFunc is not set")
	}
	return m.LoginFunc(ctx, email, password)
}

type MockBoards struct {
	CreateFunc     func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error)
	GetFunc        func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (domain.Board, error)
	GetManyFunc    func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error)
	UpdateByIDFunc func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error)
	DeleteFunc     func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) error
}

func (m *MockBoards) Get(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (domain.Board, error) {
	if m.GetFunc == nil {
		panic("BUG: Mock: BoardsService.GetFunc is not set")
	}
	return m.GetFunc(ctx, ownerID, boardID)
}

func (m *MockBoards) GetMany(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
	if m.GetManyFunc == nil {
		panic("BUG: Mock: BoardsService.GetManyFunc is not set")
	}
	return m.GetManyFunc(ctx, ownerID)
}

func (m *MockBoards) Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
	if m.CreateFunc == nil {
		panic("BUG: Mock: BoardsService.CreateFunc is not set")
	}
	return m.CreateFunc(ctx, ownerID, name, description)
}

func (m *MockBoards) UpdateByID(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
	if m.UpdateByIDFunc == nil {
		panic("BUG: Mock: BoardsService.UpdateByIDFunc is not set")
	}
	return m.UpdateByIDFunc(ctx, ownerID, boardID, name, description)
}

func (m *MockBoards) Delete(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) error {
	if m.DeleteFunc == nil {
		panic("BUG: Mock: BoardsService.DeleteFunc is not set")
	}
	return m.DeleteFunc(ctx, ownerID, boardID)
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
