package middleware

import (
	"context"
	"net/http"
	"time"
)

type Timeout struct {
	Duration time.Duration
}

func NewTimeout(d time.Duration) *Timeout {
	return &Timeout{Duration: d}
}

// Prevents unexpected resource leaks by forcing context cancellation after specified duration
func (m *Timeout) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), m.Duration)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
