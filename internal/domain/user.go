// Package domain provides DDD-style value objects with validation.
//
// Value object rules: no dependencies, no side effects, if an object exists, then its value is valid.
package domain

import (
	"slices"
	"strings"

	"goroutine/internal/secrecy"

	"github.com/google/uuid"
)

const (
	errPasswordTooShort string = "Password is too short"
	errInvalidJWTToken  string = "Invalid JWT token"
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
	return newID[userID]()
}

func ParseUserID(s string) (UserID, error) {
	return parseID[userID](s)
}

func NewUserIDFromUUID(u uuid.UUID) (UserID, error) {
	return newIDFromUUID[userID](u)
}

type PasswordHash struct {
	secrecy.SecretString
}

func NewPasswordHash(hash string) PasswordHash {
	return PasswordHash{SecretString: secrecy.SecretString(hash)}
}

type UserPassword struct {
	secrecy.SecretString
}

func NewUserPassword(password string) (UserPassword, error) {
	if len(password) < 6 || strings.TrimSpace(password) == "" {
		return UserPassword{}, &errValidation{Issues: []string{errPasswordTooShort}}
	}

	return UserPassword{SecretString: secrecy.SecretString(password)}, nil
}

type AuthToken struct {
	secrecy.SecretString
}

func NewJWTString(token string) (AuthToken, error) {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return AuthToken{}, &errValidation{Issues: []string{errInvalidJWTToken}}
	}

	parts := strings.Split(trimmed, ".")
	if len(parts) != 3 {
		return AuthToken{}, &errValidation{Issues: []string{errInvalidJWTToken}}
	}

	if slices.Contains(parts, "") {
		return AuthToken{}, &errValidation{Issues: []string{errInvalidJWTToken}}
	}

	return AuthToken{SecretString: secrecy.SecretString(trimmed)}, nil
}
