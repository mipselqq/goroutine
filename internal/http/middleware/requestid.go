package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"goroutine/internal/http/httpschema"
	"goroutine/internal/logging"
)

type generateRequestIDFn func() string

type requestID struct {
	logger *slog.Logger
	gen    generateRequestIDFn
}

func MustNewRequestID(logger *slog.Logger, gen generateRequestIDFn) *requestID {
	if gen == nil {
		panic("BUG: gen is nil")
	}

	moduleLogger := logging.WithModule(logger, "middleware.requestid")

	return &requestID{logger: moduleLogger, gen: gen}
}

func (m *requestID) Wrap(next http.Handler) http.Handler {
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
