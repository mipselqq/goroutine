package middleware

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"goroutine/internal/secrecy"

	"github.com/golang-jwt/jwt/v5"
)

func createTestToken(t *testing.T, exp int64, secret secrecy.SecretString, signingMethod jwt.SigningMethod) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub": "test@example.com",
		"exp": exp,
		"iat": time.Now().Unix(),
	}
	unsigned := jwt.NewWithClaims(signingMethod, claims)
	signedToken, err := unsigned.SignedString([]byte(secret))
	if err != nil {
		t.Errorf("BUG: failed to sign valid token")
	}
	return signedToken
}

func createTestTokenES256(t *testing.T, exp int64) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	claims := jwt.MapClaims{
		"sub": "test@example.com",
		"exp": exp,
		"iat": time.Now().Unix(),
	}
	unsigned := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token, err := unsigned.SignedString(key)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return token
}

func Bearer(token string) string {
	return fmt.Sprintf("Bearer %s", token)
}

func TestAuthMiddleware(t *testing.T) {
	t.Parallel()

	signMethod := jwt.SigningMethodHS256
	secret := secrecy.SecretString("secret")
	invalidSecret := secrecy.SecretString("we-will-hackyou")
	hourExp := time.Now().Add(time.Hour).Unix()
	expiredExp := time.Now().Add(-time.Hour).Unix()

	validToken := createTestToken(t, hourExp, secret, signMethod)
	expiredToken := createTestToken(t, expiredExp, secret, signMethod)
	invalidSignMethodToken := createTestTokenES256(t, hourExp)
	invalidSignSecretToken := createTestToken(t, hourExp, invalidSecret, signMethod)
	invalidToken := "how too cook pelmeni?"

	tests := []struct {
		name           string
		expectedStatus int
		authHeader     string
	}{
		{
			name:           "Success",
			expectedStatus: http.StatusTeapot,
			authHeader:     Bearer(validToken),
		},
		{
			name:           "Expired",
			expectedStatus: http.StatusUnauthorized,
			authHeader:     Bearer(expiredToken),
		},
		{
			name:           "Invalid Sign Method",
			expectedStatus: http.StatusUnauthorized,
			authHeader:     Bearer(invalidSignMethodToken),
		},
		{
			name:           "Invalid Sign Secret",
			expectedStatus: http.StatusUnauthorized,
			authHeader:     Bearer(invalidSignSecretToken),
		},
		{
			name:           "Invalid Token",
			expectedStatus: http.StatusUnauthorized,
			authHeader:     Bearer(invalidToken),
		},
		{
			name:           "No Authorization Header",
			expectedStatus: http.StatusUnauthorized,
			authHeader:     "",
		},
	}

	const expectedUserID = "test@example.com"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var gotUserID string
			handler := func(w http.ResponseWriter, r *http.Request) {
				gotUserID, _ = UserIDFromContext(r.Context())
				w.WriteHeader(http.StatusTeapot)
			}

			mw := NewAuth(secret)
			wrapped := mw.Wrap(http.HandlerFunc(handler))
			req, _ := http.NewRequest("GET", "/", http.NoBody)
			req.Header.Set("Authorization", tt.authHeader)
			rr := httptest.NewRecorder()

			wrapped.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
			if tt.name == "Success" && gotUserID != expectedUserID {
				t.Errorf("handler context: expected user ID %q, got %q", expectedUserID, gotUserID)
			}
		})
	}
}
