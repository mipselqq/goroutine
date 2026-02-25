package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"goroutine/internal/http/httpschema"
)

type TokenVerifier interface {
	VerifyToken(ctx context.Context, token string) (int64, error)
}

type Auth struct {
	logger   *slog.Logger
	verifier TokenVerifier
}

func NewAuth(l *slog.Logger, v TokenVerifier) *Auth {
	return &Auth{logger: l, verifier: v}
}

func (m *Auth) Wrap(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			httpschema.RespondWithError(w, m.logger, http.StatusUnauthorized, errors.New("missing authorization header"))

			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
			httpschema.RespondWithError(w, m.logger, http.StatusUnauthorized, errors.New("invalid authorization header"))
			return
		}

		token := parts[1]
		userID, err := m.verifier.VerifyToken(r.Context(), token)
		if err != nil {
			httpschema.RespondWithError(w, m.logger, http.StatusUnauthorized, errors.New("invalid token"))
			return
		}

		ctx := context.WithValue(r.Context(), httpschema.ContextKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
