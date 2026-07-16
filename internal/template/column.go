package template

import "fmt"

type ColumnCreate struct {
	Name string
}

func (t ColumnCreate) String() string {
	return fmt.Sprintf("You created column '%s'", t.Name)
}

type ColumnRename struct {
	Source string
	Target string
}

func (t ColumnRename) String() string {
	return fmt.Sprintf("You renamed column from '%s' to '%s'", t.Source, t.Target)
}

type ColumnUpdate struct {
	Name string
}

func (t ColumnUpdate) String() string {
	return fmt.Sprintf("You updated column '%s'", t.Name)
}

type ColumnMove struct {
	SourcePosition int64
	TargetPosition int64
}

func (t ColumnMove) String() string {
	return fmt.Sprintf("You moved column from position %d to position %d", t.SourcePosition, t.TargetPosition)
}

type ColumnDelete struct {
	ID string
}

func (t ColumnDelete) String() string {
	return fmt.Sprintf("You deleted column '%s'", t.ID)
}
