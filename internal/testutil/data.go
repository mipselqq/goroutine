// Package testutil contains fixtures, helpers, and a test logger. Used in all the layers.
package testutil

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/secrecy"
	"goroutine/internal/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func MustLoadDevEnv() {
	err := godotenv.Load("../../.env.dev")
	if err == nil {
		return
	}
	err = godotenv.Load("../.env.dev")
	if err == nil {
		return
	}
	panic("Failed to load dev env from file (tried ../../.env.dev and ../.env.dev)")
}

func FixedNow() time.Time       { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }
func Fixed5mFromNow() time.Time { return FixedNow().Add(5 * time.Minute) }

const timeFormat = "2006-01-02T15:04:05.000Z07:00"

func FixedTime5mFromNowStr() string {
	return Fixed5mFromNow().UTC().Format(timeFormat)
}

func FixedNowStr() string { return FixedNow().UTC().Format(timeFormat) }

func Big25KBJSON() json.RawMessage {
	return json.RawMessage(`{"a":"` + strings.Repeat("b", 25*1024) + `"}`)
}

func Must[A any, T any](fn func(A) (T, error), arg A) T {
	v, err := fn(arg)
	if err != nil {
		panic(err)
	}
	return v
}

const validUUIDv7 = "018e1000-0000-7000-8000-000000000000"

func ValidUserID() domain.UserID {
	return Must(domain.ParseUserID, validUUIDv7)
}

func ValidEmail() domain.Email {
	return Must(domain.NewEmail, "test@example.com")
}

func ValidPassword() domain.UserPassword {
	return Must(domain.NewUserPassword, "qwerty")
}

func ValidPasswordHash() string {
	return "$argon2id$v=19$m=65536,t=1,p=16$kUYJyX3h53cARKnKqFZxvQ$IXz2KOKbyVklgyVmz9ebJ1ffOgmcyMpn/GTUWsep5lk"
}

func AnotherValidPasswordHash() string {
	return "$argon2id$v=19$m=65536,t=3,p=4$bm90LXF3ZXJ0eQ$fSowp1Rof0fXhF+rXv2f6w"
}

func ValidBoardName() domain.BoardName {
	return Must(domain.NewBoardName, "Test Board")
}

func ValidBoardDescription() domain.BoardDescription {
	return Must(domain.NewBoardDescription, "Test Board Description")
}

func ValidColumnName() domain.ColumnName {
	return Must(domain.NewColumnName, "To Do")
}

func ValidColumnPosition() domain.ColumnPosition {
	return Must(domain.NewColumnPosition, 1)
}

func ValidColumnDescription() domain.ColumnDescription {
	return Must(domain.NewColumnDescription, "Test Column Description")
}

func ValidTaskName() domain.TaskName {
	return Must(domain.NewTaskName, "Write tests")
}

func ValidTaskDescription() domain.TaskDescription {
	return Must(domain.NewTaskDescription, "Cover the new endpoint with tests")
}

func ValidTaskPosition() domain.TaskPosition {
	return Must(domain.NewTaskPosition, 1)
}

func NewValidTaskPosition(t *testing.T, n int64) domain.TaskPosition {
	t.Helper()
	return Must(domain.NewTaskPosition, n)
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
	pseudoNow := FixedNow()

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
	domainName := Must(domain.NewBoardName, name)
	domainDescription := Must(domain.NewBoardDescription, description)

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
	description := ValidColumnDescription()
	position := ValidColumnPosition()
	pseudoNow := FixedNow()

	return domain.Column{
		ID:          domain.NewColumnID(),
		BoardID:     boardID,
		Name:        name,
		Description: description,
		Position:    position,
		CreatedAt:   pseudoNow,
		UpdatedAt:   pseudoNow,
	}
}

func NewValidColumn(t *testing.T, boardID domain.BoardID, name string, position int64) domain.Column {
	t.Helper()

	column := ValidColumn(boardID)

	domainName := Must(domain.NewColumnName, name)
	domainPosition := Must(domain.NewColumnPosition, position)

	column.Name = domainName
	column.Position = domainPosition

	return column
}

func UpdateValidColumn(t *testing.T, base *domain.Column, name, description string, updatedAt time.Time) domain.Column {
	t.Helper()

	domainName := Must(domain.NewColumnName, name)
	domainDescription := Must(domain.NewColumnDescription, description)

	return domain.Column{
		ID:          base.ID,
		BoardID:     base.BoardID,
		Name:        domainName,
		Description: domainDescription,
		Position:    base.Position,
		CreatedAt:   base.CreatedAt,
		UpdatedAt:   updatedAt,
	}
}

func ValidTask(columnID domain.ColumnID) domain.Task {
	name := ValidTaskName()
	description := ValidTaskDescription()
	position := ValidTaskPosition()
	pseudoNow := FixedNow()

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

	domainName := Must(domain.NewTaskName, name)
	domainDescription := Must(domain.NewTaskDescription, description)
	domainPosition := Must(domain.NewTaskPosition, position)

	task.Name = domainName
	task.Description = domainDescription
	task.Position = domainPosition

	return task
}

func UpdateValidTask(t *testing.T, base *domain.Task, name, description string, updatedAt time.Time) domain.Task {
	t.Helper()

	domainName := Must(domain.NewTaskName, name)
	domainDescription := Must(domain.NewTaskDescription, description)

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

func ValidTelegramToken() domain.TelegramToken {
	return Must(domain.NewTelegramToken, "8927121804:MOCKhk1QdJpRJdISscC0COr19kH79_4f9vw")
}

func AnotherValidTelegramToken() domain.TelegramToken {
	return Must(domain.NewTelegramToken, "8927121804:MOCKtw0QdJpRJdISscC0COr19kH79_4f9vw")
}

func ValidTelegramLinkToken() domain.TelegramLinkToken {
	return Must(domain.NewTelegramLinkToken, validUUIDv7)
}

func ValidTelegramChatID() domain.TelegramChatID {
	return Must(domain.NewTelegramChatID, int64(123456789))
}

func ValidTelegramUsername() domain.TelegramUsername {
	return Must(domain.NewTelegramUsername, "@testuser")
}

func ValidTelegramMessage() domain.TelegramMessage {
	return Must(domain.NewTelegramMessage, "Hello, world!")
}
