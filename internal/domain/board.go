package domain

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

const (
	ErrNameTooShort       string = "Name is too short"
	ErrNameTooLong        string = "Name is too long"
	ErrDescriptionTooLong string = "Description is too long"
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
		issues = append(issues, ErrNameTooShort)
	}

	if len(trimmedName) > 128 {
		issues = append(issues, ErrNameTooLong)
	}

	if len(issues) > 0 {
		return BoardName{}, &ValidationError{Issues: issues}
	}

	return BoardName{value: trimmedName}, nil
}

func (n BoardName) IsEmpty() bool {
	return n.value == ""
}

func (n BoardName) String() string {
	return n.value
}

func (n BoardName) Value() (driver.Value, error) {
	return n.value, nil
}

func (n *BoardName) Scan(value any) error {
	if value == nil {
		n.value = ""
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected type for BoardName: %T", value)
	}
	bn, err := NewBoardName(s)
	if err != nil {
		return fmt.Errorf("board name: %w: %v", ErrDataCorrupted, err)
	}
	*n = bn
	return nil
}

type BoardDescription struct {
	value string
}

func NewBoardDescription(description string) (BoardDescription, error) {
	trimmedDescription := strings.TrimSpace(description)
	var issues []string
	if len(trimmedDescription) > 1024 {
		issues = append(issues, ErrDescriptionTooLong)
	}

	if len(issues) > 0 {
		return BoardDescription{}, &ValidationError{Issues: issues}
	}

	return BoardDescription{value: trimmedDescription}, nil
}

func (d BoardDescription) String() string {
	return d.value
}

func (d BoardDescription) IsEmpty() bool {
	return d.value == ""
}

func (d BoardDescription) Value() (driver.Value, error) {
	return d.value, nil
}

func (d *BoardDescription) Scan(value any) error {
	if value == nil {
		d.value = ""
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected type for BoardDescription: %T", value)
	}
	bd, err := NewBoardDescription(s)
	if err != nil {
		return fmt.Errorf("board description: %w: %v", ErrDataCorrupted, err)
	}
	*d = bd
	return nil
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
