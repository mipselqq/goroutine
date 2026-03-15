package domain

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Issues []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", strings.Join(e.Issues, ", "))
}

func NewValidationError(issues ...string) error {
	if len(issues) == 0 {
		return nil
	}
	return &ValidationError{Issues: issues}
}
