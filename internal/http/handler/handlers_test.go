package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goroutine/internal/http/handler"
)

func TestDecodeJSONLimited(t *testing.T) {
	t.Parallel()

	t.Run("body over 20KB returns an error", func(t *testing.T) {
		t.Parallel()

		body := `{"name": "` + strings.Repeat("x", 20*1024) + `"}`
		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		var v struct {
			Name string `json:"name"`
		}
		err := handler.DecodeJSONLimited(req, &v)
		if err == nil {
			t.Fatal("got nil error, want body too large error")
		}
	})
}
