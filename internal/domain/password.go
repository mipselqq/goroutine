package domain

import (
	"errors"
	"strings"
)

var ErrInvalidPassword = errors.New("invalid password")

type Password struct {
	value string
}

func NewPassword(password string) (Password, error) {
	if strings.TrimSpace(password) == "" {
		return Password{}, ErrInvalidPassword
	}

	if len(password) < 6 {
		return Password{}, errors.New("password must be at least 6 characters")
	}

	return Password{value: password}, nil
}

func (p Password) String() string {
	return p.value
}
