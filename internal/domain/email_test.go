package domain_test

import (
	"testing"

	"go-todo/internal/domain"
)

func TestEmail(t *testing.T) {
	emailTests := []struct {
		name      string
		input     string
		expectErr bool
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
	}

	for _, tt := range emailTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewEmail(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
		})
	}
}
