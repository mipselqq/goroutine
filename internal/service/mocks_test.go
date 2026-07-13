package service_test

import (
	"context"
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/testutil"
)

type MockUserRepository struct {
	t *testing.T

	CreateFunc             func(ctx context.Context, email domain.Email, hash domain.PasswordHash) error
	GetByEmailFunc         func(ctx context.Context, email domain.Email) (domain.User, error)
	UpdateTelegramInfoFunc func(ctx context.Context, userID domain.UserID, chatID domain.TelegramChatID, username domain.TelegramUsername) error
}

func NewMockUserRepository(t *testing.T) *MockUserRepository {
	return &MockUserRepository{t: t}
}

func (m *MockUserRepository) Create(ctx context.Context, email domain.Email, hash domain.PasswordHash) error {
	testutil.AssertFuncNotNil(m.t, "UserRepository.CreateFunc", m.CreateFunc)
	return m.CreateFunc(ctx, email, hash)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email domain.Email) (domain.User, error) {
	testutil.AssertFuncNotNil(m.t, "UserRepository.GetByEmailFunc", m.GetByEmailFunc)
	return m.GetByEmailFunc(ctx, email)
}

func (m *MockUserRepository) UpdateTelegramInfo(ctx context.Context, userID domain.UserID, chatID domain.TelegramChatID, username domain.TelegramUsername) error {
	testutil.AssertFuncNotNil(m.t, "UserRepository.UpdateTelegramInfoFunc", m.UpdateTelegramInfoFunc)
	return m.UpdateTelegramInfoFunc(ctx, userID, chatID, username)
}

type MockTelegramTokenRepository struct {
	t *testing.T

	InsertLinkTokenFunc          func(ctx context.Context, token domain.TelegramLinkToken, userID domain.UserID) error
	ConsumeTelegramLinkTokenFunc func(ctx context.Context, token domain.TelegramLinkToken) (domain.UserID, error)
}

func NewMockTelegramTokenRepository(t *testing.T) *MockTelegramTokenRepository {
	return &MockTelegramTokenRepository{t: t}
}

func (m *MockTelegramTokenRepository) InsertLinkToken(ctx context.Context, token domain.TelegramLinkToken, userID domain.UserID) error {
	testutil.AssertFuncNotNil(m.t, "TelegramTokenRepository.InsertLinkTokenFunc", m.InsertLinkTokenFunc)
	return m.InsertLinkTokenFunc(ctx, token, userID)
}

func (m *MockTelegramTokenRepository) ConsumeTelegramLinkToken(ctx context.Context, token domain.TelegramLinkToken) (domain.UserID, error) {
	testutil.AssertFuncNotNil(m.t, "TelegramTokenRepository.ConsumeTelegramLinkTokenFunc", m.ConsumeTelegramLinkTokenFunc)
	return m.ConsumeTelegramLinkTokenFunc(ctx, token)
}

type MockTelegramNotifier struct {
	t *testing.T

	NotifyFunc func(ctx context.Context, chatID domain.TelegramChatID, text string) error
}

func NewMockTelegramNotifier(t *testing.T) *MockTelegramNotifier {
	return &MockTelegramNotifier{t: t}
}

func (m *MockTelegramNotifier) Notify(ctx context.Context, chatID domain.TelegramChatID, text string) error {
	testutil.AssertFuncNotNil(m.t, "TelegramNotifier.NotifyFunc", m.NotifyFunc)
	return m.NotifyFunc(ctx, chatID, text)
}

type MockBoardRepository struct {
	t *testing.T

	CreateFunc        func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error)
	GetFunc           func(ctx context.Context, id domain.BoardID) (domain.Board, error)
	ListByOwnerIDFunc func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error)
	UpdateFunc        func(ctx context.Context, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error)
	DeleteFunc        func(ctx context.Context, boardID domain.BoardID) error
}

func NewMockBoardRepository(t *testing.T) *MockBoardRepository {
	return &MockBoardRepository{t: t}
}

type MockColumnRepository struct {
	t *testing.T

	CreateFunc        func(ctx context.Context, boardID domain.BoardID, name domain.ColumnName, description domain.ColumnDescription) (domain.Column, error)
	ListByBoardIDFunc func(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error)
	GetFunc           func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error)
	UpdateFunc        func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error)
	MoveFunc          func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error)
	DeleteFunc        func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error
}

func NewMockColumnRepository(t *testing.T) *MockColumnRepository {
	return &MockColumnRepository{t: t}
}

func (m *MockBoardRepository) Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
	testutil.AssertFuncNotNil(m.t, "BoardRepository.CreateFunc", m.CreateFunc)
	return m.CreateFunc(ctx, ownerID, name, description)
}

func (m *MockBoardRepository) Get(ctx context.Context, id domain.BoardID) (domain.Board, error) {
	testutil.AssertFuncNotNil(m.t, "BoardRepository.GetFunc", m.GetFunc)
	return m.GetFunc(ctx, id)
}

func (m *MockBoardRepository) ListByOwnerID(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
	testutil.AssertFuncNotNil(m.t, "BoardRepository.ListByOwnerIDFunc", m.ListByOwnerIDFunc)
	return m.ListByOwnerIDFunc(ctx, ownerID)
}

func (m *MockBoardRepository) Update(ctx context.Context, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
	testutil.AssertFuncNotNil(m.t, "BoardRepository.UpdateFunc", m.UpdateFunc)
	return m.UpdateFunc(ctx, boardID, name, description)
}

