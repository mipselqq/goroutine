package handlers_test

import (
	"go-todo/internal/handlers"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	// Arrange
	r := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	handler := handlers.NewHealthHandler()
	handler.Health(rr, r)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned %v, expected %v", rr.Code, http.StatusOK)
	}
}
