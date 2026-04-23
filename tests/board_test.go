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

type BoardJSON struct {
	ID          string `json:"id"`
	OwnerID     string `json:"ownerId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type AggregateColumnJSON struct {
	ColumnJSON
	Tasks []TaskJSON `json:"tasks"`
}

type BoardAggregateJSON struct {
	BoardJSON
	Columns []AggregateColumnJSON `json:"columns"`
}

func TestBoard_HappyPath(t *testing.T) {
	httpClient, ts, pool := Prelude(t)

	t.Run("Full board flow", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

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
			t.Fatalf("got status %d, want %d", createResp.StatusCode, http.StatusCreated)
		}

		createdBoard := parseBoard(t, createResp)

		if _, err := uuid.Parse(createdBoard.ID); err != nil {
			t.Errorf("uuid.Parse(%q) error = %v, want nil", createdBoard.ID, err)
		}
		if _, err := uuid.Parse(createdBoard.OwnerID); err != nil {
			t.Errorf("uuid.Parse(%q) error = %v, want nil", createdBoard.OwnerID, err)
		}
		if createdBoard.Name != name {
			t.Errorf("got name %q, want %q", createdBoard.Name, name)
		}
		if createdBoard.Description != description {
			t.Errorf("got description %q, want %q", createdBoard.Description, description)
		}
		if _, err := time.Parse(timeFormat, createdBoard.CreatedAt); err != nil {
			t.Errorf("time.Parse(%q) error = %v, want nil", createdBoard.CreatedAt, err)
		}
		if _, err := time.Parse(timeFormat, createdBoard.UpdatedAt); err != nil {
			t.Errorf("time.Parse(%q) error = %v, want nil", createdBoard.UpdatedAt, err)
		}

		// 3. Get by id, store response in getByIDBoard, and perform deep comparison with createdBoard
		oneResp := ac.Do(t, http.MethodGet, "/v1/boards/"+createdBoard.ID, nil)
		defer func() {
			_ = oneResp.Body.Close()
		}()
		if oneResp.StatusCode != http.StatusOK {
			t.Fatalf("got Get by id status %d, want %d", oneResp.StatusCode, http.StatusOK)
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
			t.Fatalf("got list status %d, want %d", listResp.StatusCode, http.StatusOK)
		}

		listedBoards := parseBoardsList(t, listResp)
		if len(listedBoards) != 1 {
			t.Fatalf("got %d boards in list, want 1", len(listedBoards))
		}
		if diff := cmp.Diff(createdBoard, listedBoards[0]); diff != "" {
			t.Errorf("List item mismatch (-want +got):\n%s", diff)
		}

		// 5. Update by id with name only, store response in updatedNameBoard, validate fields, and ensure updatedAt advanced
		updatedName := "Updated Name Only"
		WaitForTimestampTicker(t)
		updateNameResp := ac.Do(t, http.MethodPatch, "/v1/boards/"+createdBoard.ID, map[string]string{
			"name": updatedName,
		})
		defer func() {
			_ = updateNameResp.Body.Close()
		}()
		if updateNameResp.StatusCode != http.StatusOK {
			t.Fatalf("got Update by id status %d, want %d", updateNameResp.StatusCode, http.StatusOK)
		}

		updatedNameBoard := parseBoard(t, updateNameResp)

		// Validation trick: revert changed fields in a clone and compare with createdBoard
		checkBoard := updatedNameBoard
		checkBoard.Name = createdBoard.Name
		checkBoard.UpdatedAt = createdBoard.UpdatedAt
		if diff := cmp.Diff(createdBoard, checkBoard); diff != "" {
			t.Errorf("UpdateByID() diff (-want +got):\n%s", diff)
		}

		// Verify specific changes
		if updatedNameBoard.Name != updatedName {
			t.Errorf("got updated name %q, want %q", updatedNameBoard.Name, updatedName)
		}

		updatedAtAfterUpdate, _ := time.Parse(timeFormat, updatedNameBoard.UpdatedAt)
		updatedAtBeforeUpdate, _ := time.Parse(timeFormat, createdBoard.UpdatedAt)
		if !updatedAtAfterUpdate.After(updatedAtBeforeUpdate) {
			t.Errorf("updatedAt must advance after Update by id; got %v, previous %v", updatedAtAfterUpdate, updatedAtBeforeUpdate)
		}

		// 6. Get by id again, store response in getByIDBoardAfterUpdate, and perform deep comparison with updatedNameBoard
		afterUpdateResp := ac.Do(t, http.MethodGet, "/v1/boards/"+createdBoard.ID, nil)
		defer func() {
			_ = afterUpdateResp.Body.Close()
		}()

		if afterUpdateResp.StatusCode != http.StatusOK {
			t.Fatalf("got Get by id status %d after update, want %d", afterUpdateResp.StatusCode, http.StatusOK)
		}

		getByIDBoardAfterUpdate := parseBoard(t, afterUpdateResp)
		if diff := cmp.Diff(updatedNameBoard, getByIDBoardAfterUpdate); diff != "" {
			t.Errorf("Get by id after update mismatch (-want +got):\n%s", diff)
		}

		// 7. One column and one task: GET /aggregate must match the created board, column, and task from their POST responses.
		colResp := ac.Do(t, http.MethodPost, "/v1/boards/"+createdBoard.ID+"/columns", map[string]string{"name": "To Do"})
		defer func() { _ = colResp.Body.Close() }()
		if colResp.StatusCode != http.StatusCreated {
			t.Fatalf("got create column status %d, want %d", colResp.StatusCode, http.StatusCreated)
		}
		col := parseColumn(t, colResp)

		taskResp := ac.Do(t, http.MethodPost, "/v1/boards/"+createdBoard.ID+"/columns/"+col.ID+"/tasks", map[string]string{
			"name":        "One task",
			"description": "E2E aggregate",
		})
		defer func() { _ = taskResp.Body.Close() }()
		if taskResp.StatusCode != http.StatusCreated {
			t.Fatalf("got create task status %d, want %d", taskResp.StatusCode, http.StatusCreated)
		}
		task := parseTask(t, taskResp)

		aggResp := ac.Do(t, http.MethodGet, "/v1/boards/"+createdBoard.ID+"/aggregate", nil)
		defer func() { _ = aggResp.Body.Close() }()
		if aggResp.StatusCode != http.StatusOK {
			t.Fatalf("got aggregate status %d, want %d", aggResp.StatusCode, http.StatusOK)
		}
		gotAgg := parseBoardAggregate(t, aggResp)

		wantAgg := BoardAggregateJSON{
			BoardJSON: getByIDBoardAfterUpdate,
			Columns: []AggregateColumnJSON{
				{ColumnJSON: col, Tasks: []TaskJSON{task}},
			},
		}
		if diff := cmp.Diff(wantAgg, gotAgg); diff != "" {
			t.Errorf("aggregate vs POST responses (-want +got):\n%s", diff)
		}

		// 8. Delete by id and verify StatusNoContent
		delResp := ac.Do(t, http.MethodDelete, "/v1/boards/"+createdBoard.ID, nil)
		defer func() {
			_ = delResp.Body.Close()
		}()
		if delResp.StatusCode != http.StatusNoContent {
			t.Fatalf("got Delete by id status %d, want %d", delResp.StatusCode, http.StatusNoContent)
		}

		// 9. List boards and ensure an empty list is returned
		listAfterDelResp := ac.Do(t, http.MethodGet, "/v1/boards", nil)
		defer func() {
			_ = listAfterDelResp.Body.Close()
		}()
		if listAfterDelResp.StatusCode != http.StatusOK {
			t.Fatalf("got list status %d after delete, want %d", listAfterDelResp.StatusCode, http.StatusOK)
		}

		listedAfterDelete := parseBoardsList(t, listAfterDelResp)
		if len(listedAfterDelete) != 0 {
			t.Fatalf("got %d boards after delete, want 0", len(listedAfterDelete))
		}
	})
}

func parseBoard(t *testing.T, resp *http.Response) BoardJSON {
	t.Helper()
	var b BoardJSON
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		t.Fatalf("Board Decode() error = %v", err)
	}
	return b
}

func parseBoardsList(t *testing.T, resp *http.Response) []BoardJSON {
	t.Helper()
	var b []BoardJSON
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		t.Fatalf("Boards list Decode() error = %v", err)
	}
	return b
}

func parseBoardAggregate(t *testing.T, resp *http.Response) BoardAggregateJSON {
	t.Helper()
	var b BoardAggregateJSON
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		t.Fatalf("Board aggregate Decode() error = %v", err)
	}
	return b
}
