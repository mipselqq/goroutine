package middleware

import (
	"log/slog"
	"net/http"
)

type CORS struct {
	allowedOrigins map[string]struct{}
}

func NewCORS(logger *slog.Logger, allowedOrigins map[string]struct{}) *CORS {
	for origin := range allowedOrigins {
		if origin == "*" {
			logger.Warn("CORS middleware is too permissive, allowing any origin")
			return &CORS{allowedOrigins: allowedOrigins}
		}
	}

	return &CORS{allowedOrigins: allowedOrigins}
}

func (m *CORS) Wrap(next http.Handler) http.HandlerFunc {
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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
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
