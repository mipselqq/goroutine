package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"goroutine/internal/secrecy"
)

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

const (
	ErrPasswordTooShort string = "Password is too short"
	ErrInvalidJWTToken  string = "Invalid JWT token"
)

type UserPassword struct {
	value secrecy.SecretString
}

type AuthToken struct {
	value secrecy.SecretString
}

func NewUserPassword(password string) (UserPassword, error) {
	if len(password) < 6 || strings.TrimSpace(password) == "" {
		return UserPassword{}, &ErrValidation{Issues: []string{ErrPasswordTooShort}}
	}

	return UserPassword{value: secrecy.SecretString(password)}, nil
}

func (p UserPassword) RevealSecret() string {
	return p.value.RevealSecret()
}

func (p UserPassword) String() string {
	return p.value.String()
}

func (p UserPassword) LogValue() slog.Value {
	return p.value.LogValue()
}

func (p UserPassword) GoString() string {
	return p.String()
}

func (p UserPassword) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// Domain knows about a little about storage, but this is pragmatic solution
func (p UserPassword) Value() (driver.Value, error) {
	return p.RevealSecret(), nil
}

func (p *UserPassword) Scan(value any) error {
	if value == nil {
		p.value = ""
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected type for Password: %T", value)
	}
	password, err := NewUserPassword(s)
	if err != nil {
		return fmt.Errorf("password: %w: %v", ErrDataCorrupted, err)
	}
	*p = password
	return nil
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

	return AuthToken{value: secrecy.SecretString(trimmed)}, nil
}

func (t AuthToken) RevealSecret() string {
	return t.value.RevealSecret()
}

func (t AuthToken) String() string {
	return t.value.String()
}

func (t AuthToken) LogValue() slog.Value {
	return t.value.LogValue()
}

func (t AuthToken) GoString() string {
	return t.String()
}

func (t AuthToken) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
