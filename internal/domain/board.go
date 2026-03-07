package domain

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const (
	ErrNameTooShort       string = "Name is too short"
	ErrNameTooLong        string = "Name is too long"
	ErrDescriptionTooLong string = "Description is too long"
)

type Board struct {
	ID          BoardID
	Name        Name
	Description Description
	CreatedAt   string
	UpdatedAt   string
}

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

type BoardID struct {
	value uuid.UUID
}

func NewBoardID() BoardID {
	id, _ := uuid.NewV7()
	return BoardID{value: id}
}

func ParseBoardID(s string) (BoardID, error) {
	b, err := uuid.Parse(s)
	if err != nil {
		return BoardID{}, fmt.Errorf("parse board id: %w", err)
	}
	return BoardID{value: b}, nil
}

func (b BoardID) String() string {
	return b.value.String()
}

func (b BoardID) IsEmpty() bool {
	return b.value == uuid.Nil
}

func (b BoardID) UUID() uuid.UUID {
	return b.value
}

// Scan implements the sql.Scanner interface.
func (b *BoardID) Scan(src any) error {
	if src == nil {
		b.value = uuid.Nil
		return nil
	}

	switch v := src.(type) {
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return err
		}
		b.value = parsed
	case []byte:
		const uuidByteLen = 16
		isUUIDBytes := len(v) == uuidByteLen

		if isUUIDBytes {
			parsed, err := uuid.FromBytes(v)
			if err != nil {
				return err
			}
			b.value = parsed
		} else { // Byte representation of string
			parsed, err := uuid.ParseBytes(v)
			if err != nil {
				return err
			}
			b.value = parsed
		}
	default:
		return fmt.Errorf("unexpected type for UserID: %T", src)
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (b BoardID) Value() (driver.Value, error) {
	if b.IsEmpty() {
		return nil, nil
	}
	return b.value.String(), nil
}
