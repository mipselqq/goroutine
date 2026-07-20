package handler_test

import (
	"context"
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/service"
	"goroutine/internal/testutil"
)

type MockAuthService struct {
	t *testing.T

	RegisterFunc func(ctx context.Context, email domain.Email, password domain.UserPassword) error
	LoginFunc    func(ctx context.Context, email domain.Email, password domain.UserPassword) (domain.AuthToken, error)
}

func NewMockAuthService(t *testing.T) *MockAuthService {
	return &MockAuthService{t: t}
}

type MockBoardService struct {
	t *testing.T

	CreateFunc        func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error)
	GetFunc           func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (domain.Board, error)
	GetAggregateFunc  func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (service.AggregateBoard, error)
	ListByOwnerIDFunc func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error)
	UpdateFunc        func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error)
	DeleteFunc        func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) error
}

func NewMockBoardService(t *testing.T) *MockBoardService {
	return &MockBoardService{t: t}
}

type MockColumnService struct {
	t *testing.T

	CreateFunc        func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName, description domain.ColumnDescription) (domain.Column, error)
	ListByBoardIDFunc func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) ([]domain.Column, error)
	UpdateFunc        func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error)
	MoveFunc          func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error)
	DeleteFunc        func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) error
}

func NewMockColumnService(t *testing.T) *MockColumnService {
	return &MockColumnService{t: t}
}

type MockTaskService struct {
	t *testing.T

	CreateFunc         func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name domain.TaskName, description domain.TaskDescription) (domain.Task, error)
	ListByColumnIDFunc func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) ([]domain.Task, error)
	UpdateFunc         func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error)
	MoveFunc           func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, targetColumnID domain.ColumnID, targetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error)
	DeleteFunc         func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID) error
}

func NewMockTaskService(t *testing.T) *MockTaskService {
	return &MockTaskService{t: t}
}

func (m *MockAuthService) Register(ctx context.Context, email domain.Email, password domain.UserPassword) error {
	testutil.AssertFuncNotNil(m.t, "authService.RegisterFunc", m.RegisterFunc)
	return m.RegisterFunc(ctx, email, password)
}

func (m *MockAuthService) Login(ctx context.Context, email domain.Email, password domain.UserPassword) (domain.AuthToken, error) {
	testutil.AssertFuncNotNil(m.t, "authService.LoginFunc", m.LoginFunc)
	return m.LoginFunc(ctx, email, password)
}

type MockUserService struct {
	t *testing.T

	CreateTelegramLinkTokenFunc func(ctx context.Context, userID domain.UserID) (domain.TelegramLinkToken, error)
	LinkTelegramByTokenFunc     func(ctx context.Context, token domain.TelegramLinkToken, chatID domain.TelegramChatID, username domain.TelegramUsername) error
}

func NewMockUserService(t *testing.T) *MockUserService {
	return &MockUserService{t: t}
}

func (m *MockUserService) CreateTelegramLinkToken(ctx context.Context, userID domain.UserID) (domain.TelegramLinkToken, error) {
	testutil.AssertFuncNotNil(m.t, "userService.CreateTelegramLinkTokenFunc", m.CreateTelegramLinkTokenFunc)
	return m.CreateTelegramLinkTokenFunc(ctx, userID)
}

func (m *MockUserService) LinkTelegramByToken(ctx context.Context, token domain.TelegramLinkToken, chatID domain.TelegramChatID, username domain.TelegramUsername) error {
	testutil.AssertFuncNotNil(m.t, "userService.LinkTelegramByTokenFunc", m.LinkTelegramByTokenFunc)
	return m.LinkTelegramByTokenFunc(ctx, token, chatID, username)
}

func (m *MockBoardService) Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
	testutil.AssertFuncNotNil(m.t, "boardsService.CreateFunc", m.CreateFunc)
	return m.CreateFunc(ctx, ownerID, name, description)
}

