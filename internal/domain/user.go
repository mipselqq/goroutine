// Package domain provides DDD-style value objects with validation.
//
// Value object rules: no dependencies, no side effects, if an object exists, then its value is valid.
package domain

import (
	"slices"
	"strings"

	"goroutine/internal/secrecy"
)

type User struct {
	ID               UserID
	Email            Email
	PasswordHash     PasswordHash
	TelegramChatID   TelegramChatID
	TelegramUsername TelegramUsername
}

type (
	userID struct{}
	UserID = UUID[userID]
)

func NewUserID() UserID {
	return NewID[userID]()
}

func ParseUserID(s string) (UserID, error) {
	return ParseID[userID](s)
}

type PasswordHash struct {
	secrecy.SecretString
}

const (
	ErrPasswordTooShort string = "Password is too short"
	ErrInvalidJWTToken  string = "Invalid JWT token"
)

type UserPassword struct {
	secrecy.SecretString
}

func NewUserPassword(password string) (UserPassword, error) {
	if len(password) < 6 || strings.TrimSpace(password) == "" {
		return UserPassword{}, &ErrValidation{Issues: []string{ErrPasswordTooShort}}
	}

	return UserPassword{SecretString: secrecy.SecretString(password)}, nil
}

type AuthToken struct {
	secrecy.SecretString
}

func NewJWTString(token string) (AuthToken, error) {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return AuthToken{}, &ErrValidation{Issues: []string{ErrInvalidJWTToken}}
	}

	parts := strings.Split(trimmed, ".")
	if len(parts) != 3 {
		return AuthToken{}, &ErrValidation{Issues: []string{ErrInvalidJWTToken}}
	}

	if slices.Contains(parts, "") {
		return AuthToken{}, &ErrValidation{Issues: []string{ErrInvalidJWTToken}}
	}

	return AuthToken{SecretString: secrecy.SecretString(trimmed)}, nil
}