func (m *MockBoardRepository) Delete(ctx context.Context, boardID domain.BoardID) error {
	testutil.AssertFuncNotNil(m.t, "BoardRepository.DeleteFunc", m.DeleteFunc)
	return m.DeleteFunc(ctx, boardID)
}

func (m *MockColumnRepository) Create(
	ctx context.Context,
	boardID domain.BoardID,
	name domain.ColumnName,
	description domain.ColumnDescription,
) (domain.Column, error) {
	testutil.AssertFuncNotNil(m.t, "ColumnRepository.CreateFunc", m.CreateFunc)
	return m.CreateFunc(ctx, boardID, name, description)
}

func (m *MockColumnRepository) ListByBoardID(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error) {
	testutil.AssertFuncNotNil(m.t, "ColumnRepository.ListByBoardIDFunc", m.ListByBoardIDFunc)
	return m.ListByBoardIDFunc(ctx, boardID)
}

func (m *MockColumnRepository) Get(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
	testutil.AssertFuncNotNil(m.t, "ColumnRepository.GetFunc", m.GetFunc)
	return m.GetFunc(ctx, columnID)
}

func (m *MockColumnRepository) Update(
	ctx context.Context,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	name *domain.ColumnName,
	description *domain.ColumnDescription,
) (domain.Column, error) {
	testutil.AssertFuncNotNil(m.t, "ColumnRepository.UpdateFunc", m.UpdateFunc)
	return m.UpdateFunc(ctx, boardID, columnID, name, description)
}

func (m *MockColumnRepository) Move(
	ctx context.Context,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	targetPosition domain.ColumnPosition,
) (domain.ColumnPosition, error) {
	testutil.AssertFuncNotNil(m.t, "ColumnRepository.MoveFunc", m.MoveFunc)
	return m.MoveFunc(ctx, boardID, columnID, targetPosition)
}

func (m *MockColumnRepository) Delete(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error {
	testutil.AssertFuncNotNil(m.t, "ColumnRepository.DeleteFunc", m.DeleteFunc)
	return m.DeleteFunc(ctx, boardID, columnID)
}

type MockTaskRepository struct {
	t *testing.T

	CreateFunc         func(ctx context.Context, columnID domain.ColumnID, name domain.TaskName, description domain.TaskDescription) (domain.Task, error)
	ListByBoardIDFunc  func(ctx context.Context, boardID domain.BoardID) ([]domain.Task, error)
	ListByColumnIDFunc func(ctx context.Context, columnID domain.ColumnID) ([]domain.Task, error)
	GetFunc            func(ctx context.Context, taskID domain.TaskID) (domain.Task, error)
	UpdateFunc         func(ctx context.Context, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error)
	MoveFunc           func(ctx context.Context, boardID domain.BoardID, currentColumnID domain.ColumnID, taskID domain.TaskID, targetColumnID domain.ColumnID, targetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error)
	DeleteFunc         func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID) error
}

func NewMockTaskRepository(t *testing.T) *MockTaskRepository {
	return &MockTaskRepository{t: t}
}

func (m *MockTaskRepository) Create(
	ctx context.Context,
	columnID domain.ColumnID,
	name domain.TaskName,
	description domain.TaskDescription,
) (domain.Task, error) {
	testutil.AssertFuncNotNil(m.t, "TaskRepository.CreateFunc", m.CreateFunc)
	return m.CreateFunc(ctx, columnID, name, description)
}

func (m *MockTaskRepository) ListByBoardID(ctx context.Context, boardID domain.BoardID) ([]domain.Task, error) {
	testutil.AssertFuncNotNil(m.t, "TaskRepository.ListByBoardIDFunc", m.ListByBoardIDFunc)
	return m.ListByBoardIDFunc(ctx, boardID)
}

func (m *MockTaskRepository) ListByColumnID(ctx context.Context, columnID domain.ColumnID) ([]domain.Task, error) {
	testutil.AssertFuncNotNil(m.t, "TaskRepository.ListByColumnIDFunc", m.ListByColumnIDFunc)
	return m.ListByColumnIDFunc(ctx, columnID)
}

func (m *MockTaskRepository) Get(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
	testutil.AssertFuncNotNil(m.t, "TaskRepository.GetFunc", m.GetFunc)
	return m.GetFunc(ctx, taskID)
}

func (m *MockTaskRepository) Update(
	ctx context.Context,
	columnID domain.ColumnID,
	taskID domain.TaskID,
	name *domain.TaskName,
	description *domain.TaskDescription,
) (domain.Task, error) {
	testutil.AssertFuncNotNil(m.t, "TaskRepository.UpdateFunc", m.UpdateFunc)
	return m.UpdateFunc(ctx, columnID, taskID, name, description)
}

func (m *MockTaskRepository) Move(
	ctx context.Context,
	boardID domain.BoardID,
	currentColumnID domain.ColumnID,
	taskID domain.TaskID,
	targetColumnID domain.ColumnID,
	targetPosition domain.TaskPosition,
) (domain.ColumnID, domain.TaskPosition, error) {
	testutil.AssertFuncNotNil(m.t, "TaskRepository.MoveFunc", m.MoveFunc)
	return m.MoveFunc(ctx, boardID, currentColumnID, taskID, targetColumnID, targetPosition)
}

func (m *MockTaskRepository) Delete(
	ctx context.Context,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	taskID domain.TaskID,
) error {
	testutil.AssertFuncNotNil(m.t, "TaskRepository.DeleteFunc", m.DeleteFunc)
	return m.DeleteFunc(ctx, boardID, columnID, taskID)
}
