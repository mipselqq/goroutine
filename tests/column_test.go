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

type columnPositionJSON struct {
	Position int64 `json:"position"`
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

		// 6. Create the second column so we can later delete it from the middle of the ordered list
		createSecondResp := ac.Do(t, http.MethodPost, "/v1/boards/"+board.ID+"/columns", map[string]string{
			"name": "In Progress",
		})
		defer func() {
			_ = createSecondResp.Body.Close()
		}()
		if createSecondResp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected second create status %d, got %d", http.StatusCreated, createSecondResp.StatusCode)
		}

		secondColumn := parseColumn(t, createSecondResp)
		if secondColumn.Position != 2 {
			t.Errorf("Expected second column position %d, got %d", 2, secondColumn.Position)
		}

		// 7. Create the third column so delete can verify position compaction for trailing items
		createThirdResp := ac.Do(t, http.MethodPost, "/v1/boards/"+board.ID+"/columns", map[string]string{
			"name": "Done",
		})
		defer func() {
			_ = createThirdResp.Body.Close()
		}()
		if createThirdResp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected third create status %d, got %d", http.StatusCreated, createThirdResp.StatusCode)
		}

		thirdColumn := parseColumn(t, createThirdResp)
		if thirdColumn.Position != 3 {
			t.Errorf("Expected third column position %d, got %d", 3, thirdColumn.Position)
		}

		// 8. Delete the middle column and expect the request to succeed
		deleteResp := ac.Do(t, http.MethodDelete, "/v1/boards/"+board.ID+"/columns/"+secondColumn.ID, nil)
		defer func() {
			_ = deleteResp.Body.Close()
		}()
		if deleteResp.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected delete status %d, got %d", http.StatusNoContent, deleteResp.StatusCode)
		}

		// 9. List columns again and verify delete preserved the first column and compacted positions
		listAfterDeleteResp := ac.Do(t, http.MethodGet, "/v1/boards/"+board.ID+"/columns", nil)
		defer func() {
			_ = listAfterDeleteResp.Body.Close()
		}()
		if listAfterDeleteResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected list status %d after delete, got %d", http.StatusOK, listAfterDeleteResp.StatusCode)
		}

		listedAfterDelete := parseColumnsList(t, listAfterDeleteResp)
		if len(listedAfterDelete) != 2 {
			t.Fatalf("Expected 2 columns after delete, got %d", len(listedAfterDelete))
		}
		if listedAfterDelete[0].ID != updatedNameColumn.ID || listedAfterDelete[0].Name != updatedNameColumn.Name || listedAfterDelete[0].Position != 1 {
			t.Errorf("Expected updated first column to stay at position 1, got id=%s name=%q position=%d", listedAfterDelete[0].ID, listedAfterDelete[0].Name, listedAfterDelete[0].Position)
		}
		if listedAfterDelete[1].ID != thirdColumn.ID || listedAfterDelete[1].Position != 2 {
			t.Errorf("Expected third column to shift to position 2, got id=%s position=%d", listedAfterDelete[1].ID, listedAfterDelete[1].Position)
		}

		// 10. Move the remaining second column to the first position and expect the new position in response
		moveResp := ac.Do(t, http.MethodPut, "/v1/boards/"+board.ID+"/columns/"+thirdColumn.ID+"/position", map[string]int64{
			"targetPosition": 1,
		})
		defer func() {
			_ = moveResp.Body.Close()
		}()
		if moveResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected move status %d, got %d", http.StatusOK, moveResp.StatusCode)
		}

		movedPosition := parseColumnPosition(t, moveResp)
		if movedPosition.Position != 1 {
			t.Errorf("Expected move response position %d, got %d", 1, movedPosition.Position)
		}

		// 11. List columns again and verify move reordered the remaining columns
		listAfterMoveResp := ac.Do(t, http.MethodGet, "/v1/boards/"+board.ID+"/columns", nil)
		defer func() {
			_ = listAfterMoveResp.Body.Close()
		}()
		if listAfterMoveResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected list status %d after move, got %d", http.StatusOK, listAfterMoveResp.StatusCode)
		}

		listedAfterMove := parseColumnsList(t, listAfterMoveResp)
		if len(listedAfterMove) != 2 {
			t.Fatalf("Expected 2 columns after move, got %d", len(listedAfterMove))
		}
		if listedAfterMove[0].ID != thirdColumn.ID || listedAfterMove[0].Position != 1 {
			t.Errorf("Expected third column to move to position 1, got id=%s position=%d", listedAfterMove[0].ID, listedAfterMove[0].Position)
		}
		if listedAfterMove[1].ID != updatedNameColumn.ID || listedAfterMove[1].Position != 2 {
			t.Errorf("Expected updated first column to shift to position 2, got id=%s position=%d", listedAfterMove[1].ID, listedAfterMove[1].Position)
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

func parseColumnPosition(t *testing.T, resp *http.Response) columnPositionJSON {
	t.Helper()
	var p columnPositionJSON
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatalf("decode column position: %v", err)
	}
	return p
}
