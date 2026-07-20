package domain

import (
	"database/sql/driver"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	errBoardNameTooShort       string = "Name is too short"
	errBoardNameTooLong        string = "Name is too long"
	errBoardDescriptionTooLong string = "Description is too long"
)

type Board struct {
	ID          BoardID
	OwnerID     UserID
	Name        BoardName
	Description BoardDescription
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type (
	boardTag struct{}
	BoardID  = UUID[boardTag]
)

func NewBoardID() BoardID {
	return newID[boardTag]()
}

func ParseBoardID(s string) (BoardID, error) {
	return parseID[boardTag](s)
}

func NewBoardIDFromUUID(u uuid.UUID) (BoardID, error) {
	return newIDFromUUID[boardTag](u)
}

type BoardName struct {
	value string
}

func NewBoardName(name string) (BoardName, error) {
	trimmedName := strings.TrimSpace(name)
	var issues []string
	if trimmedName == "" {
		issues = append(issues, errBoardNameTooShort)
	}

	if len(trimmedName) > 128 {
		issues = append(issues, errBoardNameTooLong)
	}

	if len(issues) > 0 {
		return BoardName{}, &errValidation{Issues: issues}
	}

	return BoardName{value: trimmedName}, nil
}

func (n BoardName) String() string {
	return n.value
}

func (n BoardName) Value() (driver.Value, error) {
	return n.value, nil
}

type BoardDescription struct {
	value string
}

func NewBoardDescription(description string) (BoardDescription, error) {
	trimmedDescription := strings.TrimSpace(description)
	var issues []string
	if len(trimmedDescription) > 1024 {
		issues = append(issues, errBoardDescriptionTooLong)
	}

	if len(issues) > 0 {
		return BoardDescription{}, &errValidation{Issues: issues}
	}

	return BoardDescription{value: trimmedDescription}, nil
}

func (d BoardDescription) String() string {
	return d.value
}

func (d BoardDescription) Value() (driver.Value, error) {
	return d.value, nil
}
