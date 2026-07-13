package domain

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

const (
	ErrBoardNameTooShort       string = "Name is too short"
	ErrBoardNameTooLong        string = "Name is too long"
	ErrBoardDescriptionTooLong string = "Description is too long"
)

type Board struct {
	ID          BoardID
	OwnerID     UserID
	Name        BoardName
	Description BoardDescription
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type BoardName struct {
	value string
}

func NewBoardName(name string) (BoardName, error) {
	trimmedName := strings.TrimSpace(name)
	var issues []string
	if trimmedName == "" {
		issues = append(issues, ErrBoardNameTooShort)
	}

	if len(trimmedName) > 128 {
		issues = append(issues, ErrBoardNameTooLong)
	}

	if len(issues) > 0 {
		return BoardName{}, &ErrValidation{Issues: issues}
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
		issues = append(issues, ErrBoardDescriptionTooLong)
	}

	if len(issues) > 0 {
		return BoardDescription{}, &ErrValidation{Issues: issues}
	}

	return BoardDescription{value: trimmedDescription}, nil
}

func (d BoardDescription) String() string {
	return d.value
}

func (d BoardDescription) Value() (driver.Value, error) {
	return d.value, nil
}

func (b Board) String() string {
	return fmt.Sprintf(
		"id:          %s\nownerId:     %s\nname:        %q\ndescription: %q\ncreatedAt:   %s\nupdatedAt:   %s",
		b.ID.String(),
		b.OwnerID.String(),
		b.Name.String(),
		b.Description.String(),
		b.CreatedAt.UTC().Format(time.RFC3339Nano),
		b.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
}

type (
	boardTag struct{}
	BoardID  = UUID[boardTag]
)

func NewBoardID() BoardID {
	return NewID[boardTag]()
}

func ParseBoardID(s string) (BoardID, error) {
	return ParseID[boardTag](s)
}
