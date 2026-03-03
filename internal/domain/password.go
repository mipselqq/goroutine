package domain

import (
	"strings"
)

const (
	ErrPasswordTooShort string = "Password is too short"
)

type Password struct {
	value string
}

func NewPassword(password string) (p Password, errs []string) {
	if len(password) < 6 || strings.TrimSpace(password) == "" {
		errs = append(errs, ErrPasswordTooShort)
	}

	if len(errs) > 0 {
		return Password{}, errs
	}

	return Password{value: password}, []string{}
}

func (p Password) String() string {
	return p.value
}
