//go:build e2e

package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"goroutine/internal/testutil"

	"github.com/google/uuid"
)

func TestBoard_HappyPath(t *testing.T) {
	httpClient, ts, pool := Prelude(t)

	t.Run("Full board flow: register, login, create board, list boards", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "users")
		testutil.TruncateTable(t, pool, "boards")

		ac := NewAuthenticatedClient(t, httpClient, ts.URL)

		name := testutil.ValidBoardName().String()
		description := testutil.ValidBoardDescription().String()

		createResp := ac.Do(t, http.MethodPost, "/v1/boards", map[string]string{
			"name":        name,
			"description": description,
		})
		defer func() {
			_ = createResp.Body.Close()
		}()

		if createResp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d", http.StatusCreated, createResp.StatusCode)
		}

		var bResp struct {
			ID          string `json:"id"`
			OwnerID     string `json:"ownerId"`
			Name        string `json:"name"`
			Description string `json:"description"`
			CreatedAt   string `json:"createdAt"`
		}
		if err := json.NewDecoder(createResp.Body).Decode(&bResp); err != nil {
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

		listResp := ac.Do(t, http.MethodGet, "/v1/boards", nil)
		defer func() {
			_ = listResp.Body.Close()
		}()

		if listResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected list status %d, got %d", http.StatusOK, listResp.StatusCode)
		}

		var listBody []struct {
			ID          string `json:"id"`
			OwnerID     string `json:"ownerId"`
			Name        string `json:"name"`
			Description string `json:"description"`
			CreatedAt   string `json:"createdAt"`
		}
		if err := json.NewDecoder(listResp.Body).Decode(&listBody); err != nil {
			t.Fatalf("Failed to decode list response: %v", err)
		}

		if len(listBody) != 1 {
			t.Fatalf("Expected 1 board in list, got %d", len(listBody))
		}
		item := listBody[0]
		if item.ID != bResp.ID {
			t.Errorf("List id %q, create response id %q", item.ID, bResp.ID)
		}
		if item.OwnerID != bResp.OwnerID {
			t.Errorf("List ownerId %q, create ownerId %q", item.OwnerID, bResp.OwnerID)
		}
		if item.Name != name || item.Description != description {
			t.Errorf("List item name/description mismatch: got %+v", item)
		}
		if _, err := time.Parse(time.RFC3339, item.CreatedAt); err != nil {
			t.Errorf("Invalid list createdAt: %v", err)
		}
	})
}
