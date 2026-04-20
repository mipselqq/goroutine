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
		t.Error("got empty UserID from NewUserID(), want non-empty")
	}
	if id.UUID().Version() != 7 {
		t.Errorf("got UUID version %d, want 7", id.UUID().Version())
	}
}

func TestParseUserID(t *testing.T) {
	t.Parallel()

	u := uuid.New()
	s := u.String()

	id, err := domain.ParseUserID(s)
	if err != nil {
		t.Errorf("ParseUserID(): got error %v, want nil", err)
	}

	if id.String() != s {
		t.Errorf("ParseUserID() = %q, want %q", id, s)
	}

	_, err = domain.ParseUserID("invalid")
	if err == nil {
		t.Error("got nil error from ParseUserID(\"invalid\"), want non-nil")
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
					t.Error("got nil error, want non-nil")
				} else if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("got error %v, want %v", err, tt.errIs)
				}
			} else {
				if err != nil {
					t.Errorf("got error %v, want nil", err)
				}
				if tt.input != nil && id.IsEmpty() {
					t.Error("got empty UserID after Scan, want non-empty")
				}
			}
		})
	}
}

func TestUserPassword(t *testing.T) {
	t.Parallel()

	passwordTests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Valid password",
			input:   "securePass123",
			wantErr: false,
		},
		{
			name:    "Less than 6 characters",
			input:   "12345",
			wantErr: true,
		},
		{
			name:    "Empty password",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Whitespace password",
			input:   "     ",
			wantErr: true,
		},
	}

	for _, tt := range passwordTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.NewUserPassword(tt.input)
			if tt.wantErr && err == nil {
				t.Error("got nil error, want non-nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("got error %v, want nil", err)
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
					t.Error("got nil error, want non-nil")
				} else if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("got error %v, want %v", err, tt.errIs)
				}
			} else if err != nil {
				t.Errorf("got error %v, want nil", err)
			}
		})
	}
}
