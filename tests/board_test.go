//go:build e2e

package tests

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"goroutine/internal/app"
	"goroutine/internal/config"
	"goroutine/internal/testutil"

	"github.com/prometheus/client_golang/prometheus"
)

func TestBoard_HappyPath(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../migrations")
	defer pool.Close()

	logger := testutil.NewTestLogger(t)
	cfg := config.NewAppConfigFromEnv(logger)
	logger.Info("App config", slog.Any("config", cfg))

	application := app.New(logger, pool, &cfg, prometheus.NewRegistry())

	ts := httptest.NewServer(application.Router)
	defer ts.Close()
	client := ts.Client()

	t.Run("Create board successfully", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "users")
		testutil.TruncateTable(t, pool, "boards")

		email := "board-user@example.com"
		password := "secure-password"

		regBody, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})
		regResp, err := client.Post(ts.URL+"/v1/register", "application/json", bytes.NewBuffer(regBody))
		if err != nil {
			t.Fatalf("Register failed: %v", err)
		}
		_ = regResp.Body.Close()

		loginBody, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})
		loginResp, err := client.Post(ts.URL+"/v1/login", "application/json", bytes.NewBuffer(loginBody))
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}

		var lResp struct {
			Token string `json:"token"`
		}
		if err = json.NewDecoder(loginResp.Body).Decode(&lResp); err != nil {
			t.Fatalf("Failed to decode token: %v", err)
		}
		_ = loginResp.Body.Close()

		boardBody, _ := json.Marshal(map[string]string{
			"name":        "My test board",
			"description": "This is a board description.",
		})
		req, err := http.NewRequest("POST", ts.URL+"/v1/boards", bytes.NewBuffer(boardBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+lResp.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Create board request failed: %v", err)
		}
		defer func() {
			_ = resp.Body.Close() // Calm down errcheck
		}()

		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
		}
	})
}
