package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"goroutine/internal/http/httpschema"
)

type GenerateUserID func() string

type RequestID struct {
	logger         *slog.Logger
	generateUserID GenerateUserID
}

func NewRequestID(l *slog.Logger, g GenerateUserID) *RequestID {
	return &RequestID{logger: l, generateUserID: g}
}

func (m *RequestID) Wrap(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := strings.TrimSpace(r.Header.Get("X-Request-Id"))
		if reqID == "" {
			reqID = m.generateUserID()
		}

		ctx := context.WithValue(r.Context(), httpschema.ContextKeyRequestID, reqID)
		w.Header().Set("X-Request-Id", reqID)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
