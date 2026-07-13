package domain_test

import (
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
