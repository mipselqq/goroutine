package httpschema

import (
	"errors"

	"goroutine/internal/domain"
)

func ValidateField[T any, V any](field string, val V, constructor func(V) (T, error), details *[]Detail) T {
	res, err := constructor(val)
	if err != nil {
		var ve *domain.ErrValidation
		if errors.As(err, &ve) {
			*details = append(*details, Detail{Field: field, Issues: ve.Issues})
		} else {
			*details = append(*details, Detail{Field: field, Issues: []string{err.Error()}})
		}
	}
	return res
}
