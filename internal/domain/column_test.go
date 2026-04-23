package domain_test

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"goroutine/internal/domain"
)

func TestColumnName(t *testing.T) {
	t.Parallel()

	borderlineLongName := strings.Repeat("a", 128)
	tests := []struct {
		name       string
		input      string
		wantIssues []string
		wantValue  string
	}{
		{name: "Valid", input: "In Progress", wantValue: "In Progress"},
		{name: "Long valid", input: borderlineLongName, wantValue: borderlineLongName},
		{name: "Too long but valid when trimmed", input: "   " + borderlineLongName + "   ", wantValue: borderlineLongName},
		{name: "Too long", input: borderlineLongName + "a", wantIssues: []string{domain.ErrColumnNameTooLong}},
		{name: "Empty", input: "", wantIssues: []string{domain.ErrColumnNameTooShort}},
		{name: "Whitespace", input: "   ", wantIssues: []string{domain.ErrColumnNameTooShort}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			name, err := domain.NewColumnName(tt.input)
			var gotIssues []string
			if err != nil {
				var ve *domain.ErrValidation
				if errors.As(err, &ve) {
					gotIssues = ve.Issues
				} else {
					gotIssues = []string{err.Error()}
				}
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

func TestParseColumnPosition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "Valid", input: "1"},
		{name: "Zero", input: "0", wantErr: true},
		{name: "Negative", input: "-1", wantErr: true},
		{name: "Greater than int32", input: "2147483648", wantErr: true},
		{name: "Greater than int64", input: "9223372036854775808", wantErr: true},
		{name: "Not a number", input: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.ParseColumnPosition(tt.input)
			if tt.wantErr && err == nil {
				t.Error("got nil error, want non-nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("got error %v, want nil", err)
			}
		})
	}
}

func TestColumnPosition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      int64
		wantIssues []string
		wantValue  int64
	}{
		{name: "Valid", input: 1, wantValue: 1},
		{name: "Valid max int32", input: math.MaxInt32, wantValue: math.MaxInt32},
		{name: "Zero", input: 0, wantIssues: []string{domain.ErrColumnPositionValue}},
		{name: "Negative", input: -10, wantIssues: []string{domain.ErrColumnPositionValue}},
		{name: "Overflow", input: math.MaxInt32 + 1, wantIssues: []string{domain.ErrColumnPositionValue}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			position, err := domain.NewColumnPosition(tt.input)
			var gotIssues []string
			if err != nil {
				var ve *domain.ErrValidation
				if errors.As(err, &ve) {
					gotIssues = ve.Issues
				} else {
					gotIssues = []string{err.Error()}
				}
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

func TestColumnName_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		wantErr bool
		errIs   error
	}{
		{name: "Valid", input: "Todo"},
		{name: "Too long", input: strings.Repeat("a", 129), wantErr: true, errIs: domain.ErrDataCorrupted},
		{name: "Empty", input: "", wantErr: true, errIs: domain.ErrDataCorrupted},
		{name: "Nil", input: nil},
		{name: "Bad type", input: 1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var name domain.ColumnName
			err := name.Scan(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("got nil error, want non-nil")
				} else if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("got error %v, want %v", err, tt.errIs)
				}
				return
			}

			if err != nil {
				t.Errorf("got error %v, want nil", err)
			}
		})
	}
}

func TestColumnDescription(t *testing.T) {
	t.Parallel()

	borderlineLongDescription := strings.Repeat("a", 1024)

	descriptionTests := []struct {
		name       string
		input      string
		wantIssues []string
		wantValue  string
	}{
		{
			name:       "Valid description",
			input:      "My Column Description",
			wantValue:  "My Column Description",
			wantIssues: nil,
		},
		{
			name:       "Long valid description",
			input:      borderlineLongDescription,
			wantValue:  borderlineLongDescription,
			wantIssues: nil,
		},
		{
			name:       "Too long but valid when trimmed",
			input:      "      " + borderlineLongDescription + "     ",
			wantValue:  borderlineLongDescription,
			wantIssues: nil,
		},
		{
			name:       "Too long description",
			input:      borderlineLongDescription + "a",
			wantIssues: []string{domain.ErrDescriptionTooLong},
			wantValue:  "",
		},
		{
			name:       "Empty description",
			input:      "",
			wantIssues: nil,
			wantValue:  "",
		},
		{
			name:       "Whitespace description",
			input:      "    ",
			wantIssues: nil,
			wantValue:  "",
		},
	}

	for _, tt := range descriptionTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			description, err := domain.NewColumnDescription(tt.input)

			var gotIssues []string
			if err != nil {
				var ve *domain.ErrValidation
				if errors.As(err, &ve) {
					gotIssues = ve.Issues
				} else {
					gotIssues = []string{err.Error()}
				}
			}

			if diff := cmp.Diff(tt.wantIssues, gotIssues); diff != "" {
				t.Errorf("got issues mismatch (-want +got):\n%s", diff)
			}
			if description.String() != tt.wantValue {
				t.Errorf("got value %q, want %q", description, tt.wantValue)
			}
			if description.IsEmpty() != (tt.wantValue == "") {
				t.Errorf("got is empty %t, want %t", description.IsEmpty(), tt.wantValue == "")
			}
		})
	}
}

func TestColumnDescription_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		wantErr bool
		errIs   error
	}{
		{
			name:    "Valid description",
			input:   "Valid description",
			wantErr: false,
		},
		{
			name:    "Too long description",
			input:   strings.Repeat("a", 1025),
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

			var d domain.ColumnDescription
			err := d.Scan(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("got nil error, want non-nil")
				} else if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("got error %v, want %v", err, tt.errIs)
				}
			} else if err != nil {
				t.Errorf("got error %v, want nil", err)
			}
		})
	}
}

func TestColumnPosition_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		wantErr bool
		errIs   error
	}{
		{name: "Valid int64", input: int64(1)},
		{name: "Zero", input: int64(0), wantErr: true, errIs: domain.ErrDataCorrupted},
		{name: "Nil", input: nil},
		{name: "Int32 is rejected", input: int32(2), wantErr: true},
		{name: "Bad type", input: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var position domain.ColumnPosition
			err := position.Scan(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("got nil error, want non-nil")
				} else if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("got error %v, want %v", err, tt.errIs)
				}
				return
			}

			if err != nil {
				t.Errorf("got error %v, want nil", err)
			}
		})
	}
}
