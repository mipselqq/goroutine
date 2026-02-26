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

	"github.com/google/uuid"
)

func TestAuth_HappyPath(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../migrations")
	defer pool.Close()

	logger := testutil.NewTestLogger(t)
	cfg := config.NewAppConfigFromEnv(logger)
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

		parts := strings.Split(respBody.Token, ".")
		if len(parts) != 3 {
			t.Fatal("Got invalid JWT token")
		}

		req, err := http.NewRequest("GET", ts.URL+"/whoami", nil)
		if err != nil {
			t.Fatalf("Failed to create whoami request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+respBody.Token)

		whoamiResp, err := client.Do(req)
		if err != nil {
			t.Fatalf("WhoAmI request failed: %v", err)
		}
		defer whoamiResp.Body.Close()

		if whoamiResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", whoamiResp.StatusCode)
		}

		var whoamiData struct {
			UID string `json:"uid"`
		}
		if err := json.NewDecoder(whoamiResp.Body).Decode(&whoamiData); err != nil {
			t.Fatalf("Failed to decode whoami response: %v", err)
		}

		if _, err := uuid.Parse(whoamiData.UID); err != nil {
			t.Errorf("Expected valid UUID user ID, got %s: %v", whoamiData.UID, err)
		}
	})
}
