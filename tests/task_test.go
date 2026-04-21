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

type taskJSON struct {
	ID          string `json:"id"`
	ColumnID    string `json:"columnId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Position    int64  `json:"position"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type taskPositionJSON struct {
	ColumnID string `json:"columnId"`
	Position int64  `json:"position"`
}

func TestTask_HappyPath(t *testing.T) {
	httpClient, ts, pool := Prelude(t)

	t.Run("Full task flow", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

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
			t.Fatalf("got create board status %d, want %d", createBoardResp.StatusCode, http.StatusCreated)
		}

		board := parseBoard(t, createBoardResp)

		// 2. Create two columns so we can later move a task across them
		createColumnAResp := ac.Do(t, http.MethodPost, "/v1/boards/"+board.ID+"/columns", map[string]string{
			"name": "To Do",
		})
		defer func() {
			_ = createColumnAResp.Body.Close()
		}()
		if createColumnAResp.StatusCode != http.StatusCreated {
			t.Fatalf("got create column A status %d, want %d", createColumnAResp.StatusCode, http.StatusCreated)
		}
		columnA := parseColumn(t, createColumnAResp)

		createColumnBResp := ac.Do(t, http.MethodPost, "/v1/boards/"+board.ID+"/columns", map[string]string{
			"name": "In Progress",
		})
		defer func() {
			_ = createColumnBResp.Body.Close()
		}()
		if createColumnBResp.StatusCode != http.StatusCreated {
			t.Fatalf("got create column B status %d, want %d", createColumnBResp.StatusCode, http.StatusCreated)
		}
		columnB := parseColumn(t, createColumnBResp)

		name := "Write tests"
		description := "Cover the new endpoint with tests"

		// 3. Create a task in column A, store response in createdTask, and validate response fields
		createResp := ac.Do(t, http.MethodPost, "/v1/boards/"+board.ID+"/columns/"+columnA.ID+"/tasks", map[string]string{
			"name":        name,
			"description": description,
		})
		defer func() {
			_ = createResp.Body.Close()
		}()
		if createResp.StatusCode != http.StatusCreated {
			t.Fatalf("got status %d, want %d", createResp.StatusCode, http.StatusCreated)
		}

		createdTask := parseTask(t, createResp)

		if _, err := uuid.Parse(createdTask.ID); err != nil {
			t.Errorf("uuid.Parse(%q) error = %v, want nil", createdTask.ID, err)
		}
		if _, err := uuid.Parse(createdTask.ColumnID); err != nil {
			t.Errorf("uuid.Parse(%q) error = %v, want nil", createdTask.ColumnID, err)
		}
		if createdTask.ColumnID != columnA.ID {
			t.Errorf("got column ID %q, want %q", createdTask.ColumnID, columnA.ID)
		}
		if createdTask.Name != name {
			t.Errorf("got name %q, want %q", createdTask.Name, name)
		}
		if createdTask.Description != description {
			t.Errorf("got description %q, want %q", createdTask.Description, description)
		}
		if createdTask.Position != 1 {
			t.Errorf("got position %d, want %d", createdTask.Position, 1)
		}
		if _, err := time.Parse(timeFormat, createdTask.CreatedAt); err != nil {
			t.Errorf("time.Parse(%q) error = %v, want nil", createdTask.CreatedAt, err)
		}
		if _, err := time.Parse(timeFormat, createdTask.UpdatedAt); err != nil {
			t.Errorf("time.Parse(%q) error = %v, want nil", createdTask.UpdatedAt, err)
		}

		// 4. List tasks in column A, get the first task, and perform deep comparison with createdTask
		listResp := ac.Do(t, http.MethodGet, "/v1/boards/"+board.ID+"/columns/"+columnA.ID+"/tasks", nil)
		defer func() {
			_ = listResp.Body.Close()
		}()
		if listResp.StatusCode != http.StatusOK {
			t.Fatalf("got list status %d, want %d", listResp.StatusCode, http.StatusOK)
		}

		listedTasks := parseTasksList(t, listResp)
		if len(listedTasks) != 1 {
			t.Fatalf("got %d tasks in list, want 1", len(listedTasks))
		}
		if diff := cmp.Diff(createdTask, listedTasks[0]); diff != "" {
			t.Errorf("List item mismatch (-want +got):\n%s", diff)
		}

		// 5. Update by id with name and description, store response in updatedTask, validate fields, and ensure updatedAt advanced
		updatedName := "Updated Name"
		updatedDescription := "Updated description"
		WaitForTimestampTicker(t)
		updateResp := ac.Do(t, http.MethodPatch, "/v1/boards/"+board.ID+"/columns/"+columnA.ID+"/tasks/"+createdTask.ID, map[string]string{
			"name":        updatedName,
			"description": updatedDescription,
		})
		defer func() {
			_ = updateResp.Body.Close()
		}()
		if updateResp.StatusCode != http.StatusOK {
			t.Fatalf("got Update by id status %d, want %d", updateResp.StatusCode, http.StatusOK)
		}

		updatedTask := parseTask(t, updateResp)

		// Validation trick: revert changed fields in a clone and compare with createdTask
		checkTask := updatedTask
		checkTask.Name = createdTask.Name
		checkTask.Description = createdTask.Description
		checkTask.UpdatedAt = createdTask.UpdatedAt
		if diff := cmp.Diff(createdTask, checkTask); diff != "" {
			t.Errorf("UpdateByID() diff (-want +got):\n%s", diff)
		}

		// Verify specific changes
		if updatedTask.Name != updatedName {
			t.Errorf("got updated name %q, want %q", updatedTask.Name, updatedName)
		}
		if updatedTask.Description != updatedDescription {
			t.Errorf("got updated description %q, want %q", updatedTask.Description, updatedDescription)
		}

		updatedAtAfterUpdate, err := time.Parse(timeFormat, updatedTask.UpdatedAt)
		if err != nil {
			t.Fatalf("time.Parse(%q) error = %v", updatedTask.UpdatedAt, err)
		}
		updatedAtBeforeUpdate, err := time.Parse(timeFormat, createdTask.UpdatedAt)
		if err != nil {
			t.Fatalf("time.Parse(%q) error = %v", createdTask.UpdatedAt, err)
		}
		if !updatedAtAfterUpdate.After(updatedAtBeforeUpdate) {
			t.Errorf("updatedAt must advance after Update by id; got %v, previous %v", updatedAtAfterUpdate, updatedAtBeforeUpdate)
		}

		// 6. Create the second task in column A so we can later delete it from the middle and verify shifting
		createSecondResp := ac.Do(t, http.MethodPost, "/v1/boards/"+board.ID+"/columns/"+columnA.ID+"/tasks", map[string]string{
			"name":        "Second",
			"description": "second",
		})
		defer func() {
			_ = createSecondResp.Body.Close()
		}()
		if createSecondResp.StatusCode != http.StatusCreated {
			t.Fatalf("got second create status %d, want %d", createSecondResp.StatusCode, http.StatusCreated)
		}

		secondTask := parseTask(t, createSecondResp)
		if secondTask.Position != 2 {
			t.Errorf("got second task position %d, want %d", secondTask.Position, 2)
		}

		// 7. Create the third task in column A so delete can verify position compaction for trailing items
		createThirdResp := ac.Do(t, http.MethodPost, "/v1/boards/"+board.ID+"/columns/"+columnA.ID+"/tasks", map[string]string{
			"name":        "Third",
			"description": "third",
		})
		defer func() {
			_ = createThirdResp.Body.Close()
		}()
		if createThirdResp.StatusCode != http.StatusCreated {
			t.Fatalf("got third create status %d, want %d", createThirdResp.StatusCode, http.StatusCreated)
		}

		thirdTask := parseTask(t, createThirdResp)
		if thirdTask.Position != 3 {
			t.Errorf("got third task position %d, want %d", thirdTask.Position, 3)
		}

		// 8. Delete the middle task and expect the request to succeed
		deleteResp := ac.Do(t, http.MethodDelete, "/v1/boards/"+board.ID+"/columns/"+columnA.ID+"/tasks/"+secondTask.ID, nil)
		defer func() {
			_ = deleteResp.Body.Close()
		}()
		if deleteResp.StatusCode != http.StatusNoContent {
			t.Fatalf("got delete status %d, want %d", deleteResp.StatusCode, http.StatusNoContent)
		}

		// 9. List tasks again and verify delete preserved the first task and compacted positions
		listAfterDeleteResp := ac.Do(t, http.MethodGet, "/v1/boards/"+board.ID+"/columns/"+columnA.ID+"/tasks", nil)
		defer func() {
			_ = listAfterDeleteResp.Body.Close()
		}()
		if listAfterDeleteResp.StatusCode != http.StatusOK {
			t.Fatalf("got list status %d after delete, want %d", listAfterDeleteResp.StatusCode, http.StatusOK)
		}

		listedAfterDelete := parseTasksList(t, listAfterDeleteResp)
		if len(listedAfterDelete) != 2 {
			t.Fatalf("got %d tasks after delete, want 2", len(listedAfterDelete))
		}
		if listedAfterDelete[0].ID != updatedTask.ID || listedAfterDelete[0].Position != 1 {
			t.Errorf("got id=%s position=%d, want id=%s position=%d", listedAfterDelete[0].ID, listedAfterDelete[0].Position, updatedTask.ID, 1)
		}
		if listedAfterDelete[1].ID != thirdTask.ID || listedAfterDelete[1].Position != 2 {
			t.Errorf("got id=%s position=%d, want id=%s position=%d", listedAfterDelete[1].ID, listedAfterDelete[1].Position, thirdTask.ID, 2)
		}

		// 10. Move the third task up to first position within the same column and expect the new column and position in response
		moveResp := ac.Do(t, http.MethodPut, "/v1/boards/"+board.ID+"/columns/"+columnA.ID+"/tasks/"+thirdTask.ID+"/position", map[string]any{
			"targetColumnId": columnA.ID,
			"targetPosition": 1,
		})
		defer func() {
			_ = moveResp.Body.Close()
		}()
		if moveResp.StatusCode != http.StatusOK {
			t.Fatalf("got move status %d, want %d", moveResp.StatusCode, http.StatusOK)
		}

		movedPosition := parseTaskPosition(t, moveResp)
		if movedPosition.ColumnID != columnA.ID {
			t.Errorf("got move response column id %q, want %q", movedPosition.ColumnID, columnA.ID)
		}
		if movedPosition.Position != 1 {
			t.Errorf("got move response position %d, want %d", movedPosition.Position, 1)
		}

		// 11. List tasks again and verify move reordered the remaining tasks in column A
		listAfterMoveResp := ac.Do(t, http.MethodGet, "/v1/boards/"+board.ID+"/columns/"+columnA.ID+"/tasks", nil)
		defer func() {
			_ = listAfterMoveResp.Body.Close()
		}()
		if listAfterMoveResp.StatusCode != http.StatusOK {
			t.Fatalf("got list status %d after move, want %d", listAfterMoveResp.StatusCode, http.StatusOK)
		}

		listedAfterMove := parseTasksList(t, listAfterMoveResp)
		if len(listedAfterMove) != 2 {
			t.Fatalf("got %d tasks after move, want 2", len(listedAfterMove))
		}
		if listedAfterMove[0].ID != thirdTask.ID || listedAfterMove[0].Position != 1 {
			t.Errorf("got id=%s position=%d, want id=%s position=%d", listedAfterMove[0].ID, listedAfterMove[0].Position, thirdTask.ID, 1)
		}
		if listedAfterMove[1].ID != updatedTask.ID || listedAfterMove[1].Position != 2 {
			t.Errorf("got id=%s position=%d, want id=%s position=%d", listedAfterMove[1].ID, listedAfterMove[1].Position, updatedTask.ID, 2)
		}

		// 12. Move the third task across columns from column A to column B at position 1 and expect the new column and position in response
		crossMoveResp := ac.Do(t, http.MethodPut, "/v1/boards/"+board.ID+"/columns/"+columnA.ID+"/tasks/"+thirdTask.ID+"/position", map[string]any{
			"targetColumnId": columnB.ID,
			"targetPosition": 1,
		})
		defer func() {
			_ = crossMoveResp.Body.Close()
		}()
		if crossMoveResp.StatusCode != http.StatusOK {
			t.Fatalf("got cross-column move status %d, want %d", crossMoveResp.StatusCode, http.StatusOK)
		}

		crossMovedPosition := parseTaskPosition(t, crossMoveResp)
		if crossMovedPosition.ColumnID != columnB.ID {
			t.Errorf("got cross-column move response column id %q, want %q", crossMovedPosition.ColumnID, columnB.ID)
		}
		if crossMovedPosition.Position != 1 {
			t.Errorf("got cross-column move response position %d, want %d", crossMovedPosition.Position, 1)
		}

		// 13. List tasks in column A and verify the moved task is gone, and the remaining task kept position 1
		listAAfterCrossResp := ac.Do(t, http.MethodGet, "/v1/boards/"+board.ID+"/columns/"+columnA.ID+"/tasks", nil)
		defer func() {
			_ = listAAfterCrossResp.Body.Close()
		}()
		if listAAfterCrossResp.StatusCode != http.StatusOK {
			t.Fatalf("got list column A status %d after cross move, want %d", listAAfterCrossResp.StatusCode, http.StatusOK)
		}

		listedAAfterCross := parseTasksList(t, listAAfterCrossResp)
		if len(listedAAfterCross) != 1 {
			t.Fatalf("got %d tasks in column A after cross move, want 1", len(listedAAfterCross))
		}
		if listedAAfterCross[0].ID != updatedTask.ID || listedAAfterCross[0].Position != 1 {
			t.Errorf("got id=%s position=%d, want id=%s position=%d", listedAAfterCross[0].ID, listedAAfterCross[0].Position, updatedTask.ID, 1)
		}

		// 14. List tasks in column B and verify the moved task landed at position 1 with updated columnId
		listBAfterCrossResp := ac.Do(t, http.MethodGet, "/v1/boards/"+board.ID+"/columns/"+columnB.ID+"/tasks", nil)
		defer func() {
			_ = listBAfterCrossResp.Body.Close()
		}()
		if listBAfterCrossResp.StatusCode != http.StatusOK {
			t.Fatalf("got list column B status %d after cross move, want %d", listBAfterCrossResp.StatusCode, http.StatusOK)
		}

		listedBAfterCross := parseTasksList(t, listBAfterCrossResp)
		if len(listedBAfterCross) != 1 {
			t.Fatalf("got %d tasks in column B after cross move, want 1", len(listedBAfterCross))
		}
		if listedBAfterCross[0].ID != thirdTask.ID || listedBAfterCross[0].Position != 1 || listedBAfterCross[0].ColumnID != columnB.ID {
			t.Errorf("got id=%s position=%d columnId=%s, want id=%s position=%d columnId=%s", listedBAfterCross[0].ID, listedBAfterCross[0].Position, listedBAfterCross[0].ColumnID, thirdTask.ID, 1, columnB.ID)
		}
	})
}

func parseTask(t *testing.T, resp *http.Response) taskJSON {
	t.Helper()
	var tk taskJSON
	if err := json.NewDecoder(resp.Body).Decode(&tk); err != nil {
		t.Fatalf("Task Decode() error = %v", err)
	}
	return tk
}

func parseTasksList(t *testing.T, resp *http.Response) []taskJSON {
	t.Helper()
	var tasks []taskJSON
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		t.Fatalf("Tasks list Decode() error = %v", err)
	}
	return tasks
}

func parseTaskPosition(t *testing.T, resp *http.Response) taskPositionJSON {
	t.Helper()
	var p taskPositionJSON
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		t.Fatalf("Task position Decode() error = %v", err)
	}
	return p
}
