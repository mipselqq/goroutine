package domain_test

import (
	"testing"

	"go-todo/internal/domain"
)

var passwordTests = []struct {
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

func TestPassword(t *testing.T) {
	for _, tt := range passwordTests {
		t.Run(tt.name, func(t *testing.T) {
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

func FuzzPassword(f *testing.F) {
	for _, tt := range passwordTests {
		f.Add(tt.input)
	}
	f.Fuzz(func(t *testing.T, input string) {
		_, err := domain.NewPassword(input)
		if err == nil {
		}
	})
}
