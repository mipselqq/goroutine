package domain

import (
	"strings"
)

const (
	ErrNameTooShort       string = "Name is too short"
	ErrNameTooLong        string = "Name is too long"
	ErrDescriptionTooLong string = "Description is too long"
)

type Name struct {
	value string
}

func NewName(name string) (n Name, errs []string) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		errs = append(errs, ErrNameTooShort)
	}

	if len(trimmedName) > 128 {
		errs = append(errs, ErrNameTooLong)
	}

	if len(errs) > 0 {
		return Name{}, errs
	}

	return Name{value: trimmedName}, []string{}
}

func (n Name) IsEmpty() bool {
	return n.value == ""
}

func (n Name) String() string {
	return n.value
}

type Description struct {
	value string
}

func NewDescription(description string) (d Description, errs []string) {
	trimmedDescription := strings.TrimSpace(description)
	if len(trimmedDescription) > 1024 {
		errs = append(errs, ErrDescriptionTooLong)
	}

	if len(errs) > 0 {
		return Description{}, errs
	}

	return Description{value: trimmedDescription}, []string{}
}

func (d Description) String() string {
	return d.value
}

func (d Description) IsEmpty() bool {
	return d.value == ""
}
