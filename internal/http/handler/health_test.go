package handler_test

import (
	"net/http"
	"testing"

	"goroutine/internal/http/handler"
	"goroutine/internal/testutil"
)

func TestHealthHandler(t *testing.T) {
	t.Parallel()

	req, rr := testutil.NewJSONRequestAndRecorder(t, http.MethodGet, "/health", "")
	logger := testutil.NewTestLogger(t)
	h := handler.NewHealth(logger)

	h.Health(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned %v, expected %v", rr.Code, http.StatusOK)
	}

	testutil.AssertContentType(t, rr, "application/json")
}
