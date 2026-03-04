package domain_test

import (
	"testing"

	"goroutine/internal/domain"
)

func TestEmail(t *testing.T) {
	t.Parallel()

	emailTests := []struct {
		name          string
		input         string
		expectedValue string
		expectErrs    bool
	}{
		{
			name:       "Valid email",
			input:      "test@example.com",
			expectErrs: false,
		},
		{
			name:       "Invalid email",
			input:      "invalid-email",
			expectErrs: true,
		},
		{
			name:       "Empty email",
			input:      "",
			expectErrs: true,
		},
		{
			name:       "Whitespace email",
			input:      "   ",
			expectErrs: true,
		},
		{
			name:          "Valid email with whitespace",
			input:         "   test@example.com   ",
			expectedValue: "test@example.com",
			expectErrs:    false,
		},
		{
			name:          "Uppercase email",
			input:         "TEST@EXAMPLE.COM",
			expectedValue: "test@example.com",
			expectErrs:    false,
		},
		{
			name:          "Mixed case email",
			input:         "TeSt@ExAmpLe.CoM",
			expectedValue: "test@example.com",
			expectErrs:    false,
		},
	}

	for _, tt := range emailTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			email, errs := domain.NewEmail(tt.input)

			if !tt.expectErrs {
				if tt.expectedValue != "" && email.String() != tt.expectedValue {
					t.Errorf("expected email %q, got %q", tt.expectedValue, email.String())
				}
				if len(errs) > 0 {
					t.Errorf("did not expect errors but got: %v", errs)
				}
			} else if len(errs) == 0 {
				t.Errorf("expected errors but got none")
			}
		})
	}
}
