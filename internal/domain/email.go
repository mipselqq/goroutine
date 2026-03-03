package domain

import (
	"net/mail"
	"strings"
)

type Email struct {
	value string
}

func NewEmail(email string) (e Email, errs []string) {
	trimmedEmail := strings.TrimSpace(email)
	lowercasedEmail := strings.ToLower(trimmedEmail)

	_, err := mail.ParseAddress(lowercasedEmail)
	if err != nil {
		errs = append(errs, strings.TrimPrefix(err.Error(), "mail: "))
	}

	return Email{value: lowercasedEmail}, []string{}
}

func (e Email) String() string {
	return e.value
}
