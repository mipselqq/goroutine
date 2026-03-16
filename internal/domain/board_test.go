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
				var ve *domain.ErrValidation
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
				t.Errorf("expected value %q, got %q", tt.expectedValue, name)
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
				var validationError *domain.ErrValidation
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
				t.Errorf("expected value %q, got %q", tt.expectedValue, description)
			}
			if description.IsEmpty() != (tt.expectedValue == "") {
				t.Errorf("expected is empty %t, got %t", tt.expectedValue == "", description.IsEmpty())
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
					t.Error("expected error, got nil")
				} else if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("expected error %v, got %v", tt.errIs, err)
				}
			} else if err != nil {
				t.Errorf("did not expect error, got %v", err)
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
					t.Error("expected error, got nil")
				} else if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("expected error %v, got %v", tt.errIs, err)
				}
			} else if err != nil {
				t.Errorf("did not expect error, got %v", err)
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
					t.Error("expected error, got nil")
				} else if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("expected error %v, got %v", tt.errIs, err)
				}
			} else {
				if err != nil {
					t.Errorf("did not expect error, got %v", err)
				}
				if tt.input != nil && id.IsEmpty() {
					t.Error("expected BoardID to not be empty")
				}
			}
		})
	}
}
