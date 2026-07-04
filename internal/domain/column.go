package domain

import (
	"database/sql/driver"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
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

func (n ColumnName) IsEmpty() bool {
	return n.value == ""
}

func (n ColumnName) String() string {
	return n.value
}

func (n ColumnName) Value() (driver.Value, error) {
	return n.value, nil
}

func (n *ColumnName) Scan(value any) error {
	if value == nil {
		n.value = ""
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected type for ColumnName: %T", value)
	}
	cn, err := NewColumnName(s)
	if err != nil {
		return fmt.Errorf("column name: %w: %v", ErrDataCorrupted, err)
	}
	*n = cn
	return nil
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

func (d ColumnDescription) IsEmpty() bool {
	return d.value == ""
}

func (d ColumnDescription) Value() (driver.Value, error) {
	return d.value, nil
}

func (d *ColumnDescription) Scan(value any) error {
	if value == nil {
		d.value = ""
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected type for ColumnDescription: %T", value)
	}
	cd, err := NewColumnDescription(s)
	if err != nil {
		return fmt.Errorf("column description: %w: %v", ErrDataCorrupted, err)
	}
	*d = cd
	return nil
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

// TODO: dead code
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

func (p *ColumnPosition) Scan(value any) error {
	if value == nil {
		p.value = 0
		return nil
	}

	v, ok := value.(int64)
	if !ok {
		return fmt.Errorf("unexpected type for ColumnPosition: %T", value)
	}

	position, err := NewColumnPosition(v)
	if err != nil {
		return fmt.Errorf("column position: %w: %v", ErrDataCorrupted, err)
	}

	*p = position
	return nil
}

func (c Column) String() string {
	return fmt.Sprintf(
		"id:          %s\nboardId:     %s\nname:        %q\ndescription: %q\nposition:    %d\ncreatedAt:   %s\nupdatedAt:   %s",
		c.ID.String(),
		c.BoardID.String(),
		c.Name.String(),
		c.Description.String(),
		c.Position.Int64(),
		c.CreatedAt.UTC().Format(time.RFC3339Nano),
		c.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
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
