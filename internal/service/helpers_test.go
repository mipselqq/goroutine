package service_test

import (
	"context"
	"fmt"

	"goroutine/internal/domain"
)

func AssertFuncNotNil(funcName string, fn any) {
	if fn == nil {
		panic(fmt.Sprintf("%s = nil, want configured mock", funcName))
	}
}

type MockUserRepository struct {
	CreateFunc             func(ctx context.Context, email domain.Email, hash string) error
	GetByEmailFunc         func(ctx context.Context, email domain.Email) (domain.User, error)
	UpdateTelegramInfoFunc func(ctx context.Context, userID domain.UserID, chatID domain.TelegramChatID, username domain.TelegramUsername) error
}

func (m *MockUserRepository) Create(ctx context.Context, email domain.Email, hash string) error {
	AssertFuncNotNil("UserRepository.CreateFunc", m.CreateFunc)
	return m.CreateFunc(ctx, email, hash)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email domain.Email) (domain.User, error) {
	AssertFuncNotNil("UserRepository.GetByEmailFunc", m.GetByEmailFunc)
	return m.GetByEmailFunc(ctx, email)
}

func (m *MockUserRepository) UpdateTelegramInfo(ctx context.Context, userID domain.UserID, chatID domain.TelegramChatID, username domain.TelegramUsername) error {
	AssertFuncNotNil("UserRepository.UpdateTelegramInfoFunc", m.UpdateTelegramInfoFunc)
	return m.UpdateTelegramInfoFunc(ctx, userID, chatID, username)
}

type MockTelegramTokenRepository struct {
	InsertLinkTokenFunc          func(ctx context.Context, token domain.TelegramLinkToken, userID domain.UserID) error
	ConsumeTelegramLinkTokenFunc func(ctx context.Context, token domain.TelegramLinkToken) (domain.UserID, error)
}

func (m *MockTelegramTokenRepository) InsertLinkToken(ctx context.Context, token domain.TelegramLinkToken, userID domain.UserID) error {
	AssertFuncNotNil("TelegramTokenRepository.InsertLinkTokenFunc", m.InsertLinkTokenFunc)
	return m.InsertLinkTokenFunc(ctx, token, userID)
}

func (m *MockTelegramTokenRepository) ConsumeTelegramLinkToken(ctx context.Context, token domain.TelegramLinkToken) (domain.UserID, error) {
	AssertFuncNotNil("TelegramTokenRepository.ConsumeTelegramLinkTokenFunc", m.ConsumeTelegramLinkTokenFunc)
	return m.ConsumeTelegramLinkTokenFunc(ctx, token)
}

type MockTelegramNotifier struct {
	NotifyFunc func(ctx context.Context, chatID domain.TelegramChatID, text string) error
}

func (m *MockTelegramNotifier) Notify(ctx context.Context, chatID domain.TelegramChatID, text string) error {
	AssertFuncNotNil("TelegramNotifier.NotifyFunc", m.NotifyFunc)
	return m.NotifyFunc(ctx, chatID, text)
}

type MockBoardRepository struct {
	CreateFunc func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error)
	GetFunc    func(ctx context.Context, id domain.BoardID) (domain.Board, error)
	ListFunc   func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error)
	UpdateFunc func(ctx context.Context, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error)
	DeleteFunc func(ctx context.Context, boardID domain.BoardID) error
}

type MockColumnRepository struct {
	CreateFunc func(ctx context.Context, boardID domain.BoardID, name domain.ColumnName, description domain.ColumnDescription) (domain.Column, error)
	ListFunc   func(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error)
	GetFunc    func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error)
	UpdateFunc func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error)
	MoveFunc   func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error)
	DeleteFunc func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error
}

func (m *MockBoardRepository) Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
	AssertFuncNotNil("BoardRepository.CreateFunc", m.CreateFunc)
	return m.CreateFunc(ctx, ownerID, name, description)
}

func (m *MockBoardRepository) Get(ctx context.Context, id domain.BoardID) (domain.Board, error) {
	AssertFuncNotNil("BoardRepository.GetFunc", m.GetFunc)
	return m.GetFunc(ctx, id)
}

func (m *MockBoardRepository) List(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
	AssertFuncNotNil("BoardRepository.ListFunc", m.ListFunc)
	return m.ListFunc(ctx, ownerID)
}

