package service

import (
	"testing"
	"time"
)

func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		attempts int
		want     time.Duration
	}{
		{attempts: 0, want: time.Second},
		{attempts: 1, want: 2 * time.Second},
		{attempts: 2, want: 4 * time.Second},
		{attempts: 3, want: 8 * time.Second},
		{attempts: 8, want: 256 * time.Second},
		{attempts: 9, want: 5 * time.Minute},
		{attempts: 20, want: 5 * time.Minute},
	}

	for _, tt := range tests {
		got := CalculateBackoff(tt.attempts, time.Second, 5*time.Minute)
		if got != tt.want {
			t.Errorf("CalculateBackoff(%d) = %v, want %v", tt.attempts, got, tt.want)
		}
	}
}
