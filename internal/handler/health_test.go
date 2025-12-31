package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-todo/internal/handler"
)

func TestHealthHandler(t *testing.T) {
	t.Parallel()

	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	rr := httptest.NewRecorder()
	h := handler.NewHealth()

	// Act
	h.Health(rr, req)

	// Assert
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned %v, expected %v", rr.Code, http.StatusOK)
	}
}
