package template

import (
	"fmt"

	"goroutine/internal/domain"
)

type TaskCreate struct {
	Name string
}

func (t TaskCreate) String() string {
	return fmt.Sprintf("You created task '%s'", t.Name)
}

type TaskRename struct {
	Source string
	Target string
}

func (t TaskRename) String() string {
	return fmt.Sprintf("You renamed task from '%s' to '%s'", t.Source, t.Target)
}

type TaskUpdate struct {
	Name string
}

func (t TaskUpdate) String() string {
	return fmt.Sprintf("You updated task '%s'", t.Name)
}

type TaskMove struct {
	SourceColumnID domain.ColumnID
	TargetColumnID domain.ColumnID
	SourcePosition int64
	TargetPosition int64
}

func (t TaskMove) String() string {
	return fmt.Sprintf(
		"You moved task from column '%s' at position %d to column '%s' at position %d",
		t.SourceColumnID,
		t.SourcePosition,
		t.TargetColumnID,
		t.TargetPosition,
	)
}

type TaskDelete struct {
	Name string
}

func (t TaskDelete) String() string {
	return fmt.Sprintf("You deleted task '%s'", t.Name)
}
