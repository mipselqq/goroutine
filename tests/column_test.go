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

type ColumnJSON struct {
	ID          string `json:"id"`
	BoardID     string `json:"boardId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Position    int64  `json:"position"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type ColumnPositionJSON struct {
	Position int64 `json:"position"`
}

func TestColumn_HappyPath(t *testing.T) {
	p := Prelude(t)

	testutil.TruncateAllTables(t, p.Pool)

	// 1. Register (already done via ac client)
	ac := CreateUserAndAuthenticateClient(t, p.HTTPClient, p.Server.URL)

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
		t.Fatalf("got create board status %d, want %d", createBoardResp.StatusCode, http.StatusCreated)
	}

	board := parseBoard(t, createBoardResp)

	name := "To Do"
	columnDescription := testutil.ValidColumnDescription().String()

	// 2. Create a column, store response in createdColumn, and validate response fields
	createResp := ac.Do(t, http.MethodPost, "/v1/boards/"+board.ID+"/columns", map[string]string{
		"name":        name,
		"description": columnDescription,
	})
	defer func() {
		_ = createResp.Body.Close()
	}()
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("got status %d, want %d", createResp.StatusCode, http.StatusCreated)
	}

	createdColumn := parseColumn(t, createResp)

	_, err := uuid.Parse(createdColumn.ID)
	if err != nil {
		t.Errorf("uuid.Parse(%q) error = %v, want nil", createdColumn.ID, err)
	}
	_, err = uuid.Parse(createdColumn.BoardID)
	if err != nil {
		t.Errorf("uuid.Parse(%q) error = %v, want nil", createdColumn.BoardID, err)
	}
	if createdColumn.BoardID != board.ID {
		t.Errorf("got board ID %q, want %q", createdColumn.BoardID, board.ID)
	}
	if createdColumn.Name != name {
		t.Errorf("got name %q, want %q", createdColumn.Name, name)
	}
	if createdColumn.Description != columnDescription {
		t.Errorf("got description %q, want %q", createdColumn.Description, columnDescription)
	}
	if createdColumn.Position != 1 {
		t.Errorf("got position %d, want %d", createdColumn.Position, 1)
	}
	_, err = time.Parse(timeFormat, createdColumn.CreatedAt)
	if err != nil {
		t.Errorf("time.Parse(%q) error = %v, want nil", createdColumn.CreatedAt, err)
	}
	_, err = time.Parse(timeFormat, createdColumn.UpdatedAt)
	if err != nil {
		t.Errorf("time.Parse(%q) error = %v, want nil", createdColumn.UpdatedAt, err)
	}

	// 3. List columns, get the first column, and perform deep comparison with createdColumn
	listResp := ac.Do(t, http.MethodGet, "/v1/boards/"+board.ID+"/columns", nil)
	defer func() {
		_ = listResp.Body.Close()
	}()
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("got list status %d, want %d", listResp.StatusCode, http.StatusOK)
	}

	listedColumns := parseColumnsList(t, listResp)
	if len(listedColumns) != 1 {
		t.Fatalf("got %d columns in list, want 1", len(listedColumns))
	}
	if diff := cmp.Diff(createdColumn, listedColumns[0]); diff != "" {
		t.Errorf("List item mismatch (-want +got):\n%s", diff)
	}

	// 4. Update by id with name only, store response in updatedNameColumn, validate fields, and ensure updatedAt advanced
	updatedName := "Updated Name Only"
	WaitForTimestampTicker(t)
	updateNameResp := ac.Do(t, http.MethodPatch, "/v1/boards/"+board.ID+"/columns/"+createdColumn.ID, map[string]string{
		"name": updatedName,
	})
	defer func() {
		_ = updateNameResp.Body.Close()
	}()
	if updateNameResp.StatusCode != http.StatusOK {
		t.Fatalf("got Update by id status %d, want %d", updateNameResp.StatusCode, http.StatusOK)
	}

	updatedNameColumn := parseColumn(t, updateNameResp)

	// Validation trick: revert changed fields in a clone and compare with createdColumn
	checkColumn := updatedNameColumn
	checkColumn.Name = createdColumn.Name
	checkColumn.UpdatedAt = createdColumn.UpdatedAt
	if diff := cmp.Diff(createdColumn, checkColumn); diff != "" {
		t.Errorf("Update() diff (-want +got):\n%s", diff)
	}

	// Verify specific changes
	if updatedNameColumn.Name != updatedName {
		t.Errorf("got updated name %q, want %q", updatedNameColumn.Name, updatedName)
	}
	if updatedNameColumn.Description != createdColumn.Description {
		t.Errorf("got description %q after name-only update, want %q", updatedNameColumn.Description, createdColumn.Description)
	}

	updatedDesc := "Updated column description body"
	WaitForTimestampTicker(t)
	updateDescResp := ac.Do(t, http.MethodPatch, "/v1/boards/"+board.ID+"/columns/"+createdColumn.ID, map[string]string{
		"description": updatedDesc,
	})
	defer func() {
		_ = updateDescResp.Body.Close()
	}()
	if updateDescResp.StatusCode != http.StatusOK {
		t.Fatalf("got Update description status %d, want %d", updateDescResp.StatusCode, http.StatusOK)
	}
	updatedDescColumn := parseColumn(t, updateDescResp)
	if updatedDescColumn.Description != updatedDesc {
		t.Errorf("got description %q, want %q", updatedDescColumn.Description, updatedDesc)
	}
	if updatedDescColumn.Name != updatedName {
		t.Errorf("got name %q after description update, want %q", updatedDescColumn.Name, updatedName)
	}

	updatedAtAfterName, err := time.Parse(timeFormat, updatedNameColumn.UpdatedAt)
	if err != nil {
		t.Fatalf("time.Parse(%q) error = %v", updatedNameColumn.UpdatedAt, err)
	}
	updatedAtBeforeUpdate, err := time.Parse(timeFormat, createdColumn.UpdatedAt)
	if err != nil {
		t.Fatalf("time.Parse(%q) error = %v", createdColumn.UpdatedAt, err)
	}
	if !updatedAtAfterName.After(updatedAtBeforeUpdate) {
		t.Errorf("updatedAt must advance after name-only update; got %v, previous %v", updatedAtAfterName, updatedAtBeforeUpdate)
	}

	updatedAtAfterDesc, err := time.Parse(timeFormat, updatedDescColumn.UpdatedAt)
	if err != nil {
		t.Fatalf("time.Parse(%q) error = %v", updatedDescColumn.UpdatedAt, err)
	}
	if !updatedAtAfterDesc.After(updatedAtAfterName) {
		t.Errorf("updatedAt must advance after description update; got %v, previous %v", updatedAtAfterDesc, updatedAtAfterName)
	}

	// 5. List columns again, get the first column, and perform deep comparison with updatedDescColumn
	listAfterUpdateResp := ac.Do(t, http.MethodGet, "/v1/boards/"+board.ID+"/columns", nil)
	defer func() {
		_ = listAfterUpdateResp.Body.Close()
	}()
	if listAfterUpdateResp.StatusCode != http.StatusOK {
		t.Fatalf("got list status %d after update, want %d", listAfterUpdateResp.StatusCode, http.StatusOK)
	}

	listedAfterUpdate := parseColumnsList(t, listAfterUpdateResp)
	if len(listedAfterUpdate) != 1 {
		t.Fatalf("got %d columns after update, want 1", len(listedAfterUpdate))
	}
	if diff := cmp.Diff(updatedDescColumn, listedAfterUpdate[0]); diff != "" {
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
		t.Fatalf("got second create status %d, want %d", createSecondResp.StatusCode, http.StatusCreated)
	}

	secondColumn := parseColumn(t, createSecondResp)
	if secondColumn.Position != 2 {
		t.Errorf("got second column position %d, want %d", secondColumn.Position, 2)
	}

	// 7. Create the third column so delete can verify position compaction for trailing items
	createThirdResp := ac.Do(t, http.MethodPost, "/v1/boards/"+board.ID+"/columns", map[string]string{
		"name": "Done",
	})
	defer func() {
		_ = createThirdResp.Body.Close()
	}()
	if createThirdResp.StatusCode != http.StatusCreated {
		t.Fatalf("got third create status %d, want %d", createThirdResp.StatusCode, http.StatusCreated)
	}

	thirdColumn := parseColumn(t, createThirdResp)
	if thirdColumn.Position != 3 {
		t.Errorf("got third column position %d, want %d", thirdColumn.Position, 3)
	}

	// 8. Delete the middle column and expect the request to succeed
	deleteResp := ac.Do(t, http.MethodDelete, "/v1/boards/"+board.ID+"/columns/"+secondColumn.ID, nil)
	defer func() {
		_ = deleteResp.Body.Close()
	}()
	if deleteResp.StatusCode != http.StatusNoContent {
		t.Fatalf("got delete status %d, want %d", deleteResp.StatusCode, http.StatusNoContent)
	}

	// 9. List columns again and verify delete preserved the first column and compacted positions
	listAfterDeleteResp := ac.Do(t, http.MethodGet, "/v1/boards/"+board.ID+"/columns", nil)
	defer func() {
		_ = listAfterDeleteResp.Body.Close()
	}()
	if listAfterDeleteResp.StatusCode != http.StatusOK {
		t.Fatalf("got list status %d after delete, want %d", listAfterDeleteResp.StatusCode, http.StatusOK)
	}

	listedAfterDelete := parseColumnsList(t, listAfterDeleteResp)
	if len(listedAfterDelete) != 2 {
		t.Fatalf("got %d columns after delete, want 2", len(listedAfterDelete))
	}
	if listedAfterDelete[0].ID != updatedDescColumn.ID || listedAfterDelete[0].Name != updatedDescColumn.Name || listedAfterDelete[0].Position != 1 {
		t.Errorf("got id=%s name=%q position=%d, want id=%s name=%q position=%d", listedAfterDelete[0].ID, listedAfterDelete[0].Name, listedAfterDelete[0].Position, updatedDescColumn.ID, updatedDescColumn.Name, 1)
	}
	if listedAfterDelete[1].ID != thirdColumn.ID || listedAfterDelete[1].Position != 2 {
		t.Errorf("got id=%s position=%d, want id=%s position=%d", listedAfterDelete[1].ID, listedAfterDelete[1].Position, thirdColumn.ID, 2)
	}

	// 10. Move the remaining second column to the first position and expect the new position in response
	moveResp := ac.Do(t, http.MethodPut, "/v1/boards/"+board.ID+"/columns/"+thirdColumn.ID+"/position", map[string]int64{
		"targetPosition": 1,
	})
	defer func() {
		_ = moveResp.Body.Close()
	}()
	if moveResp.StatusCode != http.StatusOK {
		t.Fatalf("got move status %d, want %d", moveResp.StatusCode, http.StatusOK)
	}

	movedPosition := parseColumnPosition(t, moveResp)
	if movedPosition.Position != 1 {
		t.Errorf("got move response position %d, want %d", movedPosition.Position, 1)
	}

	// 11. List columns again and verify move reordered the remaining columns
	listAfterMoveResp := ac.Do(t, http.MethodGet, "/v1/boards/"+board.ID+"/columns", nil)
	defer func() {
		_ = listAfterMoveResp.Body.Close()
	}()
	if listAfterMoveResp.StatusCode != http.StatusOK {
		t.Fatalf("got list status %d after move, want %d", listAfterMoveResp.StatusCode, http.StatusOK)
	}

	listedAfterMove := parseColumnsList(t, listAfterMoveResp)
	if len(listedAfterMove) != 2 {
		t.Fatalf("got %d columns after move, want 2", len(listedAfterMove))
	}
	if listedAfterMove[0].ID != thirdColumn.ID || listedAfterMove[0].Position != 1 {
		t.Errorf("got id=%s position=%d, want id=%s position=%d", listedAfterMove[0].ID, listedAfterMove[0].Position, thirdColumn.ID, 1)
	}
	if listedAfterMove[1].ID != updatedDescColumn.ID || listedAfterMove[1].Position != 2 {
		t.Errorf("got id=%s position=%d, want id=%s position=%d", listedAfterMove[1].ID, listedAfterMove[1].Position, updatedDescColumn.ID, 2)
	}
}

func parseColumn(t *testing.T, resp *http.Response) ColumnJSON {
	t.Helper()
	var c ColumnJSON
	err := json.NewDecoder(resp.Body).Decode(&c)
	if err != nil {
		t.Fatalf("Column Decode() error = %v", err)
	}
	return c
}

func parseColumnsList(t *testing.T, resp *http.Response) []ColumnJSON {
	t.Helper()
	var c []ColumnJSON
	err := json.NewDecoder(resp.Body).Decode(&c)
	if err != nil {
		t.Fatalf("Columns list Decode() error = %v", err)
	}
	return c
}

func parseColumnPosition(t *testing.T, resp *http.Response) ColumnPositionJSON {
	t.Helper()
	var p ColumnPositionJSON
	err := json.NewDecoder(resp.Body).Decode(&p)
	if err != nil {
		t.Fatalf("Column position Decode() error = %v", err)
	}
	return p
}
