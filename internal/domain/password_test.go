package domain_test

import (
	"testing"

	"goroutine/internal/domain"
)

func TestPassword(t *testing.T) {
	t.Parallel()

	passwordTests := []struct {
		name       string
		input      string
		expectErrs bool
	}{
		{
			name:       "Valid password",
			input:      "securePass123",
			expectErrs: false,
		},
		{
			name:       "Less than 6 characters",
			input:      "12345",
			expectErrs: true,
		},
		{
			name:       "Empty password",
			input:      "",
			expectErrs: true,
		},
		{
			name:       "Whitespace password",
			input:      "     ",
			expectErrs: true,
		},
	}

	for _, tt := range passwordTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, errs := domain.NewPassword(tt.input)
			if tt.expectErrs && len(errs) == 0 {
				t.Errorf("expected errors but got none")
			}
			if !tt.expectErrs && len(errs) > 0 {
				t.Errorf("did not expect errors but got: %v", errs)
			}
		})
	}
}
