package domain

import (
	"database/sql/driver"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	ErrTaskNameTooShort       = "Name is too short"
	ErrTaskNameTooLong        = "Name is too long"
	ErrTaskDescriptionTooLong = "Description is too long"
	ErrTaskPositionValue      = "Position is invalid"
)

type Task struct {
	ID          TaskID
	ColumnID    ColumnID
	Name        TaskName
	Description TaskDescription
	Position    TaskPosition
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TaskName struct {
	value string
}

func NewTaskName(name string) (TaskName, error) {
	trimmedName := strings.TrimSpace(name)
	var issues []string
	if trimmedName == "" {
		issues = append(issues, ErrTaskNameTooShort)
	}
	if len(trimmedName) > 128 {
		issues = append(issues, ErrTaskNameTooLong)
	}
	if len(issues) > 0 {
		return TaskName{}, &ErrValidation{Issues: issues}
	}

	return TaskName{value: trimmedName}, nil
}

func (n TaskName) IsEmpty() bool {
	return n.value == ""
}

func (n TaskName) String() string {
	return n.value
}

func (n TaskName) Value() (driver.Value, error) {
	return n.value, nil
}

func (n *TaskName) Scan(value any) error {
	if value == nil {
		n.value = ""
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected type for TaskName: %T", value)
	}
	tn, err := NewTaskName(s)
	if err != nil {
		return fmt.Errorf("task name: %w: %v", ErrDataCorrupted, err)
	}
	*n = tn
	return nil
}

type TaskDescription struct {
	value string
}

func NewTaskDescription(description string) (TaskDescription, error) {
	trimmedDescription := strings.TrimSpace(description)
	var issues []string
	if len(trimmedDescription) > 1024 {
		issues = append(issues, ErrTaskDescriptionTooLong)
	}

	if len(issues) > 0 {
		return TaskDescription{}, &ErrValidation{Issues: issues}
	}

	return TaskDescription{value: trimmedDescription}, nil
}

func (d TaskDescription) IsEmpty() bool {
	return d.value == ""
}

func (d TaskDescription) String() string {
	return d.value
}

func (d TaskDescription) Value() (driver.Value, error) {
	return d.value, nil
}

func (d *TaskDescription) Scan(value any) error {
	if value == nil {
		d.value = ""
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected type for TaskDescription: %T", value)
	}
	td, err := NewTaskDescription(s)
	if err != nil {
		return fmt.Errorf("task description: %w: %v", ErrDataCorrupted, err)
	}
	*d = td
	return nil
}

type TaskPosition struct {
	value int32
}

func NewTaskPosition(position int64) (TaskPosition, error) {
	if position <= 0 || position > math.MaxInt32 {
		return TaskPosition{}, &ErrValidation{Issues: []string{ErrTaskPositionValue}}
	}

	return TaskPosition{value: int32(position)}, nil
}

func (p TaskPosition) Int64() int64 {
	return int64(p.value)
}

func (p TaskPosition) Value() (driver.Value, error) {
	return p.value, nil
}

func (p *TaskPosition) Scan(value any) error {
	if value == nil {
		p.value = 0
		return nil
	}

	v, ok := value.(int64)
	if !ok {
		return fmt.Errorf("unexpected type for TaskPosition: %T", value)
	}

	position, err := NewTaskPosition(v)
	if err != nil {
		return fmt.Errorf("task position: %w: %v", ErrDataCorrupted, err)
	}

	*p = position
	return nil
}

func (t Task) String() string {
	return fmt.Sprintf(
		"id:          %s\ncolumnId:    %s\nname:        %q\ndescription: %q\nposition:    %d\ncreatedAt:   %s\nupdatedAt:   %s",
		t.ID.String(),
		t.ColumnID.String(),
		t.Name.String(),
		t.Description.String(),
		t.Position.Int64(),
		t.CreatedAt.UTC().Format(time.RFC3339Nano),
		t.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
}

type (
	taskTag struct{}
	TaskID  = UUID[taskTag]
)

func NewTaskID() TaskID {
	return NewID[taskTag]()
}

func ParseTaskID(s string) (TaskID, error) {
	return ParseID[taskTag](s)
}
