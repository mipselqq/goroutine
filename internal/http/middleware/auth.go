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
	VerifyToken(ctx context.Context, token domain.AuthToken) (domain.UserID, error)
}

type Auth struct {
	logger    *slog.Logger
	verifier  TokenVerifier
	responder *httpschema.ErrorResponder
}

func NewAuth(logger *slog.Logger, verifier TokenVerifier, responder *httpschema.ErrorResponder) *Auth {
	return &Auth{logger: logger, verifier: verifier, responder: responder}
}

func (m *Auth) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			m.responder.InvalidAuthHeader(
				w, []httpschema.Detail{{Field: "Authorization", Issues: []string{"Missing authorization header"}}},
			)
			return
		}

		issues := []string{}
		authHeader := strings.TrimSpace(header)
		parts := strings.Fields(authHeader)

		if len(parts) != 2 {
			issues = append(issues, "Invalid authorization header")
		} else if !strings.EqualFold(parts[0], "bearer") {
			issues = append(issues, "No Bearer prefix")
		}

		if len(issues) > 0 {
			m.responder.InvalidAuthHeader(
				w, []httpschema.Detail{{Field: "Authorization", Issues: issues}},
			)
			return
		}

		token, err := domain.NewJWTString(parts[1])
		if err != nil {
			m.responder.InvalidToken(
				w, []httpschema.Detail{{Field: "Authorization", Issues: []string{"Invalid token"}}},
			)
			return
		}

		userID, err := m.verifier.VerifyToken(r.Context(), token)
		if err != nil {
			m.responder.InvalidToken(
				w, []httpschema.Detail{{Field: "Authorization", Issues: []string{"Invalid token"}}},
			)
			return
		}

		ctx := context.WithValue(r.Context(), httpschema.ContextKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
