package httpschema_test

import (
	"testing"

	"goroutine/internal/http/httpschema"
)

func TestMapCodeToDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		code        string
		description string
	}{
		{
			name:        "known code",
			code:        "INVALID_CREDENTIALS",
			description: "Invalid login or password",
		},
		{
			name:        "known move error code",
			code:        "INDEX_OUT_OF_BOUNDS",
			description: "Index out of bounds",
		},
		{
			name:        "known conflict error code",
			code:        "USER_ALREADY_EXISTS",
			description: "User already exists",
		},
		// Other cases omitted since they're identical to the known code
		{
			name:        "unknown code fallback",
			code:        "SOME_NON_EXISTENT_CODE",
			description: "Unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			description := httpschema.MapCodeToDescription(tt.code)
			if description != tt.description {
				t.Errorf("got %q, want %q", description, tt.description)
			}
		})
	}
}
