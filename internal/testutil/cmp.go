package testutil

import (
	"github.com/google/go-cmp/cmp"

	"goroutine/internal/domain"
)

func DomainCmpOpts() cmp.Option {
	return cmp.AllowUnexported(
		domain.BoardID{},
		domain.BoardName{},
		domain.BoardDescription{},
		domain.UserID{},
		domain.ColumnID{},
		domain.ColumnName{},
		domain.ColumnPosition{},
	)
}
