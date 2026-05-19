package domain_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"goroutine/internal/domain"
	"goroutine/internal/testutil"
)

func TestTelegramLinkToken(t *testing.T) {
	t.Parallel()

	uuidv7, _ := uuid.NewV7()
	uuidv4 := uuid.New()

	tests := []struct {
		name       string
		input      string
		wantIssues []string
	}{
		{
			name:  "Valid UUIDv7 token",
			input: uuidv7.String(),
		},
		{
			name:       "Deny UUIDv4 token",
			input:      uuidv4.String(),
			wantIssues: []string{domain.ErrInvalidTelegramLinkToken},
		},
		{
			name:       "Empty token",
			input:      "",
			wantIssues: []string{domain.ErrInvalidTelegramLinkToken},
		},
		{
			name:       "Invalid format",
			input:      "some-random-string",
			wantIssues: []string{domain.ErrInvalidTelegramLinkToken},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			token, err := domain.NewTelegramLinkToken(tt.input)
			var gotIssues []string
			if err != nil {
				gotIssues = domain.ExtractValidationIssues(err)
			}

			if diff := cmp.Diff(tt.wantIssues, gotIssues); diff != "" {
				t.Errorf("got issues mismatch (-want +got):\n%s", diff)
			}

			if tt.wantIssues == nil {
				testutil.AssertSecretHidden(t, tt.input, token)
				if token.RevealSecret() != tt.input {
					t.Errorf("RevealSecret() = %q, want %q", token.RevealSecret(), tt.input)
				}
			}
		})
	}
}

func TestTelegramLinkToken_Scan(t *testing.T) {
	t.Parallel()

	uuidv7, _ := uuid.NewV7()

	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
		errIs   error
	}{
		{
			name:    "Valid UUIDv7 string",
			input:   uuidv7.String(),
			want:    uuidv7.String(),
			wantErr: false,
		},
		{
			name:    "Deny UUIDv4 string",
			input:   uuid.New().String(),
			wantErr: true,
			errIs:   domain.ErrDataCorrupted,
		},
		{
			name:    "Null value",
			input:   nil,
			want:    "",
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

			var tok domain.TelegramLinkToken
			err := tok.Scan(tt.input)

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
				if tok.RevealSecret() != tt.want {
					t.Errorf("got %q, want %q", tok.RevealSecret(), tt.want)
				}
				if tt.want != "" {
					testutil.AssertSecretHidden(t, tt.want, tok)
				}
			}
		})
	}
}

func TestTelegramChatID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      int64
		wantIssues []string
	}{
		{
			name:  "Valid positive user ID",
			input: 123456789,
		},
		{
			name:  "Valid negative group ID",
			input: -1001234567890,
		},
		{
			name:       "Invalid zero ID",
			input:      0,
			wantIssues: []string{domain.ErrInvalidTelegramChatID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			chatID, err := domain.NewTelegramChatID(tt.input)
			var gotIssues []string
			if err != nil {
				gotIssues = domain.ExtractValidationIssues(err)
			}

			if diff := cmp.Diff(tt.wantIssues, gotIssues); diff != "" {
				t.Errorf("got issues mismatch (-want +got):\n%s", diff)
			}

			if tt.wantIssues == nil && chatID.Int64() != tt.input {
				t.Errorf("Int64() = %d, want %d", chatID.Int64(), tt.input)
			}
		})
	}
}

func TestParseTelegramChatID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{
			name:    "Valid positive string",
			input:   " 123456 ",
			want:    123456,
			wantErr: false,
		},
		{
			name:    "Valid negative string",
			input:   "-10023456",
			want:    -10023456,
			wantErr: false,
		},
		{
			name:    "Invalid zero",
			input:   "0",
			wantErr: true,
		},
		{
			name:    "Non-numeric string",
			input:   "abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			chatID, err := domain.ParseTelegramChatID(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("got nil error, want non-nil")
				}
			} else {
				if err != nil {
					t.Errorf("got unexpected error: %v", err)
				}
				if chatID.Int64() != tt.want {
					t.Errorf("got %d, want %d", chatID.Int64(), tt.want)
				}
			}
		})
	}
}

func TestTelegramChatID_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    int64
		wantErr bool
	}{
		{
			name:    "Int64 type",
			input:   int64(9876),
			want:    9876,
			wantErr: false,
		},
		{
			name:    "Null value",
			input:   nil,
			want:    0,
			wantErr: false,
		},
		{
			name:    "Zero value is invalid",
			input:   int64(0),
			wantErr: true,
		},
		{
			name:    "Invalid type (string)",
			input:   "-555",
			wantErr: true,
		},
		{
			name:    "Invalid type (int)",
			input:   123,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var chatID domain.TelegramChatID
			err := chatID.Scan(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("got nil error, want non-nil")
				}
			} else {
				if err != nil {
					t.Errorf("got unexpected error: %v", err)
				}
				if chatID.Int64() != tt.want {
					t.Errorf("got %d, want %d", chatID.Int64(), tt.want)
				}
			}
		})
	}
}

func TestTelegramUsername(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		wantIssues []string
	}{
		{
			name:  "Valid username",
			input: "@user_123",
		},
		{
			name:  "Valid username with uppercase",
			input: "@MyAwesomeUser",
		},
		{
			name:       "Invalid: no at symbol",
			input:      "user_123",
			wantIssues: []string{domain.ErrInvalidTelegramUsername},
		},
		{
			name:       "Invalid: too short",
			input:      "@a_bc",
			wantIssues: []string{domain.ErrInvalidTelegramUsername},
		},
		{
			name:  "Valid: minimum length 5",
			input: "@a_b_c",
		},
		{
			name:       "Invalid: too long",
			input:      "@a_very_long_username_that_is_definitely_more_than_thirty_two_chars",
			wantIssues: []string{domain.ErrInvalidTelegramUsername},
		},
		{
			name:       "Invalid: starting with underscore",
			input:      "@_username",
			wantIssues: []string{domain.ErrInvalidTelegramUsername},
		},
		{
			name:       "Invalid: starting with number",
			input:      "@1username",
			wantIssues: []string{domain.ErrInvalidTelegramUsername},
		},
		{
			name:       "Invalid characters (dash)",
			input:      "@user-name",
			wantIssues: []string{domain.ErrInvalidTelegramUsername},
		},
		{
			name:       "Invalid characters (dot)",
			input:      "@user.name",
			wantIssues: []string{domain.ErrInvalidTelegramUsername},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			username, err := domain.NewTelegramUsername(tt.input)
			var gotIssues []string
			if err != nil {
				gotIssues = domain.ExtractValidationIssues(err)
			}

			if diff := cmp.Diff(tt.wantIssues, gotIssues); diff != "" {
				t.Errorf("got issues mismatch (-want +got):\n%s", diff)
			}

			if tt.wantIssues == nil && username.String() != tt.input {
				t.Errorf("got %q, want %q", username.String(), tt.input)
			}
		})
	}
}

func TestTelegramUsername_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name:    "Valid string",
			input:   "@user_123",
			want:    "@user_123",
			wantErr: false,
		},
		{
			name:    "Null value",
			input:   nil,
			want:    "",
			wantErr: false,
		},
		{
			name:    "Invalid string",
			input:   "no_at_symbol",
			wantErr: true,
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

			var username domain.TelegramUsername
			err := username.Scan(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("got nil error, want non-nil")
				}
			} else {
				if err != nil {
					t.Errorf("got unexpected error: %v", err)
				}
				if username.String() != tt.want {
					t.Errorf("got %q, want %q", username.String(), tt.want)
				}
			}
		})
	}
}
