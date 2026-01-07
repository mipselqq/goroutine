package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-todo/internal/handler"
	"go-todo/internal/testutil"
)

func TestHealthHandler(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	rr := httptest.NewRecorder()
	h := handler.NewHealth(testutil.CreateTestLogger(t))

	h.Health(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned %v, expected %v", rr.Code, http.StatusOK)
	}
}
