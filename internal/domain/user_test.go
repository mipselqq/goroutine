package domain_test

import (
	"errors"
	"testing"

	"goroutine/internal/domain"

	"github.com/google/uuid"
)

func TestNewUserID(t *testing.T) {
	t.Parallel()

	id := domain.NewUserID()

	if id.IsEmpty() {
		t.Error("NewUserID() should not be empty")
	}
	if id.UUID().Version() != 7 {
		t.Errorf("Expected UUID v7, got v%d", id.UUID().Version())
	}
}

func TestParseUserID(t *testing.T) {
	t.Parallel()

	u := uuid.New()
	s := u.String()

	id, err := domain.ParseUserID(s)
	if err != nil {
		t.Errorf("ParseUserID() error = %v", err)
	}

	if id.String() != s {
		t.Errorf("Expected %q, got %q", s, id.String())
	}

	_, err = domain.ParseUserID("invalid")
	if err == nil {
		t.Error("ParseUserID() with invalid string should return error")
	}
}

func TestUserID_Scan(t *testing.T) {
	t.Parallel()

	u := uuid.New()

	tests := []struct {
		name    string
		input   any
		wantErr bool
		errIs   error
	}{
		{
			name:    "Valid string",
			input:   u.String(),
			wantErr: false,
		},
		{
			name:    "Valid bytes",
			input:   u[:],
			wantErr: false,
		},
		{
			name:    "Invalid string",
			input:   "invalid-uuid",
			wantErr: true,
			errIs:   domain.ErrDataCorrupted,
		},
		{
			name:    "Invalid bytes",
			input:   []byte("invalid"),
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

			var id domain.UserID
			err := id.Scan(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("expected error %v, got %v", tt.errIs, err)
				}
			} else {
				if err != nil {
					t.Errorf("did not expect error, got %v", err)
				}
				if tt.input != nil && id.IsEmpty() {
					t.Error("expected UserID to not be empty")
				}
			}
		})
	}
}

func TestUserPassword(t *testing.T) {
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

			_, err := domain.NewUserPassword(tt.input)
			if tt.expectErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("did not expect error but got: %v", err)
			}
		})
	}
}

func TestUserPassword_Scan(t *testing.T) {
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

			var p domain.UserPassword
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
