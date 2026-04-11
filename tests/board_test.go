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

		// Register
		ac := CreateUserAndAuthenticateClient(t, httpClient, ts.URL)

		name := testutil.ValidBoardName().String()
		description := testutil.ValidBoardDescription().String()

		// Create
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

		// List
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

		// Get by id
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

		// Update
		updatedName := "Updated Board Name"
		updatedDescription := "Updated Board Description"
		updateResp := ac.Do(t, http.MethodPatch, "/v1/boards/"+bResp.ID, map[string]string{
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

		// Partial update: name only
		partialName := "Updated Name Only"
		updateNameOnlyResp := ac.Do(t, http.MethodPatch, "/v1/boards/"+bResp.ID, map[string]string{
			"name": partialName,
		})
		defer func() {
			_ = updateNameOnlyResp.Body.Close()
		}()
		if updateNameOnlyResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected partial update status %d, got %d", http.StatusOK, updateNameOnlyResp.StatusCode)
		}
		var updatedNameOnly boardJSON
		if err := json.NewDecoder(updateNameOnlyResp.Body).Decode(&updatedNameOnly); err != nil {
			t.Fatalf("Decode partial update response: %v", err)
		}
		if updatedNameOnly.Name != partialName {
			t.Errorf("Expected partial updated name %q, got %q", partialName, updatedNameOnly.Name)
		}
		if updatedNameOnly.Description != updatedDescription {
			t.Errorf("Expected description to stay %q, got %q", updatedDescription, updatedNameOnly.Description)
		}

		updateNullResp := ac.Do(t, http.MethodPatch, "/v1/boards/"+bResp.ID, map[string]any{
			"name":        nil,
			"description": nil,
		})
		defer func() {
			_ = updateNullResp.Body.Close()
		}()
		if updateNullResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected null update status %d, got %d", http.StatusOK, updateNullResp.StatusCode)
		}
		var updatedNull boardJSON
		if err := json.NewDecoder(updateNullResp.Body).Decode(&updatedNull); err != nil {
			t.Fatalf("Decode null update response: %v", err)
		}
		if updatedNull.Name != updatedNameOnly.Name || updatedNull.Description != updatedNameOnly.Description {
			t.Errorf("Expected null update to keep fields, got %+v", updatedNull)
		}

		// Get by id after update
		afterUpdateResp := ac.Do(t, http.MethodGet, "/v1/boards/"+bResp.ID, nil)
		defer func() {
			_ = afterUpdateResp.Body.Close()
		}()
		if afterUpdateResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected get-by-id status %d after update, got %d", http.StatusOK, afterUpdateResp.StatusCode)
		}

		var afterUpdate boardJSON
		if err := json.NewDecoder(afterUpdateResp.Body).Decode(&afterUpdate); err != nil {
			t.Fatalf("Decode get-by-id after update: %v", err)
		}
		if afterUpdate != updatedNull {
			t.Errorf("GET /v1/boards/{id} after update differs from final update response:\ngot  %+v\nwant %+v", afterUpdate, updatedNull)
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
		// TODO(refactor-1): use factor out to edge case, this is not a happy path
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

		// Delete
		delResp := ac.Do(t, http.MethodDelete, "/v1/boards/"+bResp.ID, nil)
		defer func() {
			_ = delResp.Body.Close()
		}()
		if delResp.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected delete status %d, got %d", http.StatusNoContent, delResp.StatusCode)
		}

		// List after delete
		afterDelResp := ac.Do(t, http.MethodGet, "/v1/boards/"+bResp.ID, nil)
		defer func() {
			_ = afterDelResp.Body.Close()
		}()
		if afterDelResp.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected 404 after delete, got %d", afterDelResp.StatusCode)
		}

		// Get by id after delete
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
