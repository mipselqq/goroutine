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
