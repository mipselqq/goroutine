package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"goroutine/internal/domain"
	"goroutine/internal/http/httpschema"
)

type TokenVerifier interface {
	VerifyToken(ctx context.Context, token string) (domain.UserID, error)
}

type Auth struct {
	logger    *slog.Logger
	verifier  TokenVerifier
	responder *httpschema.ErrorResponder
}

func NewAuth(l *slog.Logger, v TokenVerifier, r *httpschema.ErrorResponder) *Auth {
	return &Auth{logger: l, verifier: v, responder: r}
}

func (m *Auth) Wrap(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			m.responder.Unauthorized(
				w, "INVALID_AUTH_HEADER",
				[]httpschema.Detail{{Field: "Authorization", Issues: []string{"Missing authorization header"}}},
			)
			return
		}

		issues := []string{}
		authHeader := strings.TrimSpace(header)
		parts := strings.Fields(authHeader)

		// TODO: test middleware http responses as well
		if len(parts) != 2 {
			issues = append(issues, "Invalid authorization header")
		} else if !strings.EqualFold(parts[0], "bearer") {
			issues = append(issues, "No Bearer prefix")
		}

		if len(issues) > 0 {
			m.responder.Unauthorized(
				w, "INVALID_AUTH_HEADER",
				[]httpschema.Detail{{Field: "Authorization", Issues: issues}},
			)
			return
		}

		token := parts[1]
		userID, err := m.verifier.VerifyToken(r.Context(), token)
		if err != nil {
			m.responder.Unauthorized(
				w, "INVALID_TOKEN",
				[]httpschema.Detail{{Field: "Authorization", Issues: []string{"Invalid token"}}},
			)
			return
		}

		ctx := context.WithValue(r.Context(), httpschema.ContextKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
