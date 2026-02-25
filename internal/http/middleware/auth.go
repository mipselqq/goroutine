package middleware

import (
	"context"
	"net/http"
	"strings"

	"goroutine/internal/secrecy"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "user_id"

func UserIDFromContext(ctx context.Context) (userID string, ok bool) {
	v := ctx.Value(userIDKey)
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

type Auth struct {
	secret secrecy.SecretString
}

func NewAuth(secret secrecy.SecretString) *Auth {
	return &Auth{secret: secret}
}

func (m *Auth) Wrap(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if tokenStr == r.Header.Get("Authorization") {
			http.Error(w, "missing or invalid authorization", http.StatusUnauthorized)
			return
		}

		// TODO: use service.VerifyToken instead
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.secret.RevealSecret()), nil
		})
		if err != nil || !token.Valid {
			// TODO: use responders from handler package
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		sub, _ := claims["sub"].(string)
		if sub == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, sub)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
