package domain_test

import (
	"testing"

	"goroutine/internal/domain"
)

func TestEmail(t *testing.T) {
	emailTests := []struct {
		name          string
		input         string
		expectedValue string
		expectErr     bool
	}{
		{
			name:      "Valid email",
			input:     "test@example.com",
			expectErr: false,
		},
		{
			name:      "Invalid email",
			input:     "invalid-email",
			expectErr: true,
		},
		{
			name:      "Empty email",
			input:     "",
			expectErr: true,
		},
		{
			name:      "Whitespace email",
			input:     "   ",
			expectErr: true,
		},
		{
			name:          "Valid email with whitespace",
			input:         "   test@example.com   ",
			expectedValue: "test@example.com",
			expectErr:     false,
		},
		{
			name:          "Uppercase email",
			input:         "TEST@EXAMPLE.COM",
			expectedValue: "test@example.com",
			expectErr:     false,
		},
		{
			name:          "Mixed case email",
			input:         "TeSt@ExAmpLe.CoM",
			expectedValue: "test@example.com",
			expectErr:     false,
		},
	}

	for _, tt := range emailTests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := domain.NewEmail(tt.input)

			if !tt.expectErr {
				if tt.expectedValue != "" && email.String() != tt.expectedValue {
					t.Errorf("expected email %s, got %s", tt.expectedValue, email.String())
				}
				if err != nil {
					t.Errorf("did not expect error but got: %v", err)
				}
			} else if err == nil {
				t.Errorf("expected error but got none")
			}
		})
	}
}
