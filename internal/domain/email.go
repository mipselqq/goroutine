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
	lowercasedEmail := strings.ToLower(trimmedEmail)

	_, err := mail.ParseAddress(lowercasedEmail)
	if err != nil {
		return Email{}, ErrInvalidEmail
	}

	return Email{value: lowercasedEmail}, nil
}

func (e Email) String() string {
	return e.value
}
