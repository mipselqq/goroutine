package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"goroutine/internal/http/httpschema"
)

type GenerateRequestIDFn func() string

type RequestID struct {
	logger              *slog.Logger
	generateRequestIDFn GenerateRequestIDFn
}

func MustNewRequestID(l *slog.Logger, g GenerateRequestIDFn) *RequestID {
	if g == nil {
		panic("BUG: generateRequestIDFn is nil")
	}

	return &RequestID{logger: l, generateRequestIDFn: g}
}

func (m *RequestID) Wrap(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := strings.TrimSpace(r.Header.Get("X-Request-Id"))
		if reqID == "" {
			reqID = m.generateRequestIDFn()
		}

		ctx := context.WithValue(r.Context(), httpschema.ContextKeyRequestID, reqID)
		w.Header().Set("X-Request-Id", reqID)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
