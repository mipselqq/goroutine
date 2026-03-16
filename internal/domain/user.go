package domain

import (
	"database/sql/driver"
	"fmt"
	"strings"
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
)

type UserPassword struct {
	value string
}

func NewUserPassword(password string) (UserPassword, error) {
	if len(password) < 6 || strings.TrimSpace(password) == "" {
		return UserPassword{}, &ErrValidation{Issues: []string{ErrPasswordTooShort}}
	}

	return UserPassword{value: password}, nil
}

func (p UserPassword) String() string {
	return p.value
}

// Domain knows about a little about storage, but this is pragmatic solution
func (p UserPassword) Value() (driver.Value, error) {
	return p.value, nil
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
