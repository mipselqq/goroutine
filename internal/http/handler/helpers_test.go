package handler_test

import (
	"context"
	"fmt"

	"goroutine/internal/domain"
	"goroutine/internal/service"
	"goroutine/internal/testutil"
)

type MockAuthService struct {
	RegisterFunc func(ctx context.Context, email domain.Email, password domain.UserPassword) error
	LoginFunc    func(ctx context.Context, email domain.Email, password domain.UserPassword) (string, error)
}

func AssertFuncNotNil(funcName string, fn any) {
	if fn == nil {
		panic(fmt.Sprintf("%s = nil, want configured mock", funcName))
	}
}

func (m *MockAuthService) Register(ctx context.Context, email domain.Email, password domain.UserPassword) error {
	AssertFuncNotNil("AuthService.RegisterFunc", m.RegisterFunc)
	return m.RegisterFunc(ctx, email, password)
}

func (m *MockAuthService) Login(ctx context.Context, email domain.Email, password domain.UserPassword) (string, error) {
	AssertFuncNotNil("AuthService.LoginFunc", m.LoginFunc)
	return m.LoginFunc(ctx, email, password)
}

type MockBoardService struct {
	CreateFunc       func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error)
	GetFunc          func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (domain.Board, error)
	GetAggregateFunc func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (service.AggregateBoard, error)
	GetManyFunc      func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error)
	UpdateByIDFunc   func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error)
	DeleteFunc       func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) error
}

type MockColumnService struct {
	CreateFunc     func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName) (domain.Column, error)
	ListFunc       func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) ([]domain.Column, error)
	UpdateByIDFunc func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName) (domain.Column, error)
	MoveFunc       func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error)
	DeleteFunc     func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) error
}

type MockTaskService struct {
	CreateFunc     func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name domain.TaskName, description domain.TaskDescription) (domain.Task, error)
	ListFunc       func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) ([]domain.Task, error)
	UpdateByIDFunc func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error)
	MoveFunc       func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, targetColumnID domain.ColumnID, targetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error)
	DeleteFunc     func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID) error
}

func (m *MockBoardService) Get(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (domain.Board, error) {
	AssertFuncNotNil("BoardsService.GetFunc", m.GetFunc)
	return m.GetFunc(ctx, ownerID, boardID)
}

func (m *MockBoardService) GetAggregate(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (service.AggregateBoard, error) {
	AssertFuncNotNil("BoardsService.GetAggregateFunc", m.GetAggregateFunc)
	return m.GetAggregateFunc(ctx, ownerID, boardID)
}

func (m *MockBoardService) GetMany(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
	AssertFuncNotNil("BoardsService.GetManyFunc", m.GetManyFunc)
	return m.GetManyFunc(ctx, ownerID)
}

func (m *MockBoardService) Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
	AssertFuncNotNil("BoardsService.CreateFunc", m.CreateFunc)
	return m.CreateFunc(ctx, ownerID, name, description)
}

func (m *MockBoardService) UpdateByID(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
	AssertFuncNotNil("BoardsService.UpdateByIDFunc", m.UpdateByIDFunc)
	return m.UpdateByIDFunc(ctx, ownerID, boardID, name, description)
}

func (m *MockBoardService) Delete(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) error {
	AssertFuncNotNil("BoardsService.DeleteFunc", m.DeleteFunc)
	return m.DeleteFunc(ctx, ownerID, boardID)
}

func (m *MockColumnService) Create(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName) (domain.Column, error) {
	AssertFuncNotNil("ColumnsService.CreateFunc", m.CreateFunc)
	return m.CreateFunc(ctx, callerID, boardID, name)
}

func (m *MockColumnService) List(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) ([]domain.Column, error) {
	AssertFuncNotNil("ColumnsService.ListFunc", m.ListFunc)
	return m.ListFunc(ctx, callerID, boardID)
}

func (m *MockColumnService) UpdateByID(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName) (domain.Column, error) {
	AssertFuncNotNil("ColumnsService.UpdateByIDFunc", m.UpdateByIDFunc)
	return m.UpdateByIDFunc(ctx, callerID, boardID, columnID, name)
}

func (m *MockColumnService) Move(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error) {
	AssertFuncNotNil("ColumnsService.MoveFunc", m.MoveFunc)
	return m.MoveFunc(ctx, callerID, boardID, columnID, targetPosition)
}

func (m *MockColumnService) Delete(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) error {
	AssertFuncNotNil("ColumnsService.DeleteFunc", m.DeleteFunc)
	return m.DeleteFunc(ctx, callerID, boardID, columnID)
}

func (m *MockTaskService) Create(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name domain.TaskName, description domain.TaskDescription) (domain.Task, error) {
	AssertFuncNotNil("TasksService.CreateFunc", m.CreateFunc)
	return m.CreateFunc(ctx, callerID, boardID, columnID, name, description)
}

func (m *MockTaskService) List(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) ([]domain.Task, error) {
	AssertFuncNotNil("TasksService.ListFunc", m.ListFunc)
	return m.ListFunc(ctx, callerID, boardID, columnID)
}

func (m *MockTaskService) UpdateByID(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error) {
	AssertFuncNotNil("TasksService.UpdateByIDFunc", m.UpdateByIDFunc)
	return m.UpdateByIDFunc(ctx, callerID, boardID, columnID, taskID, name, description)
}

func (m *MockTaskService) Move(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, targetColumnID domain.ColumnID, targetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error) {
	AssertFuncNotNil("TasksService.MoveFunc", m.MoveFunc)
	return m.MoveFunc(ctx, callerID, boardID, columnID, taskID, targetColumnID, targetPosition)
}

func (m *MockTaskService) Delete(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID) error {
	AssertFuncNotNil("TasksService.DeleteFunc", m.DeleteFunc)
	return m.DeleteFunc(ctx, callerID, boardID, columnID, taskID)
}

func columnNotFoundByFieldError(field string) map[string]any {
	return map[string]any{
		"code":      "COLUMN_NOT_FOUND",
		"message":   "Column not found",
		"timestamp": testutil.FixedTimeNowStr(),
		"details": []any{
			map[string]any{"field": field, "issues": []string{"Column not found"}},
		},
	}
}

func taskNotFoundErrorBody() map[string]any {
	return map[string]any{
		"code":      "TASK_NOT_FOUND",
		"message":   "Task not found",
		"timestamp": testutil.FixedTimeNowStr(),
		"details": []any{
			map[string]any{"field": "taskId", "issues": []string{"Task not found"}},
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
		"details":   []any{},
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

func userAlreadyExistsErrorBody() map[string]any {
	return map[string]any{
		"code":      "USER_ALREADY_EXISTS",
		"message":   "User already exists",
		"timestamp": testutil.FixedTimeNowStr(),
		"details": []any{
			map[string]any{"field": "email", "issues": []string{"Email already registered"}},
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