func (m *MockBoardRepository) Update(ctx context.Context, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
	AssertFuncNotNil("BoardRepository.UpdateFunc", m.UpdateFunc)
	return m.UpdateFunc(ctx, boardID, name, description)
}

func (m *MockBoardRepository) Delete(ctx context.Context, boardID domain.BoardID) error {
	AssertFuncNotNil("BoardRepository.DeleteFunc", m.DeleteFunc)
	return m.DeleteFunc(ctx, boardID)
}

func (m *MockColumnRepository) Create(
	ctx context.Context,
	boardID domain.BoardID,
	name domain.ColumnName,
	description domain.ColumnDescription,
) (domain.Column, error) {
	AssertFuncNotNil("ColumnRepository.CreateFunc", m.CreateFunc)
	return m.CreateFunc(ctx, boardID, name, description)
}

func (m *MockColumnRepository) List(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error) {
	AssertFuncNotNil("ColumnRepository.ListFunc", m.ListFunc)
	return m.ListFunc(ctx, boardID)
}

func (m *MockColumnRepository) Get(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
	AssertFuncNotNil("ColumnRepository.GetFunc", m.GetFunc)
	return m.GetFunc(ctx, columnID)
}

func (m *MockColumnRepository) Update(
	ctx context.Context,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	name *domain.ColumnName,
	description *domain.ColumnDescription,
) (domain.Column, error) {
	AssertFuncNotNil("ColumnRepository.UpdateFunc", m.UpdateFunc)
	return m.UpdateFunc(ctx, boardID, columnID, name, description)
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

type MockTaskRepository struct {
	CreateFunc         func(ctx context.Context, columnID domain.ColumnID, name domain.TaskName, description domain.TaskDescription) (domain.Task, error)
	ListByBoardIDFunc  func(ctx context.Context, boardID domain.BoardID) ([]domain.Task, error)
	ListByColumnIDFunc func(ctx context.Context, columnID domain.ColumnID) ([]domain.Task, error)
	GetFunc            func(ctx context.Context, taskID domain.TaskID) (domain.Task, error)
	UpdateFunc         func(ctx context.Context, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error)
	MoveFunc           func(ctx context.Context, boardID domain.BoardID, currentColumnID domain.ColumnID, taskID domain.TaskID, targetColumnID domain.ColumnID, targetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error)
	DeleteFunc         func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID) error
}

func (m *MockTaskRepository) Create(
	ctx context.Context,
	columnID domain.ColumnID,
	name domain.TaskName,
	description domain.TaskDescription,
) (domain.Task, error) {
	AssertFuncNotNil("TaskRepository.CreateFunc", m.CreateFunc)
	return m.CreateFunc(ctx, columnID, name, description)
}

func (m *MockTaskRepository) ListByBoardID(ctx context.Context, boardID domain.BoardID) ([]domain.Task, error) {
	AssertFuncNotNil("TaskRepository.ListByBoardIDFunc", m.ListByBoardIDFunc)
	return m.ListByBoardIDFunc(ctx, boardID)
}

func (m *MockTaskRepository) ListByColumnID(ctx context.Context, columnID domain.ColumnID) ([]domain.Task, error) {
	AssertFuncNotNil("TaskRepository.ListByColumnIDFunc", m.ListByColumnIDFunc)
	return m.ListByColumnIDFunc(ctx, columnID)
}

func (m *MockTaskRepository) Get(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
	AssertFuncNotNil("TaskRepository.GetFunc", m.GetFunc)
	return m.GetFunc(ctx, taskID)
}

func (m *MockTaskRepository) Update(
	ctx context.Context,
	columnID domain.ColumnID,
	taskID domain.TaskID,
	name *domain.TaskName,
	description *domain.TaskDescription,
) (domain.Task, error) {
	AssertFuncNotNil("TaskRepository.UpdateFunc", m.UpdateFunc)
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
	AssertFuncNotNil("TaskRepository.MoveFunc", m.MoveFunc)
	return m.MoveFunc(ctx, boardID, currentColumnID, taskID, targetColumnID, targetPosition)
}

func (m *MockTaskRepository) Delete(
	ctx context.Context,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	taskID domain.TaskID,
) error {
	AssertFuncNotNil("TaskRepository.DeleteFunc", m.DeleteFunc)
	return m.DeleteFunc(ctx, boardID, columnID, taskID)
}
