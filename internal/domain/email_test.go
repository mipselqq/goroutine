package domain_test

import (
	"errors"
	"testing"

	"goroutine/internal/domain"
)

func TestEmail(t *testing.T) {
	t.Parallel()

	emailTests := []struct {
		name      string
		input     string
		wantValue string
		wantErr   bool
	}{
		{
			name:    "Valid email",
			input:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "Invalid email",
			input:   "invalid-email",
			wantErr: true,
		},
		{
			name:    "Empty email",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Whitespace email",
			input:   "   ",
			wantErr: true,
		},
		{
			name:      "Valid email with whitespace",
			input:     "   test@example.com   ",
			wantValue: "test@example.com",
			wantErr:   false,
		},
		{
			name:      "Uppercase email",
			input:     "TEST@EXAMPLE.COM",
			wantValue: "test@example.com",
			wantErr:   false,
		},
		{
			name:      "Mixed case email",
			input:     "TeSt@ExAmpLe.CoM",
			wantValue: "test@example.com",
			wantErr:   false,
		},
	}

	for _, tt := range emailTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			email, err := domain.NewEmail(tt.input)

			if !tt.wantErr {
				if tt.wantValue != "" && email.String() != tt.wantValue {
					t.Errorf("got email %q, want %q", email, tt.wantValue)
				}
				if err != nil {
					t.Errorf("got error %v, want nil", err)
				}
			} else if err == nil {
				t.Error("got nil error, want non-nil")
			}
		})
	}
}

func TestEmail_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		wantErr bool
		errKind error
	}{
		{
			name:    "Valid email",
			input:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "Invalid email",
			input:   "invalid-email",
			wantErr: true,
			errKind: domain.ErrDataCorrupted,
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

			var e domain.Email
			err := e.Scan(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("got nil error, want non-nil")
				} else if tt.errKind != nil && !errors.Is(err, tt.errKind) {
					t.Errorf("got error %v, want %v", err, tt.errKind)
				}
			} else if err != nil {
				t.Errorf("got error %v, want nil", err)
			}
		})
	}
}
