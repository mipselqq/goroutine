package domain

import (
	"errors"
	"strings"
)

var ErrInvalidTitle = errors.New("invalid title")

type Title struct {
	value string
}

func NewTitle(title string) (Title, error) {
	if strings.TrimSpace(title) == "" {
		return Title{}, ErrInvalidTitle
	}

	return Title{value: title}, nil
}

func (t Title) String() string {
	return t.value
}
