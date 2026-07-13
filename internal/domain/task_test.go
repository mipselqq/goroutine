package domain_test

import (
	"math"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"goroutine/internal/domain"
)

func TestTaskName(t *testing.T) {
	t.Parallel()

	borderlineLongName := strings.Repeat("a", 128)
	tests := []struct {
		name       string
		input      string
		wantIssues []string
		wantValue  string
	}{
		{name: "Valid", input: "Write tests", wantValue: "Write tests"},
		{name: "Long valid", input: borderlineLongName, wantValue: borderlineLongName},
		{name: "Too long but valid when trimmed", input: "   " + borderlineLongName + "   ", wantValue: borderlineLongName},
		{name: "Too long", input: borderlineLongName + "a", wantIssues: []string{domain.ErrTaskNameTooLong}},
		{name: "Empty", input: "", wantIssues: []string{domain.ErrTaskNameTooShort}},
		{name: "Whitespace", input: "   ", wantIssues: []string{domain.ErrTaskNameTooShort}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			name, err := domain.NewTaskName(tt.input)
			var gotIssues []string
			if err != nil {
				gotIssues = domain.ExtractValidationIssues(err)
			}

			if diff := cmp.Diff(tt.wantIssues, gotIssues); diff != "" {
				t.Errorf("got issues mismatch (-want +got):\n%s", diff)
			}
			if name.String() != tt.wantValue {
				t.Errorf("got value %q, want %q", name.String(), tt.wantValue)
			}
		})
	}
}

func TestTaskDescription(t *testing.T) {
	t.Parallel()

	borderlineLongDescription := strings.Repeat("a", 1024)

	tests := []struct {
		name       string
		input      string
		wantIssues []string
		wantValue  string
	}{
		{name: "Valid", input: "Write integration tests for tasks", wantValue: "Write integration tests for tasks"},
		{name: "Long valid", input: borderlineLongDescription, wantValue: borderlineLongDescription},
		{name: "Too long but valid when trimmed", input: "   " + borderlineLongDescription + "   ", wantValue: borderlineLongDescription},
		{name: "Too long", input: borderlineLongDescription + "a", wantIssues: []string{domain.ErrTaskDescriptionTooLong}},
		{name: "Empty", input: "", wantValue: ""},
		{name: "Whitespace", input: "    ", wantValue: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			description, err := domain.NewTaskDescription(tt.input)
			var gotIssues []string
			if err != nil {
				gotIssues = domain.ExtractValidationIssues(err)
			}

			if diff := cmp.Diff(tt.wantIssues, gotIssues); diff != "" {
				t.Errorf("got issues mismatch (-want +got):\n%s", diff)
			}
			if description.String() != tt.wantValue {
				t.Errorf("got value %q, want %q", description.String(), tt.wantValue)
			}
		})
	}
}

func TestTaskPosition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      int64
		wantIssues []string
		wantValue  int64
	}{
		{name: "Valid", input: 1, wantValue: 1},
		{name: "Valid max int32", input: math.MaxInt32, wantValue: math.MaxInt32},
		{name: "Zero", input: 0, wantIssues: []string{domain.ErrTaskPositionValue}},
		{name: "Negative", input: -10, wantIssues: []string{domain.ErrTaskPositionValue}},
		{name: "Overflow", input: math.MaxInt32 + 1, wantIssues: []string{domain.ErrTaskPositionValue}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			position, err := domain.NewTaskPosition(tt.input)
			var gotIssues []string
			if err != nil {
				gotIssues = domain.ExtractValidationIssues(err)
			}

			if diff := cmp.Diff(tt.wantIssues, gotIssues); diff != "" {
				t.Errorf("got issues mismatch (-want +got):\n%s", diff)
			}
			if tt.wantIssues == nil && position.Int64() != tt.wantValue {
				t.Errorf("got value %d, want %d", position.Int64(), tt.wantValue)
			}
		})
	}
}
