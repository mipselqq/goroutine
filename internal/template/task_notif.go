package template

import (
	"fmt"

	"goroutine/internal/domain"
)

type TaskCreateNotif struct {
	Name domain.TaskName
}

func (t TaskCreateNotif) String() string {
	return fmt.Sprintf("You created task '%s'", t.Name)
}

type TaskRenameNotif struct {
	Source domain.TaskName
	Target domain.TaskName
}

func (t TaskRenameNotif) String() string {
	return fmt.Sprintf("You renamed task from '%s' to '%s'", t.Source, t.Target)
}

type TaskDescriptionUpdateNotif struct {
	Name domain.TaskName
}

func (t TaskDescriptionUpdateNotif) String() string {
	return fmt.Sprintf("You updated description of task '%s'", t.Name)
}

type TaskMoveNotif struct {
	SourceColumnID domain.ColumnID
	TargetColumnID domain.ColumnID
	SourcePosition domain.TaskPosition
	TargetPosition domain.TaskPosition
}

func (t TaskMoveNotif) String() string {
	return fmt.Sprintf(
		"You moved task from column '%s' at position %d to column '%s' at position %d",
		t.SourceColumnID,
		t.SourcePosition.Int64(),
		t.TargetColumnID,
		t.TargetPosition.Int64(),
	)
}

type TaskDeleteNotif struct {
	Name domain.TaskName
}

func (t TaskDeleteNotif) String() string {
	return fmt.Sprintf("You deleted task '%s'", t.Name)
}
