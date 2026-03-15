package httpschema

import (
	"errors"

	"goroutine/internal/domain"
)

func ValidateField[T any](field, val string, constructor func(string) (T, error), details *[]Detail) T {
	res, err := constructor(val)
	if err != nil {
		var ve *domain.ValidationError
		if errors.As(err, &ve) {
			*details = append(*details, Detail{Field: field, Issues: ve.Issues})
		} else {
			*details = append(*details, Detail{Field: field, Issues: []string{err.Error()}})
		}
	}
	return res
}
