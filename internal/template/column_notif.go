package template

import (
	"fmt"

	"goroutine/internal/domain"
)

type ColumnCreateNotif struct {
	Name domain.ColumnName
}

func (t ColumnCreateNotif) String() string {
	return fmt.Sprintf("You created column '%s'", t.Name)
}

type ColumnRenameNotif struct {
	Source domain.ColumnName
	Target domain.ColumnName
}

func (t ColumnRenameNotif) String() string {
	return fmt.Sprintf("You renamed column from '%s' to '%s'", t.Source, t.Target)
}

type ColumnDescriptionUpdateNotif struct {
	Name domain.ColumnName
}

func (t ColumnDescriptionUpdateNotif) String() string {
	return fmt.Sprintf("You updated description of column '%s'", t.Name)
}

type ColumnMoveNotif struct {
	SourcePosition domain.ColumnPosition
	TargetPosition domain.ColumnPosition
}

func (t ColumnMoveNotif) String() string {
	return fmt.Sprintf("You moved column from position %d to position %d", t.SourcePosition.Int64(), t.TargetPosition.Int64())
}

type ColumnDeleteNotif struct {
	ID domain.ColumnID
}

func (t ColumnDeleteNotif) String() string {
	return fmt.Sprintf("You deleted column '%s'", t.ID)
}
