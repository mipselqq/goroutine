package domain

import (
	"database/sql/driver"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	ErrColumnNameTooShort       = "Name is too short"
	ErrColumnNameTooLong        = "Name is too long"
	ErrColumnDescriptionTooLong = "Description is too long"
	ErrColumnPositionValue      = "Position is invalid"
)

type Column struct {
	ID          ColumnID
	BoardID     BoardID
	Name        ColumnName
	Description ColumnDescription
	Position    ColumnPosition
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type (
	columnTag struct{}
	ColumnID  = UUID[columnTag]
)

func NewColumnID() ColumnID {
	return NewID[columnTag]()
}

func ParseColumnID(s string) (ColumnID, error) {
	return ParseID[columnTag](s)
}

func UUIDToColumnID(u uuid.UUID) ColumnID { return ColumnID{value: u} }

type ColumnName struct {
	value string
}

func NewColumnName(name string) (ColumnName, error) {
	trimmedName := strings.TrimSpace(name)
	var issues []string
	if trimmedName == "" {
		issues = append(issues, ErrColumnNameTooShort)
	}
	if len(trimmedName) > 128 {
		issues = append(issues, ErrColumnNameTooLong)
	}
	if len(issues) > 0 {
		return ColumnName{}, &ErrValidation{Issues: issues}
	}

	return ColumnName{value: trimmedName}, nil
}

func (n ColumnName) String() string {
	return n.value
}

type ColumnDescription struct {
	value string
}

func NewColumnDescription(description string) (ColumnDescription, error) {
	trimmedDescription := strings.TrimSpace(description)
	var issues []string
	if len(trimmedDescription) > 1024 {
		issues = append(issues, ErrColumnDescriptionTooLong)
	}

	if len(issues) > 0 {
		return ColumnDescription{}, &ErrValidation{Issues: issues}
	}

	return ColumnDescription{value: trimmedDescription}, nil
}

func (d ColumnDescription) String() string {
	return d.value
}

type ColumnPosition struct {
	value int32
}

func NewColumnPosition(position int64) (ColumnPosition, error) {
	if position <= 0 || position > math.MaxInt32 {
		return ColumnPosition{}, &ErrValidation{Issues: []string{ErrColumnPositionValue}}
	}

	return ColumnPosition{value: int32(position)}, nil
}

func ParseColumnPosition(position string) (ColumnPosition, error) {
	v, err := strconv.ParseInt(position, 10, 64)
	if err != nil {
		return ColumnPosition{}, &ErrValidation{Issues: []string{ErrColumnPositionValue}}
	}

	return NewColumnPosition(v)
}

func (p ColumnPosition) Int64() int64 {
	return int64(p.value)
}

func (p ColumnPosition) Value() (driver.Value, error) {
	return p.value, nil
}
