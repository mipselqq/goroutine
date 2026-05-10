package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goroutine/internal/http/handler"
)

func TestDecodeJSONLimited(t *testing.T) {
	t.Parallel()

	t.Run("body under 20KB decodes successfully", func(t *testing.T) {
		t.Parallel()

		body := `{"name": "Test Board"}`
		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		var v struct {
			Name string `json:"name"`
		}
		err := handler.DecodeJSONLimited(req, &v)
		if err != nil {
			t.Fatalf("got error %v, want nil", err)
		}
		if v.Name != "Test Board" {
			t.Errorf("got name %q, want %q", v.Name, "Test Board")
		}
	})

	t.Run("body over 20KB returns ErrBodyTooLarge", func(t *testing.T) {
		t.Parallel()

		body := `{"name": "` + strings.Repeat("x", 20*1024) + `"}`
		req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		var v struct {
			Name string `json:"name"`
		}
		err := handler.DecodeJSONLimited(req, &v)
		if !errors.Is(err, handler.ErrBodyTooLarge) {
			t.Fatalf("got error %v, want ErrBodyTooLarge", err)
		}
	})
}
