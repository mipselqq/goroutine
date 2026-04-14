//go:build e2e

package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"goroutine/internal/testutil"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

const timeFormat = "2006-01-02T15:04:05.000Z07:00"

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

	t.Run("Full board flow", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "users")
		testutil.TruncateTable(t, pool, "boards")

		// 1. Register (already done via ac client)
		ac := CreateUserAndAuthenticateClient(t, httpClient, ts.URL)

		name := testutil.ValidBoardName().String()
		description := testutil.ValidBoardDescription().String()

		// 2. Create a board, store response in createdBoard, and validate response fields
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

		createdBoard := parseBoard(t, createResp)

		if _, err := uuid.Parse(createdBoard.ID); err != nil {
			t.Errorf("Invalid board ID: %v", err)
		}
		if _, err := uuid.Parse(createdBoard.OwnerID); err != nil {
			t.Errorf("Invalid owner ID: %v", err)
		}
		if createdBoard.Name != name {
			t.Errorf("Expected name %q, got %q", name, createdBoard.Name)
		}
		if createdBoard.Description != description {
			t.Errorf("Expected description %q, got %q", description, createdBoard.Description)
		}
		if _, err := time.Parse(timeFormat, createdBoard.CreatedAt); err != nil {
			t.Errorf("Invalid createdAt: %v", err)
		}
		if _, err := time.Parse(timeFormat, createdBoard.UpdatedAt); err != nil {
			t.Errorf("Invalid updatedAt: %v", err)
		}

		// 3. Get by id, store response in getByIDBoard, and perform deep comparison with createdBoard
		oneResp := ac.Do(t, http.MethodGet, "/v1/boards/"+createdBoard.ID, nil)
		defer func() {
			_ = oneResp.Body.Close()
		}()
		if oneResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected Get by id status %d, got %d", http.StatusOK, oneResp.StatusCode)
		}

		getByIDBoard := parseBoard(t, oneResp)
		if diff := cmp.Diff(createdBoard, getByIDBoard); diff != "" {
			t.Errorf("Get by id mismatch (-want +got):\n%s", diff)
		}

		// 4. List boards, get the first board, and perform deep comparison with createdBoard
		listResp := ac.Do(t, http.MethodGet, "/v1/boards", nil)
		defer func() {
			_ = listResp.Body.Close()
		}()

		if listResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected list status %d, got %d", http.StatusOK, listResp.StatusCode)
		}

		listedBoards := parseBoardsList(t, listResp)
		if len(listedBoards) != 1 {
			t.Fatalf("Expected 1 board in list, got %d", len(listedBoards))
		}
		if diff := cmp.Diff(createdBoard, listedBoards[0]); diff != "" {
			t.Errorf("List item mismatch (-want +got):\n%s", diff)
		}

		// 5. Update by id with name only, store response in updatedNameBoard, validate fields, and ensure updatedAt advanced
		updatedName := "Updated Name Only"
		updateNameResp := ac.Do(t, http.MethodPatch, "/v1/boards/"+createdBoard.ID, map[string]string{
			"name": updatedName,
		})
		defer func() {
			_ = updateNameResp.Body.Close()
		}()
		if updateNameResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected Update by id status %d, got %d", http.StatusOK, updateNameResp.StatusCode)
		}

		updatedNameBoard := parseBoard(t, updateNameResp)

		// Validation trick: revert changed fields in a clone and compare with createdBoard
		checkBoard := updatedNameBoard
		checkBoard.Name = createdBoard.Name
		checkBoard.UpdatedAt = createdBoard.UpdatedAt
		if diff := cmp.Diff(createdBoard, checkBoard); diff != "" {
			t.Errorf("Update by id changed unexpected fields (-want +got):\n%s", diff)
		}

		// Verify specific changes
		if updatedNameBoard.Name != updatedName {
			t.Errorf("Expected updated name %q, got %q", updatedName, updatedNameBoard.Name)
		}

		updatedAtTime, _ := time.Parse(timeFormat, updatedNameBoard.UpdatedAt)
		baselineUpdatedAt, _ := time.Parse(timeFormat, createdBoard.UpdatedAt)
		if !updatedAtTime.After(baselineUpdatedAt) {
			t.Errorf("updatedAt must advance after Update by id; got %v, previous %v", updatedAtTime, baselineUpdatedAt)
		}

		// 6. Get by id again, store response in getByIDBoardAfterUpdate, and perform deep comparison with updatedNameBoard
		afterUpdateResp := ac.Do(t, http.MethodGet, "/v1/boards/"+createdBoard.ID, nil)
		defer func() {
			_ = afterUpdateResp.Body.Close()
		}()

		if afterUpdateResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected Get by id status %d after update, got %d", http.StatusOK, afterUpdateResp.StatusCode)
		}

		getByIDBoardAfterUpdate := parseBoard(t, afterUpdateResp)
		if diff := cmp.Diff(updatedNameBoard, getByIDBoardAfterUpdate); diff != "" {
			t.Errorf("Get by id after update mismatch (-want +got):\n%s", diff)
		}

		// 7. Delete by id and verify StatusNoContent
		delResp := ac.Do(t, http.MethodDelete, "/v1/boards/"+createdBoard.ID, nil)
		defer func() {
			_ = delResp.Body.Close()
		}()
		if delResp.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected Delete by id status %d, got %d", http.StatusNoContent, delResp.StatusCode)
		}

		// 8. List boards and ensure an empty list is returned
		listAfterDelResp := ac.Do(t, http.MethodGet, "/v1/boards", nil)
		defer func() {
			_ = listAfterDelResp.Body.Close()
		}()
		if listAfterDelResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected list status %d after delete, got %d", http.StatusOK, listAfterDelResp.StatusCode)
		}

		listedAfterDelete := parseBoardsList(t, listAfterDelResp)
		if len(listedAfterDelete) != 0 {
			t.Fatalf("Expected 0 boards after delete, got %d", len(listedAfterDelete))
		}
	})
}

func parseBoard(t *testing.T, resp *http.Response) boardJSON {
	t.Helper()
	var b boardJSON
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		t.Fatalf("Failed to decode board: %v", err)
	}
	return b
}

func parseBoardsList(t *testing.T, resp *http.Response) []boardJSON {
	t.Helper()
	var b []boardJSON
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		t.Fatalf("Failed to decode boards list: %v", err)
	}
	return b
}
