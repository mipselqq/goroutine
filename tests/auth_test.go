//go:build e2e

package tests

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goroutine/internal/app"
	"goroutine/internal/config"
	"goroutine/internal/testutil"
)

func TestAuth_HappyPath(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../migrations")
	defer pool.Close()

	cfg := config.NewAppConfigFromEnv()
	logger := testutil.CreateTestLogger(t)
	logger.Info("App config", slog.Any("config", cfg))

	application := app.New(logger, pool, &cfg)

	ts := httptest.NewServer(application.Router)
	defer ts.Close()
	client := ts.Client()

	t.Run("Full auth flow: register then login", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "users")

		email := "e2e-user@example.com"
		password := "very-strong-password"

		regBody, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})

		regResp, err := client.Post(ts.URL+"/register", "application/json", bytes.NewBuffer(regBody))
		if err != nil {
			t.Fatalf("Register request failed: %v", err)
		}
		if regResp.StatusCode != http.StatusOK {
			t.Errorf("Expeted register status 200, got %d", regResp.StatusCode)
		}
		_ = regResp.Body.Close()

		loginBody, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})

		loginResp, err := client.Post(ts.URL+"/login", "application/json", bytes.NewBuffer(loginBody))
		if err != nil {
			t.Fatalf("Login request failed: %v", err)
		}
		defer func() {
			_ = loginResp.Body.Close() // Calm down errcheck
		}()

		if loginResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected login status 200, got %d", loginResp.StatusCode)
		}

		var respBody struct {
			Token string `json:"token"`
		}
		err = json.NewDecoder(loginResp.Body).Decode(&respBody)
		if err != nil {
			t.Fatalf("Failed to decode login response: %v", err)
		}

		// TODO: validate token by sending a request to some protected endpoint
		parts := strings.Split(respBody.Token, ".")
		if len(parts) != 3 {
			t.Fatal("Got invalid JWT token")
		}
	})
}
