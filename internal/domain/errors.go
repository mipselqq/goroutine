package domain

import (
	"errors"
	"fmt"
	"strings"
)

type errValidation struct {
	Issues []string
}

func (e *errValidation) Error() string {
	return fmt.Sprintf("validation error: %s", strings.Join(e.Issues, ", "))
}

func ExtractValidationIssues(err error) []string {
	var ve *errValidation
	if errors.As(err, &ve) {
		return ve.Issues
	}
	return []string{err.Error()}
}
