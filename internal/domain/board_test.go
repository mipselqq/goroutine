package domain_test

import (
	"reflect"
	"strings"
	"testing"

	"goroutine/internal/domain"
)

func TestName(t *testing.T) {
	t.Parallel()

	borderlineLongName := strings.Repeat("a", 128)

	nameTests := []struct {
		name           string
		input          string
		expectedErrors []string
		expectedValue  string
	}{
		{
			name:           "Valid name",
			input:          "My Todo Name",
			expectedErrors: []string{},
			expectedValue:  "My Todo Name",
		},
		{
			name:           "Long valid name",
			input:          borderlineLongName,
			expectedErrors: []string{},
			expectedValue:  borderlineLongName,
		},
		{
			name:           "Too long but valid when trimmed",
			input:          "      " + borderlineLongName + "     ",
			expectedErrors: []string{},
			expectedValue:  borderlineLongName,
		},
		{
			name:           "Too long name",
			input:          borderlineLongName + "a",
			expectedErrors: []string{"Name is too long"},
			expectedValue:  "",
		},
		{
			name:           "Empty name",
			input:          "",
			expectedErrors: []string{"Name is too short"},
			expectedValue:  "",
		},
		{
			name:           "Whitespace name",
			input:          "    ",
			expectedErrors: []string{"Name is too short"},
			expectedValue:  "",
		},
	}

	for _, tt := range nameTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			name, errs := domain.NewBoardName(tt.input)

			if !reflect.DeepEqual(errs, tt.expectedErrors) {
				t.Errorf("expected errors %v, got %v", tt.expectedErrors, errs)
			}
			if name.String() != tt.expectedValue {
				t.Errorf("expected value %q, got %q", tt.expectedValue, name.String())
			}
			if name.IsEmpty() != (tt.expectedValue == "") {
				t.Errorf("expected is empty %t, got %t", tt.expectedValue == "", name.IsEmpty())
			}
		})
	}
}

func TestDescription(t *testing.T) {
	t.Parallel()

	borderlineLongDescription := strings.Repeat("a", 1024)

	descriptionTests := []struct {
		name           string
		input          string
		expectedErrors []string
		expectedValue  string
	}{
		{
			name:           "Valid description",
			input:          "My Todo Description",
			expectedValue:  "My Todo Description",
			expectedErrors: []string{},
		},
		{
			name:           "Long valid description",
			input:          borderlineLongDescription,
			expectedValue:  borderlineLongDescription,
			expectedErrors: []string{},
		},
		{
			name:           "Too long but valid when trimmed",
			input:          "      " + borderlineLongDescription + "     ",
			expectedValue:  borderlineLongDescription,
			expectedErrors: []string{},
		},
		{
			name:           "Too long description",
			input:          borderlineLongDescription + "a",
			expectedErrors: []string{"Description is too long"},
			expectedValue:  "",
		},
		{
			name:           "Empty description",
			input:          "",
			expectedErrors: []string{},
			expectedValue:  "",
		},
		{
			name:           "Whitespace description",
			input:          "    ",
			expectedErrors: []string{},
			expectedValue:  "",
		},
	}

	for _, tt := range descriptionTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			description, errs := domain.NewBoardDescription(tt.input)
			if !reflect.DeepEqual(errs, tt.expectedErrors) {
				t.Errorf("expected errors %v, got %v", tt.expectedErrors, errs)
			}
			if description.String() != tt.expectedValue {
				t.Errorf("expected value %q, got %q", tt.expectedValue, description.String())
			}
			if description.IsEmpty() != (tt.expectedValue == "") {
				t.Errorf("expected is empty %t, got %t", tt.expectedValue == "", description.IsEmpty())
			}
		})
	}
}
