package domain

import (
	"database/sql/driver"
	"fmt"
	"net/mail"
	"strings"
)

type Email struct {
	value string
}

const ErrInvalidEmail = "Invalid email"

func NewEmail(email string) (Email, error) {
	trimmedEmail := strings.TrimSpace(email)
	lowercasedEmail := strings.ToLower(trimmedEmail)

	_, err := mail.ParseAddress(lowercasedEmail)
	if err != nil {
		return Email{}, &ValidationError{Issues: []string{ErrInvalidEmail}}
	}

	return Email{value: lowercasedEmail}, nil
}

func (e Email) String() string {
	return e.value
}

func (e Email) Value() (driver.Value, error) {
	return e.value, nil
}

func (e *Email) Scan(value any) error {
	if value == nil {
		e.value = ""
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected type for Email: %T", value)
	}
	email, err := NewEmail(s)
	if err != nil {
		return fmt.Errorf("email: %w: %v", ErrDataCorrupted, err)
	}
	*e = email
	return nil
}
