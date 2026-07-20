package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"goroutine/internal/http/httpschema"
	"goroutine/internal/logging"
)

type GenerateRequestIDFn func() string

type RequestID struct {
	logger *slog.Logger
	gen    GenerateRequestIDFn
}

func MustNewRequestID(logger *slog.Logger, gen GenerateRequestIDFn) *RequestID {
	if gen == nil {
		panic("BUG: gen is nil")
	}

	moduleLogger := logging.WithModule(logger, "middleware.requestid")

	return &RequestID{logger: moduleLogger, gen: gen}
}

func (m *RequestID) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := strings.TrimSpace(r.Header.Get("X-Request-Id"))
		if reqID == "" {
			reqID = m.gen()
		}

		ctx := context.WithValue(r.Context(), httpschema.ContextKeyRequestID, reqID)
		w.Header().Set("X-Request-Id", reqID)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
