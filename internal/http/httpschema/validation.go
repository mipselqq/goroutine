package httpschema

import "goroutine/internal/domain"

func ValidateField[T any, V any](field string, val V, constructor func(V) (T, error), details *[]Detail) T {
	res, err := constructor(val)
	if err != nil {
		*details = append(*details, Detail{Field: field, Issues: domain.ExtractValidationIssues(err)})
	}
	return res
}
