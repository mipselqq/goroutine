package domain_test

import (
	"errors"
	"testing"

	"goroutine/internal/domain"
)

func TestEmail(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			email, err := domain.NewEmail(tt.input)

			if !tt.expectErr {
				if tt.expectedValue != "" && email.String() != tt.expectedValue {
					t.Errorf("expected email %q, got %q", tt.expectedValue, email.String())
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

func TestEmail_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     any
		expectErr bool
		errKind   error
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
			errKind:   domain.ErrDataCorrupted,
		},
		{
			name:      "Null value",
			input:     nil,
			expectErr: false,
		},
		{
			name:      "Invalid type",
			input:     123,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var e domain.Email
			err := e.Scan(tt.input)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errKind != nil && !errors.Is(err, tt.errKind) {
					t.Errorf("expected error %v, got %v", tt.errKind, err)
				}
			} else if err != nil {
				t.Errorf("did not expect error, got %v", err)
			}
		})
	}
}
