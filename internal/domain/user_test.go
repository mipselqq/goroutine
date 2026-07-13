package domain_test

import (
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/secrecy"
	"goroutine/internal/testutil"

	"github.com/google/uuid"
)

func TestNewUserID(t *testing.T) {
	t.Parallel()

	id := domain.NewUserID()

	if id.String() == "" {
		t.Error("got empty UserID from NewUserID(), want non-empty")
	}
	if id.UUID().Version() != 7 {
		t.Errorf("got UUID version %d, want 7", id.UUID().Version())
	}
}

func TestParseUserID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid UUID",
			input:   uuid.New().String(),
			wantErr: false,
		},
		{
			name:    "invalid string",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "nil UUID",
			input:   "00000000-0000-0000-0000-000000000000",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			id, err := domain.ParseUserID(tt.input)
			if tt.wantErr && err == nil {
				t.Errorf("got id %q, want error", id)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("got error %v, want nil", err)
			}
			if !tt.wantErr && id.String() != tt.input {
				t.Errorf("got %q, want %q", id, tt.input)
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

			password, err := domain.NewUserPassword(tt.input)
			if tt.wantErr && err == nil {
				t.Error("got nil error, want non-nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("got error %v, want nil", err)
			}
			if !tt.wantErr {
				testutil.AssertSecretHidden(t, tt.input, password)
			}
		})
	}
}

func TestNewJWTString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "Valid JWT",
			input:   "header.payload.signature",
			wantErr: false,
		},
		{
			name:    "Empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Whitespace string",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "Too few parts",
			input:   "header.payload",
			wantErr: true,
		},
		{
			name:    "Too many parts",
			input:   "header.payload.signature.extra",
			wantErr: true,
		},
		{
			name:    "Empty middle part",
			input:   "header..signature",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			token, err := domain.NewJWTString(tt.input)
			if tt.wantErr && err == nil {
				t.Error("got nil error, want non-nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("got error %v, want nil", err)
			}
			if !tt.wantErr {
				testutil.AssertSecretHidden(t, tt.input, token)
				if token.RevealSecret() != tt.input {
					t.Errorf("got revealed token %q, want %q", token.RevealSecret(), tt.input)
				}
			}
		})
	}
}

func TestPasswordHash(t *testing.T) {
	t.Parallel()

	raw := "$argon2id$v=19$m=65536,t=1,p=16$hashhashhashhashhashhash"
	hash := domain.PasswordHash{SecretString: secrecy.SecretString(raw)}

	testutil.AssertSecretHidden(t, raw, hash)

	if hash.RevealSecret() != raw {
		t.Errorf("got revealed hash %q, want %q", hash.RevealSecret(), raw)
	}
}
