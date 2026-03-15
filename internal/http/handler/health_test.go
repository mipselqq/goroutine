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

	testutil.AssertStatusCode(t, rr, http.StatusOK)
	testutil.AssertContentType(t, rr, "application/json")
	testutil.AssertResponseBody(t, rr, map[string]string{"status": "ok"})
}
