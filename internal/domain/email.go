package domain

import (
	"database/sql/driver"
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
		return Email{}, &ErrValidation{Issues: []string{ErrInvalidEmail}}
	}

	return Email{value: lowercasedEmail}, nil
}

func (e Email) String() string {
	return e.value
}

func (e Email) Value() (driver.Value, error) {
	return e.value, nil
}
