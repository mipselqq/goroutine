package domain

import (
	"errors"
	"fmt"
	"strings"
)

var ErrDataCorrupted = errors.New("invalid data appeared in the database")

type ErrValidation struct {
	Issues []string
}

func (e *ErrValidation) Error() string {
	return fmt.Sprintf("validation error: %s", strings.Join(e.Issues, ", "))
}

func NewValidationError(issues ...string) error {
	if len(issues) == 0 {
		return nil
	}
	return &ErrValidation{Issues: issues}
}

func ExtractValidationIssues(err error) []string {
	var ve *ErrValidation
	if errors.As(err, &ve) {
		return ve.Issues
	}
	return []string{err.Error()}
}
