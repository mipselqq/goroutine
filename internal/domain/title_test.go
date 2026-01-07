package domain_test

import (
	"testing"

	"go-todo/internal/domain"
)

var titleTests = []struct {
	name      string
	input     string
	expectErr bool
}{
	{
		name:      "Valid title",
		input:     "My Todo Title",
		expectErr: false,
	},
	{
		name:      "Empty title",
		input:     "",
		expectErr: true,
	},
	{
		name:      "Whitespace title",
		input:     "    ",
		expectErr: true,
	},
}

func TestTitle(t *testing.T) {
	for _, tt := range titleTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewTitle(tt.input)
			if tt.expectErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("did not expect error but got: %v", err)
			}
		})
	}
}

func FuzzTitle(f *testing.F) {
	for _, tt := range titleTests {
		f.Add(tt.input)
	}
	f.Fuzz(func(t *testing.T, input string) {
		_, err := domain.NewTitle(input)
		if err == nil {
		}
	})
}
