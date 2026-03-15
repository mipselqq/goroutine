package domain_test

import (
	"errors"
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
		expectedIssues []string
		expectedValue  string
	}{
		{
			name:           "Valid name",
			input:          "My Todo Name",
			expectedIssues: nil,
			expectedValue:  "My Todo Name",
		},
		{
			name:           "Long valid name",
			input:          borderlineLongName,
			expectedIssues: nil,
			expectedValue:  borderlineLongName,
		},
		{
			name:           "Too long but valid when trimmed",
			input:          "      " + borderlineLongName + "     ",
			expectedIssues: nil,
			expectedValue:  borderlineLongName,
		},
		{
			name:           "Too long name",
			input:          borderlineLongName + "a",
			expectedIssues: []string{"Name is too long"},
			expectedValue:  "",
		},
		{
			name:           "Empty name",
			input:          "",
			expectedIssues: []string{"Name is too short"},
			expectedValue:  "",
		},
		{
			name:           "Whitespace name",
			input:          "    ",
			expectedIssues: []string{"Name is too short"},
			expectedValue:  "",
		},
		{
			name:           "Single character name",
			input:          "a",
			expectedIssues: nil,
			expectedValue:  "a",
		},
	}

	for _, tt := range nameTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			name, err := domain.NewBoardName(tt.input)

			var actualIssues []string
			if err != nil {
				var ve *domain.ValidationError
				if errors.As(err, &ve) {
					actualIssues = ve.Issues
				} else {
					actualIssues = []string{err.Error()}
				}
			}

			if !reflect.DeepEqual(actualIssues, tt.expectedIssues) {
				t.Errorf("expected issues %v, got %v", tt.expectedIssues, actualIssues)
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
		expectedIssues []string
		expectedValue  string
	}{
		{
			name:           "Valid description",
			input:          "My Todo Description",
			expectedValue:  "My Todo Description",
			expectedIssues: nil,
		},
		{
			name:           "Long valid description",
			input:          borderlineLongDescription,
			expectedValue:  borderlineLongDescription,
			expectedIssues: nil,
		},
		{
			name:           "Too long but valid when trimmed",
			input:          "      " + borderlineLongDescription + "     ",
			expectedValue:  borderlineLongDescription,
			expectedIssues: nil,
		},
		{
			name:           "Too long description",
			input:          borderlineLongDescription + "a",
			expectedIssues: []string{"Description is too long"},
			expectedValue:  "",
		},
		{
			name:           "Empty description",
			input:          "",
			expectedIssues: nil,
			expectedValue:  "",
		},
		{
			name:           "Whitespace description",
			input:          "    ",
			expectedIssues: nil,
			expectedValue:  "",
		},
	}

	for _, tt := range descriptionTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			description, err := domain.NewBoardDescription(tt.input)

			var actualIssues []string
			if err != nil {
				var validationError *domain.ValidationError
				if errors.As(err, &validationError) {
					actualIssues = validationError.Issues
				} else {
					actualIssues = []string{err.Error()}
				}
			}

			if !reflect.DeepEqual(actualIssues, tt.expectedIssues) {
				t.Errorf("expected issues %v, got %v", tt.expectedIssues, actualIssues)
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