func (m *MockBoardService) Get(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (domain.Board, error) {
	testutil.AssertFuncNotNil(m.t, "boardsService.GetFunc", m.GetFunc)
	return m.GetFunc(ctx, ownerID, boardID)
}

func (m *MockBoardService) GetAggregate(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (service.AggregateBoard, error) {
	testutil.AssertFuncNotNil(m.t, "boardsService.GetAggregateFunc", m.GetAggregateFunc)
	return m.GetAggregateFunc(ctx, ownerID, boardID)
}

func (m *MockBoardService) ListByOwnerID(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
	testutil.AssertFuncNotNil(m.t, "boardsService.ListByOwnerIDFunc", m.ListByOwnerIDFunc)
	return m.ListByOwnerIDFunc(ctx, ownerID)
}

func (m *MockBoardService) Update(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
	testutil.AssertFuncNotNil(m.t, "boardsService.UpdateFunc", m.UpdateFunc)
	return m.UpdateFunc(ctx, ownerID, boardID, name, description)
}

func (m *MockBoardService) Delete(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) error {
	testutil.AssertFuncNotNil(m.t, "boardsService.DeleteFunc", m.DeleteFunc)
	return m.DeleteFunc(ctx, ownerID, boardID)
}

func (m *MockColumnService) Create(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName, description domain.ColumnDescription) (domain.Column, error) {
	testutil.AssertFuncNotNil(m.t, "columnsService.CreateFunc", m.CreateFunc)
	return m.CreateFunc(ctx, callerID, boardID, name, description)
}

func (m *MockColumnService) ListByBoardID(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) ([]domain.Column, error) {
	testutil.AssertFuncNotNil(m.t, "columnsService.ListByBoardIDFunc", m.ListByBoardIDFunc)
	return m.ListByBoardIDFunc(ctx, callerID, boardID)
}

func (m *MockColumnService) Update(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error) {
	testutil.AssertFuncNotNil(m.t, "columnsService.UpdateFunc", m.UpdateFunc)
	return m.UpdateFunc(ctx, callerID, boardID, columnID, name, description)
}

func (m *MockColumnService) Move(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error) {
	testutil.AssertFuncNotNil(m.t, "columnsService.MoveFunc", m.MoveFunc)
	return m.MoveFunc(ctx, callerID, boardID, columnID, targetPosition)
}

func (m *MockColumnService) Delete(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) error {
	testutil.AssertFuncNotNil(m.t, "columnsService.DeleteFunc", m.DeleteFunc)
	return m.DeleteFunc(ctx, callerID, boardID, columnID)
}

func (m *MockTaskService) Create(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name domain.TaskName, description domain.TaskDescription) (domain.Task, error) {
	testutil.AssertFuncNotNil(m.t, "tasksService.CreateFunc", m.CreateFunc)
	return m.CreateFunc(ctx, callerID, boardID, columnID, name, description)
}

func (m *MockTaskService) ListByColumnID(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) ([]domain.Task, error) {
	testutil.AssertFuncNotNil(m.t, "tasksService.ListByColumnIDFunc", m.ListByColumnIDFunc)
	return m.ListByColumnIDFunc(ctx, callerID, boardID, columnID)
}

func (m *MockTaskService) Update(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error) {
	testutil.AssertFuncNotNil(m.t, "tasksService.UpdateFunc", m.UpdateFunc)
	return m.UpdateFunc(ctx, callerID, boardID, columnID, taskID, name, description)
}

func (m *MockTaskService) Move(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, targetColumnID domain.ColumnID, targetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error) {
	testutil.AssertFuncNotNil(m.t, "tasksService.MoveFunc", m.MoveFunc)
	return m.MoveFunc(ctx, callerID, boardID, columnID, taskID, targetColumnID, targetPosition)
}

func (m *MockTaskService) Delete(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID) error {
	testutil.AssertFuncNotNil(m.t, "tasksService.DeleteFunc", m.DeleteFunc)
	return m.DeleteFunc(ctx, callerID, boardID, columnID, taskID)
}

type MockNotifier struct {
	t *testing.T

	NotifyFunc func(ctx context.Context, chatID domain.TelegramChatID, text domain.TelegramMessage) error
}

func NewMockNotifier(t *testing.T) *MockNotifier {
	return &MockNotifier{t: t}
}

func (m *MockNotifier) Notify(ctx context.Context, chatID domain.TelegramChatID, text domain.TelegramMessage) error {
	testutil.AssertFuncNotNil(m.t, "notifier.NotifyFunc", m.NotifyFunc)
	return m.NotifyFunc(ctx, chatID, text)
}
