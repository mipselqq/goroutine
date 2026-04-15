//go:build e2e

package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/testutil"
)

type columnJSON struct {
	ID        string `json:"id"`
	BoardID   string `json:"boardId"`
	Name      string `json:"name"`
	Position  int64  `json:"position"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func TestColumn_Create(t *testing.T) {
	httpClient, ts, pool := Prelude(t)

	testutil.TruncateTable(t, pool, "columns")
	testutil.TruncateTable(t, pool, "boards")
	testutil.TruncateTable(t, pool, "users")

	ac := CreateUserAndAuthenticateClient(t, httpClient, ts.URL)

	boardName := testutil.ValidBoardName().String()
	boardDescription := testutil.ValidBoardDescription().String()

	createBoardResp := ac.Do(t, http.MethodPost, "/v1/boards", map[string]string{
		"name":        boardName,
		"description": boardDescription,
	})
	defer func() { _ = createBoardResp.Body.Close() }()
	if createBoardResp.StatusCode != http.StatusCreated {
		t.Fatalf("create board status = %d, want %d", createBoardResp.StatusCode, http.StatusCreated)
	}

	board := parseBoard(t, createBoardResp)

	createColumnResp := ac.Do(t, http.MethodPost, "/v1/boards/"+board.ID+"/columns", map[string]string{"name": "To Do"})
	defer func() { _ = createColumnResp.Body.Close() }()
	if createColumnResp.StatusCode != http.StatusCreated {
		t.Fatalf("create first column status = %d, want %d", createColumnResp.StatusCode, http.StatusCreated)
	}
	firstColumn := parseColumn(t, createColumnResp)
	if firstColumn.BoardID != board.ID {
		t.Errorf("first column boardId = %q, want %q", firstColumn.BoardID, board.ID)
	}
	if firstColumn.Position != 1 {
		t.Errorf("first column position = %d, want 1", firstColumn.Position)
	}
	if _, err := domain.ParseColumnID(firstColumn.ID); err != nil {
		t.Errorf("first column id invalid: %v", err)
	}
	if _, err := time.Parse(timeFormat, firstColumn.CreatedAt); err != nil {
		t.Errorf("first column createdAt invalid: %v", err)
	}
	if _, err := time.Parse(timeFormat, firstColumn.UpdatedAt); err != nil {
		t.Errorf("first column updatedAt invalid: %v", err)
	}

	createSecondColumnResp := ac.Do(t, http.MethodPost, "/v1/boards/"+board.ID+"/columns", map[string]string{"name": "In Progress"})
	defer func() { _ = createSecondColumnResp.Body.Close() }()
	if createSecondColumnResp.StatusCode != http.StatusCreated {
		t.Fatalf("create second column status = %d, want %d", createSecondColumnResp.StatusCode, http.StatusCreated)
	}
	secondColumn := parseColumn(t, createSecondColumnResp)
	if secondColumn.Position != 2 {
		t.Errorf("second column position = %d, want 2", secondColumn.Position)
	}

	invalidColumnResp := ac.Do(t, http.MethodPost, "/v1/boards/"+board.ID+"/columns", map[string]string{"name": "   "})
	defer func() { _ = invalidColumnResp.Body.Close() }()
	if invalidColumnResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid column status = %d, want %d", invalidColumnResp.StatusCode, http.StatusBadRequest)
	}
	var validationErr struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(invalidColumnResp.Body).Decode(&validationErr); err != nil {
		t.Fatalf("decode validation error: %v", err)
	}
	if validationErr.Code != "VALIDATION_ERROR" {
		t.Errorf("validation error code = %q, want %q", validationErr.Code, "VALIDATION_ERROR")
	}

	missingBoardID := domain.NewBoardID().String()
	missingBoardResp := ac.Do(t, http.MethodPost, "/v1/boards/"+missingBoardID+"/columns", map[string]string{"name": "Done"})
	defer func() { _ = missingBoardResp.Body.Close() }()
	if missingBoardResp.StatusCode != http.StatusNotFound {
		t.Fatalf("missing board status = %d, want %d", missingBoardResp.StatusCode, http.StatusNotFound)
	}
	var notFoundErr struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(missingBoardResp.Body).Decode(&notFoundErr); err != nil {
		t.Fatalf("decode missing board error: %v", err)
	}
	if notFoundErr.Code != "BOARD_NOT_FOUND" {
		t.Errorf("missing board code = %q, want %q", notFoundErr.Code, "BOARD_NOT_FOUND")
	}
}

func parseColumn(t *testing.T, resp *http.Response) columnJSON {
	t.Helper()
	var c columnJSON
	if err := json.NewDecoder(resp.Body).Decode(&c); err != nil {
		t.Fatalf("decode column: %v", err)
	}
	return c
}
