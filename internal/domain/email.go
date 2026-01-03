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
	email = strings.TrimSpace(email)
	if email == "" {
		return Email{}, ErrInvalidEmail
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return Email{}, ErrInvalidEmail
	}

	return Email{value: email}, nil
}

func (e Email) String() string {
	return e.value
}
