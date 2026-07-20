package middleware

import (
	"log/slog"
	"net/http"

	"goroutine/internal/logging"
)

// Every method used by NewRouter + OPTIONS for preflight
const corsAllowedMethods = "DELETE, GET, OPTIONS, PATCH, POST, PUT"

type cors struct {
	allowedOrigins map[string]struct{}
}

func NewCORS(logger *slog.Logger, allowedOrigins map[string]struct{}) *cors {
	moduleLogger := logging.WithModule(logger, "middleware.cors")

	for origin := range allowedOrigins {
		if origin == "*" {
			moduleLogger.Warn("CORS middleware is too permissive, allowing any origin")
			return &cors{allowedOrigins: allowedOrigins}
		}
	}

	return &cors{allowedOrigins: allowedOrigins}
}

func (m *cors) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Origin")

		origin := r.Header.Get("Origin")

		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		if _, contains := m.allowedOrigins[origin]; !contains {
			if _, containsAll := m.allowedOrigins["*"]; !containsAll {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", corsAllowedMethods)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
