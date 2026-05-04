package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"goroutine/internal/domain"
)

func TestName(t *testing.T) {
	t.Parallel()

	borderlineLongName := strings.Repeat("a", 128)

	nameTests := []struct {
		name       string
		input      string
		wantIssues []string
		wantValue  string
	}{
		{
			name:       "Valid name",
			input:      "My Todo Name",
			wantIssues: nil,
			wantValue:  "My Todo Name",
		},
		{
			name:       "Long valid name",
			input:      borderlineLongName,
			wantIssues: nil,
			wantValue:  borderlineLongName,
		},
		{
			name:       "Too long but valid when trimmed",
			input:      "      " + borderlineLongName + "     ",
			wantIssues: nil,
			wantValue:  borderlineLongName,
		},
		{
			name:       "Too long name",
			input:      borderlineLongName + "a",
			wantIssues: []string{"Name is too long"},
			wantValue:  "",
		},
		{
			name:       "Empty name",
			input:      "",
			wantIssues: []string{"Name is too short"},
			wantValue:  "",
		},
		{
			name:       "Whitespace name",
			input:      "    ",
			wantIssues: []string{"Name is too short"},
			wantValue:  "",
		},
		{
			name:       "Single character name",
			input:      "a",
			wantIssues: nil,
			wantValue:  "a",
		},
	}

	for _, tt := range nameTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			name, err := domain.NewBoardName(tt.input)

			var gotIssues []string
			if err != nil {
				gotIssues = domain.ExtractValidationIssues(err)
			}

			if diff := cmp.Diff(tt.wantIssues, gotIssues); diff != "" {
				t.Errorf("got issues mismatch (-want +got):\n%s", diff)
			}
			if name.String() != tt.wantValue {
				t.Errorf("got value %q, want %q", name, tt.wantValue)
			}
			if name.IsEmpty() != (tt.wantValue == "") {
				t.Errorf("got is empty %t, want %t", name.IsEmpty(), tt.wantValue == "")
			}
		})
	}
}

func TestDescription(t *testing.T) {
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
			input:      "My Todo Description",
			wantValue:  "My Todo Description",
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
			wantIssues: []string{"Description is too long"},
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

			description, err := domain.NewBoardDescription(tt.input)

			var gotIssues []string
			if err != nil {
				gotIssues = domain.ExtractValidationIssues(err)
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

func TestBoardName_Scan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		wantErr bool
		errIs   error
	}{
		{
			name:    "Valid name",
			input:   "Valid Board Name",
			wantErr: false,
		},
		{
			name:    "Too long name",
			input:   strings.Repeat("a", 129),
			wantErr: true,
			errIs:   domain.ErrDataCorrupted,
		},
		{
			name:    "Empty name",
			input:   "",
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

			var n domain.BoardName
			err := n.Scan(tt.input)

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

func TestBoardDescription_Scan(t *testing.T) {
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

			var d domain.BoardDescription
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

func TestBoardID_Scan(t *testing.T) {
	t.Parallel()

	u := domain.NewBoardID()
	uid := u.UUID()

	tests := []struct {
		name    string
		input   any
		wantErr bool
		errIs   error
	}{
		{
			name:    "Valid string",
			input:   u.String(),
			wantErr: false,
		},
		{
			name:    "Valid bytes",
			input:   uid[:],
			wantErr: false,
		},
		{
			name:    "Invalid string",
			input:   "invalid-uuid",
			wantErr: true,
			errIs:   domain.ErrDataCorrupted,
		},
		{
			name:    "Invalid bytes",
			input:   []byte("invalid"),
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

			var id domain.BoardID
			err := id.Scan(tt.input)

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
				if tt.input != nil && id.IsEmpty() {
					t.Error("got empty BoardID after Scan, want non-empty")
				}
			}
		})
	}
}
