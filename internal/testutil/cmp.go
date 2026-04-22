package testutil

import (
	"github.com/google/go-cmp/cmp"

	"goroutine/internal/domain"
)

func CmpAllowUnexported() cmp.Option {
	return cmp.AllowUnexported(
		domain.BoardID{},
		domain.BoardName{},
		domain.BoardDescription{},
		domain.UserID{},
		domain.ColumnID{},
		domain.ColumnName{},
		domain.ColumnPosition{},
		domain.TaskID{},
		domain.TaskName{},
		domain.TaskDescription{},
		domain.TaskPosition{},
	)
}
