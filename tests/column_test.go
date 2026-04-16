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

type columnJSON struct {
	ID        string `json:"id"`
	BoardID   string `json:"boardId"`
	Name      string `json:"name"`
	Position  int64  `json:"position"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func TestColumn_HappyPath(t *testing.T) {
	httpClient, ts, pool := Prelude(t)

	t.Run("Full column flow", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

		// 1. Register (already done via ac client)
		ac := CreateUserAndAuthenticateClient(t, httpClient, ts.URL)

		boardName := testutil.ValidBoardName().String()
		boardDescription := testutil.ValidBoardDescription().String()

		createBoardResp := ac.Do(t, http.MethodPost, "/v1/boards", map[string]string{
			"name":        boardName,
			"description": boardDescription,
		})
		defer func() {
			_ = createBoardResp.Body.Close()
		}()
		if createBoardResp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected create board status %d, got %d", http.StatusCreated, createBoardResp.StatusCode)
		}

		board := parseBoard(t, createBoardResp)

		name := "To Do"

		// 2. Create a column, store response in createdColumn, and validate response fields
		createResp := ac.Do(t, http.MethodPost, "/v1/boards/"+board.ID+"/columns", map[string]string{
			"name": name,
		})
		defer func() {
			_ = createResp.Body.Close()
		}()
		if createResp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d", http.StatusCreated, createResp.StatusCode)
		}

		createdColumn := parseColumn(t, createResp)

		if _, err := uuid.Parse(createdColumn.ID); err != nil {
			t.Errorf("Invalid column ID: %v", err)
		}
		if _, err := uuid.Parse(createdColumn.BoardID); err != nil {
			t.Errorf("Invalid board ID: %v", err)
		}
		if createdColumn.BoardID != board.ID {
			t.Errorf("Expected board ID %q, got %q", board.ID, createdColumn.BoardID)
		}
		if createdColumn.Name != name {
			t.Errorf("Expected name %q, got %q", name, createdColumn.Name)
		}
		if createdColumn.Position != 1 {
			t.Errorf("Expected position %d, got %d", 1, createdColumn.Position)
		}
		if _, err := time.Parse(timeFormat, createdColumn.CreatedAt); err != nil {
			t.Errorf("Invalid createdAt: %v", err)
		}
		if _, err := time.Parse(timeFormat, createdColumn.UpdatedAt); err != nil {
			t.Errorf("Invalid updatedAt: %v", err)
		}

		// 3. List columns, get the first column, and perform deep comparison with createdColumn
		listResp := ac.Do(t, http.MethodGet, "/v1/boards/"+board.ID+"/columns", nil)
		defer func() {
			_ = listResp.Body.Close()
		}()
		if listResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected list status %d, got %d", http.StatusOK, listResp.StatusCode)
		}

		listedColumns := parseColumnsList(t, listResp)
		if len(listedColumns) != 1 {
			t.Fatalf("Expected 1 column in list, got %d", len(listedColumns))
		}
		if diff := cmp.Diff(createdColumn, listedColumns[0]); diff != "" {
			t.Errorf("List item mismatch (-want +got):\n%s", diff)
		}

		// 4. Update by id with name only, store response in updatedNameColumn, validate fields, and ensure updatedAt advanced
		updatedName := "Updated Name Only"
		updateNameResp := ac.Do(t, http.MethodPatch, "/v1/boards/"+board.ID+"/columns/"+createdColumn.ID, map[string]string{
			"name": updatedName,
		})
		defer func() {
			_ = updateNameResp.Body.Close()
		}()
		if updateNameResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected Update by id status %d, got %d", http.StatusOK, updateNameResp.StatusCode)
		}

		updatedNameColumn := parseColumn(t, updateNameResp)

		// Validation trick: revert changed fields in a clone and compare with createdColumn
		checkColumn := updatedNameColumn
		checkColumn.Name = createdColumn.Name
		checkColumn.UpdatedAt = createdColumn.UpdatedAt
		if diff := cmp.Diff(createdColumn, checkColumn); diff != "" {
			t.Errorf("Update by id changed unexpected fields (-want +got):\n%s", diff)
		}

		// Verify specific changes
		if updatedNameColumn.Name != updatedName {
			t.Errorf("Expected updated name %q, got %q", updatedName, updatedNameColumn.Name)
		}

		updatedAtTime, _ := time.Parse(timeFormat, updatedNameColumn.UpdatedAt)
		baselineUpdatedAt, _ := time.Parse(timeFormat, createdColumn.UpdatedAt)
		if !updatedAtTime.After(baselineUpdatedAt) {
			t.Errorf("updatedAt must advance after Update by id; got %v, previous %v", updatedAtTime, baselineUpdatedAt)
		}

		// 5. List columns again, get the first column, and perform deep comparison with updatedNameColumn
		listAfterUpdateResp := ac.Do(t, http.MethodGet, "/v1/boards/"+board.ID+"/columns", nil)
		defer func() {
			_ = listAfterUpdateResp.Body.Close()
		}()
		if listAfterUpdateResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected list status %d after update, got %d", http.StatusOK, listAfterUpdateResp.StatusCode)
		}

		listedAfterUpdate := parseColumnsList(t, listAfterUpdateResp)
		if len(listedAfterUpdate) != 1 {
			t.Fatalf("Expected 1 column after update, got %d", len(listedAfterUpdate))
		}
		if diff := cmp.Diff(updatedNameColumn, listedAfterUpdate[0]); diff != "" {
			t.Errorf("List item after update mismatch (-want +got):\n%s", diff)
		}
	})
}

func parseColumn(t *testing.T, resp *http.Response) columnJSON {
	t.Helper()
	var c columnJSON
	if err := json.NewDecoder(resp.Body).Decode(&c); err != nil {
		t.Fatalf("decode column: %v", err)
	}
	return c
}

func parseColumnsList(t *testing.T, resp *http.Response) []columnJSON {
	t.Helper()
	var c []columnJSON
	if err := json.NewDecoder(resp.Body).Decode(&c); err != nil {
		t.Fatalf("decode columns list: %v", err)
	}
	return c
}
