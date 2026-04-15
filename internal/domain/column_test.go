package domain_test

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"

	"goroutine/internal/domain"
)

func TestColumnName(t *testing.T) {
	t.Parallel()

	borderlineLongName := strings.Repeat("a", 128)
	tests := []struct {
		name           string
		input          string
		expectedIssues []string
		expectedValue  string
	}{
		{name: "Valid", input: "In Progress", expectedValue: "In Progress"},
		{name: "Long valid", input: borderlineLongName, expectedValue: borderlineLongName},
		{name: "Too long but valid when trimmed", input: "   " + borderlineLongName + "   ", expectedValue: borderlineLongName},
		{name: "Too long", input: borderlineLongName + "a", expectedIssues: []string{domain.ErrColumnNameTooLong}},
		{name: "Empty", input: "", expectedIssues: []string{domain.ErrColumnNameTooShort}},
		{name: "Whitespace", input: "   ", expectedIssues: []string{domain.ErrColumnNameTooShort}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			name, err := domain.NewColumnName(tt.input)
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
				t.Errorf("expected value %q, got %q", tt.expectedValue, name.String())
			}
		})
	}
}

func TestColumnPosition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          int64
		expectedIssues []string
		expectedValue  int64
	}{
		{name: "Valid", input: 1, expectedValue: 1},
		{name: "Valid max int32", input: math.MaxInt32, expectedValue: math.MaxInt32},
		{name: "Zero", input: 0, expectedIssues: []string{domain.ErrColumnPositionValue}},
		{name: "Negative", input: -10, expectedIssues: []string{domain.ErrColumnPositionValue}},
		{name: "Overflow", input: math.MaxInt32 + 1, expectedIssues: []string{domain.ErrColumnPositionValue}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			position, err := domain.NewColumnPosition(tt.input)
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
			if tt.expectedIssues == nil && position.Int64() != tt.expectedValue {
				t.Errorf("expected value %d, got %d", tt.expectedValue, position.Int64())
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
					t.Error("expected error, got nil")
				} else if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("expected error %v, got %v", tt.errIs, err)
				}
				return
			}

			if err != nil {
				t.Errorf("did not expect error, got %v", err)
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
		{name: "Valid int32", input: int32(1)},
		{name: "Valid int64", input: int64(2)},
		{name: "Zero", input: int64(0), wantErr: true, errIs: domain.ErrDataCorrupted},
		{name: "Overflow", input: int64(math.MaxInt32 + 1), wantErr: true, errIs: domain.ErrDataCorrupted},
		{name: "Nil", input: nil},
		{name: "Bad type", input: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var position domain.ColumnPosition
			err := position.Scan(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("expected error %v, got %v", tt.errIs, err)
				}
				return
			}

			if err != nil {
				t.Errorf("did not expect error, got %v", err)
			}
		})
	}
}
