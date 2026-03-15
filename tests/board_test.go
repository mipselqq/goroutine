//go:build e2e

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"goroutine/internal/testutil"

	"github.com/google/uuid"
)

func TestBoard_HappyPath(t *testing.T) {
	client, ts, pool := E2EPrelude(t)

	t.Run("Create board successfully", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "users")
		testutil.TruncateTable(t, pool, "boards")

		email := testutil.ValidEmail().String()
		password := testutil.ValidPassword().String()

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

		name := testutil.ValidBoardName().String()
		description := testutil.ValidBoardDescription().String()
		boardBody, _ := json.Marshal(map[string]string{
			"name":        name,
			"description": description,
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

		var bResp struct {
			ID          string `json:"id"`
			OwnerID     string `json:"ownerId"`
			Name        string `json:"name"`
			Description string `json:"description"`
			CreatedAt   string `json:"createdAt"`
		}
		if err = json.NewDecoder(resp.Body).Decode(&bResp); err != nil {
			t.Fatalf("Failed to decode board response: %v", err)
		}

		if _, err := uuid.Parse(bResp.ID); err != nil {
			t.Errorf("Invalid board ID: %v", err)
		}
		if _, err := uuid.Parse(bResp.OwnerID); err != nil {
			t.Errorf("Invalid owner ID: %v", err)
		}
		if bResp.Name != name {
			t.Errorf("Expected name %q, got %q", name, bResp.Name)
		}
		if bResp.Description != description {
			t.Errorf("Expected description %q, got %q", description, bResp.Description)
		}
		if _, err := time.Parse(time.RFC3339, bResp.CreatedAt); err != nil {
			t.Errorf("Invalid createdAt: %v", err)
		}
	})
}
