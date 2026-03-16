package domain_test

import (
	"errors"
	"testing"

	"goroutine/internal/domain"
)

func TestPassword(t *testing.T) {
	t.Parallel()

	passwordTests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "Valid password",
			input:     "securePass123",
			expectErr: false,
		},
		{
			name:      "Less than 6 characters",
			input:     "12345",
			expectErr: true,
		},
		{
			name:      "Empty password",
			input:     "",
			expectErr: true,
		},
		{
			name:      "Whitespace password",
			input:     "     ",
			expectErr: true,
		},
	}

	for _, tt := range passwordTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.NewPassword(tt.input)
			if tt.expectErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("did not expect error but got: %v", err)
			}
		})
	}
}

func TestPassword_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		wantErr bool
		errIs   error
	}{
		{
			name:    "Valid password",
			input:   "securePass123",
			wantErr: false,
		},
		{
			name:    "Too short password",
			input:   "12345",
			wantErr: true,
			errIs:   domain.ErrDataCorrupted,
		},
		{
			name:    "Empty password",
			input:   "",
			wantErr: true,
			errIs:   domain.ErrDataCorrupted,
		},
		{
			name:    "Null value",
			input:   nil,
			wantErr: false,
		},
		{
			name:    "Invalid type",
			input:   123,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var p domain.Password
			err := p.Scan(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("expected error %v, got %v", tt.errIs, err)
				}
			} else if err != nil {
				t.Errorf("did not expect error, got %v", err)
			}
		})
	}
}
