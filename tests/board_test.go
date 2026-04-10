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

type boardJSON struct {
	ID          string `json:"id"`
	OwnerID     string `json:"ownerId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func TestBoard_HappyPath(t *testing.T) {
	httpClient, ts, pool := Prelude(t)

	t.Run("Full board flow: register, login, create board, list boards, get by id", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "users")
		testutil.TruncateTable(t, pool, "boards")

		ac := CreateUserAndAuthenticateClient(t, httpClient, ts.URL)

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

		var bResp boardJSON
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
		if _, err := time.Parse(time.RFC3339, bResp.UpdatedAt); err != nil {
			t.Errorf("Invalid updatedAt: %v", err)
		}

		listResp := ac.Do(t, http.MethodGet, "/v1/boards", nil)
		defer func() {
			_ = listResp.Body.Close()
		}()

		if listResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected list status %d, got %d", http.StatusOK, listResp.StatusCode)
		}

		var listBody []boardJSON
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
		if _, err := time.Parse(time.RFC3339, item.UpdatedAt); err != nil {
			t.Errorf("Invalid list updatedAt: %v", err)
		}

		oneResp := ac.Do(t, http.MethodGet, "/v1/boards/"+bResp.ID, nil)
		defer func() {
			_ = oneResp.Body.Close()
		}()
		if oneResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected get-by-id status %d, got %d", http.StatusOK, oneResp.StatusCode)
		}
		var one boardJSON
		if err := json.NewDecoder(oneResp.Body).Decode(&one); err != nil {
			t.Fatalf("Decode get-by-id: %v", err)
		}
		if one != bResp {
			t.Errorf("GET /v1/boards/{id} body differs from create response:\ngot  %+v\nwant %+v", one, bResp)
		}

		updatedName := "Updated Board Name"
		updatedDescription := "Updated Board Description"
		updateResp := ac.Do(t, http.MethodPut, "/v1/boards/"+bResp.ID, map[string]string{
			"name":        updatedName,
			"description": updatedDescription,
		})
		defer func() {
			_ = updateResp.Body.Close()
		}()
		if updateResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected update status %d, got %d", http.StatusOK, updateResp.StatusCode)
		}

		var updated boardJSON
		if err := json.NewDecoder(updateResp.Body).Decode(&updated); err != nil {
			t.Fatalf("Decode update response: %v", err)
		}
		if updated.ID != bResp.ID {
			t.Errorf("Expected updated id %q, got %q", bResp.ID, updated.ID)
		}
		if updated.OwnerID != bResp.OwnerID {
			t.Errorf("Expected updated ownerId %q, got %q", bResp.OwnerID, updated.OwnerID)
		}
		if updated.Name != updatedName {
			t.Errorf("Expected updated name %q, got %q", updatedName, updated.Name)
		}
		if updated.Description != updatedDescription {
			t.Errorf("Expected updated description %q, got %q", updatedDescription, updated.Description)
		}

		oneAfterUpdateResp := ac.Do(t, http.MethodGet, "/v1/boards/"+bResp.ID, nil)
		defer func() {
			_ = oneAfterUpdateResp.Body.Close()
		}()
		if oneAfterUpdateResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected get-by-id status %d after update, got %d", http.StatusOK, oneAfterUpdateResp.StatusCode)
		}

		var oneAfterUpdate boardJSON
		if err := json.NewDecoder(oneAfterUpdateResp.Body).Decode(&oneAfterUpdate); err != nil {
			t.Fatalf("Decode get-by-id after update: %v", err)
		}
		if oneAfterUpdate != updated {
			t.Errorf("GET /v1/boards/{id} after update differs from update response:\ngot  %+v\nwant %+v", oneAfterUpdate, updated)
		}

		randomID := uuid.New().String()
		notFoundResp := ac.Do(t, http.MethodGet, "/v1/boards/"+randomID, nil)
		defer func() {
			_ = notFoundResp.Body.Close()
		}()
		if notFoundResp.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected 404 for unknown board, got %d", notFoundResp.StatusCode)
		}

		acOther := CreateUserAndAuthenticateClient(t, httpClient, ts.URL)
		crossResp := acOther.Do(t, http.MethodGet, "/v1/boards/"+bResp.ID, nil)
		defer func() {
			_ = crossResp.Body.Close()
		}()
		if crossResp.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected 404 when other user requests board by id, got %d", crossResp.StatusCode)
		}

		otherListResp := acOther.Do(t, http.MethodGet, "/v1/boards", nil)
		defer func() {
			_ = otherListResp.Body.Close()
		}()
		if otherListResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected list status %d for other user, got %d", http.StatusOK, otherListResp.StatusCode)
		}
		var otherList []boardJSON
		if err := json.NewDecoder(otherListResp.Body).Decode(&otherList); err != nil {
			t.Fatalf("Decode other user list: %v", err)
		}
		if len(otherList) != 0 {
			t.Fatalf("Expected other user to have 0 boards, got %d", len(otherList))
		}

		delResp := ac.Do(t, http.MethodDelete, "/v1/boards/"+bResp.ID, nil)
		defer func() {
			_ = delResp.Body.Close()
		}()
		if delResp.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected delete status %d, got %d", http.StatusNoContent, delResp.StatusCode)
		}

		afterDelResp := ac.Do(t, http.MethodGet, "/v1/boards/"+bResp.ID, nil)
		defer func() {
			_ = afterDelResp.Body.Close()
		}()
		if afterDelResp.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected 404 after delete, got %d", afterDelResp.StatusCode)
		}

		listAfterDel := ac.Do(t, http.MethodGet, "/v1/boards", nil)
		defer func() {
			_ = listAfterDel.Body.Close()
		}()
		if listAfterDel.StatusCode != http.StatusOK {
			t.Fatalf("Expected list status %d after delete, got %d", http.StatusOK, listAfterDel.StatusCode)
		}
		var listAfterBody []boardJSON
		if err := json.NewDecoder(listAfterDel.Body).Decode(&listAfterBody); err != nil {
			t.Fatalf("Decode list after delete: %v", err)
		}
		if len(listAfterBody) != 0 {
			t.Fatalf("Expected 0 boards after delete, got %d", len(listAfterBody))
		}
	})
}
