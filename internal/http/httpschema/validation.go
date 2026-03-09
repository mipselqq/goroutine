package httpschema

func ValidateField[T any](field, val string, constructor func(string) (T, []string), details *[]Detail) T {
	res, errs := constructor(val)
	if len(errs) > 0 {
		*details = append(*details, Detail{Field: field, Issues: errs})
	}
	return res
}
