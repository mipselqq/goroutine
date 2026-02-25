package domain

import (
	"errors"
	"net/mail"
	"strings"
)

var ErrInvalidEmail = errors.New("invalid email")

type Email struct {
	value string
}

func NewEmail(email string) (Email, error) {
	trimmedEmail := strings.TrimSpace(email)

	if email == "" {
		return Email{}, ErrInvalidEmail
	}

	normalizedEmail := strings.ToLower(trimmedEmail)

	_, err := mail.ParseAddress(normalizedEmail)
	if err != nil {
		return Email{}, ErrInvalidEmail
	}

	return Email{value: normalizedEmail}, nil
}

func (e Email) String() string {
	return e.value
}
