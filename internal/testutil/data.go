package testutil

import (
	"fmt"
	"testing"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/secrecy"
	"goroutine/internal/service"

	"github.com/golang-jwt/jwt/v5"
)

func FixedTimeNow() time.Time       { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }
func FixedTime5mFromNow() time.Time { return FixedTimeNow().Add(5 * time.Minute) }

const timeFormat = "2006-01-02T15:04:05.000Z07:00"

func FixedTime5mFromNowStr() string {
	return FixedTime5mFromNow().UTC().Format(timeFormat)
}

func FixedTimeNowStr() string { return FixedTimeNow().UTC().Format(timeFormat) }

func must[T any](fn func(string) (T, error), s string) T {
	v, err := fn(s)
	if err != nil {
		panic(fmt.Errorf("testutil: BUG: value is no longer valid: %w", err))
	}
	return v
}

func ValidUserID() domain.UserID {
	return must(domain.ParseUserID, "018e1000-0000-7000-8000-000000000000")
}

func ValidEmail() domain.Email {
	return must(domain.NewEmail, "test@example.com")
}

func ValidPassword() domain.UserPassword {
	return must(domain.NewUserPassword, "qwerty")
}

func ValidPasswordHash() string {
	return "$argon2id$v=19$m=65536,t=1,p=16$kUYJyX3h53cARKnKqFZxvQ$IXz2KOKbyVklgyVmz9ebJ1ffOgmcyMpn/GTUWsep5lk"
}

func AnotherValidPasswordHash() string {
	return "$argon2id$v=19$m=65536,t=3,p=4$bm90LXF3ZXJ0eQ$fSowp1Rof0fXhF+rXv2f6w"
}

func ValidBoardName() domain.BoardName {
	return must(domain.NewBoardName, "Test Board")
}

func ValidBoardDescription() domain.BoardDescription {
	return must(domain.NewBoardDescription, "Test Board Description")
}

func ValidColumnName() domain.ColumnName {
	return must(domain.NewColumnName, "To Do")
}

func ValidColumnPosition() domain.ColumnPosition {
	position, err := domain.NewColumnPosition(1)
	if err != nil {
		panic(fmt.Errorf("testutil: BUG: value is no longer valid: %w", err))
	}

	return position
}

func ValidTaskName() domain.TaskName {
	return must(domain.NewTaskName, "Write tests")
}

func ValidTaskDescription() domain.TaskDescription {
	return must(domain.NewTaskDescription, "Cover the new endpoint with tests")
}

func ValidTaskPosition() domain.TaskPosition {
	position, err := domain.NewTaskPosition(1)
	if err != nil {
		panic(fmt.Errorf("testutil: BUG: value is no longer valid: %w", err))
	}

	return position
}

func ValidJWTSecret() secrecy.SecretString {
	return secrecy.SecretString("secret")
}

func ValidJWTOptions() service.JWTOptions {
	return service.JWTOptions{
		JWTSecret:     ValidJWTSecret(),
		Exp:           time.Hour,
		SigningMethod: jwt.SigningMethodHS256,
	}
}

func ValidBoard() domain.Board {
	name := ValidBoardName()
	description := ValidBoardDescription()
	id := domain.NewBoardID()
	userID := ValidUserID()
	pseudoNow := FixedTimeNow()

	validBoard := domain.Board{
		ID:          id,
		OwnerID:     userID,
		Name:        name,
		Description: description,
		CreatedAt:   pseudoNow,
		UpdatedAt:   pseudoNow,
	}

	return validBoard
}

func UpdateValidBoard(t *testing.T, base *domain.Board, name, description string, updatedAt time.Time) domain.Board {
	t.Helper()
	domainName, err := domain.NewBoardName(name)
	if err != nil {
		t.Fatalf("NewBoardName() error = %v", err)
	}
	domainDescription, err := domain.NewBoardDescription(description)
	if err != nil {
		t.Fatalf("NewBoardDescription() error = %v", err)
	}

	return domain.Board{
		ID:          base.ID,
		OwnerID:     base.OwnerID,
		Name:        domainName,
		Description: domainDescription,
		CreatedAt:   base.CreatedAt,
		UpdatedAt:   updatedAt,
	}
}

func ValidColumn(boardID domain.BoardID) domain.Column {
	name := ValidColumnName()
	position := ValidColumnPosition()
	pseudoNow := FixedTimeNow()

	return domain.Column{
		ID:        domain.NewColumnID(),
		BoardID:   boardID,
		Name:      name,
		Position:  position,
		CreatedAt: pseudoNow,
		UpdatedAt: pseudoNow,
	}
}

func NewValidColumn(t *testing.T, boardID domain.BoardID, name string, position int64) domain.Column {
	t.Helper()

	column := ValidColumn(boardID)

	domainName, err := domain.NewColumnName(name)
	if err != nil {
		t.Fatalf("NewColumnName() error = %v", err)
	}

	domainPosition, err := domain.NewColumnPosition(position)
	if err != nil {
		t.Fatalf("NewColumnPosition() error = %v", err)
	}

	column.Name = domainName
	column.Position = domainPosition

	return column
}

func UpdateValidColumn(t *testing.T, base *domain.Column, name string, updatedAt time.Time) domain.Column {
	t.Helper()

	domainName, err := domain.NewColumnName(name)
	if err != nil {
		t.Fatalf("NewColumnName() error = %v", err)
	}

	return domain.Column{
		ID:        base.ID,
		BoardID:   base.BoardID,
		Name:      domainName,
		Position:  base.Position,
		CreatedAt: base.CreatedAt,
		UpdatedAt: updatedAt,
	}
}

func ValidTask(columnID domain.ColumnID) domain.Task {
	name := ValidTaskName()
	description := ValidTaskDescription()
	position := ValidTaskPosition()
	pseudoNow := FixedTimeNow()

	return domain.Task{
		ID:          domain.NewTaskID(),
		ColumnID:    columnID,
		Name:        name,
		Description: description,
		Position:    position,
		CreatedAt:   pseudoNow,
		UpdatedAt:   pseudoNow,
	}
}

func NewValidTask(t *testing.T, columnID domain.ColumnID, name, description string, position int64) domain.Task {
	t.Helper()

	task := ValidTask(columnID)

	domainName, err := domain.NewTaskName(name)
	if err != nil {
		t.Fatalf("NewTaskName() error = %v", err)
	}

	domainDescription, err := domain.NewTaskDescription(description)
	if err != nil {
		t.Fatalf("NewTaskDescription() error = %v", err)
	}

	domainPosition, err := domain.NewTaskPosition(position)
	if err != nil {
		t.Fatalf("NewTaskPosition() error = %v", err)
	}

	task.Name = domainName
	task.Description = domainDescription
	task.Position = domainPosition

	return task
}

func UpdateValidTask(t *testing.T, base *domain.Task, name, description string, updatedAt time.Time) domain.Task {
	t.Helper()

	domainName, err := domain.NewTaskName(name)
	if err != nil {
		t.Fatalf("NewTaskName() error = %v", err)
	}

	domainDescription, err := domain.NewTaskDescription(description)
	if err != nil {
		t.Fatalf("NewTaskDescription() error = %v", err)
	}

	return domain.Task{
		ID:          base.ID,
		ColumnID:    base.ColumnID,
		Name:        domainName,
		Description: domainDescription,
		Position:    base.Position,
		CreatedAt:   base.CreatedAt,
		UpdatedAt:   updatedAt,
	}
}
