package domain

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

const (
	ErrPasswordTooShort string = "Password is too short"
)

type Password struct {
	value string
}

func NewPassword(password string) (Password, error) {
	if len(password) < 6 || strings.TrimSpace(password) == "" {
		return Password{}, &ValidationError{Issues: []string{ErrPasswordTooShort}}
	}

	return Password{value: password}, nil
}

func (p Password) String() string {
	return p.value
}

// Domain knows about a little about storage, but this is pragmatic solution
func (p Password) Value() (driver.Value, error) {
	return p.value, nil
}

func (p *Password) Scan(value any) error {
	if value == nil {
		p.value = ""
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected type for Password: %T", value)
	}
	password, err := NewPassword(s)
	if err != nil {
		return fmt.Errorf("password: %w: %v", ErrDataCorrupted, err)
	}
	*p = password
	return nil
}
