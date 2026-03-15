package testutil

import (
	"bytes"
	"mime"
	"net/http"
	"net/http/httptest"
	"testing"
)

func AssertContentType(t *testing.T, rr *httptest.ResponseRecorder, expectedMediaType string) {
	t.Helper()

	contentType := rr.Header().Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatalf("Failed to parse MIME %q", contentType)
	}
	if mediaType != expectedMediaType {
		t.Errorf("Expected %q, got %q", expectedMediaType, mediaType)
	}
}

func AssertResponseBody(t *testing.T, rr *httptest.ResponseRecorder, expectedBody string) {
	t.Helper()

	if expectedBody != "" {
		actualBody := bytes.TrimSpace(rr.Body.Bytes())
		if string(actualBody) != expectedBody {
			t.Logf("Expected body:")
			t.Logf("%q", expectedBody)
			t.Logf("Got:")
			t.Logf("%q", string(actualBody))
			t.Fail()
		}
	}
}

func NewJSONRequestAndRecorder(t *testing.T, method, url, body string) (*http.Request, *httptest.ResponseRecorder) {
	t.Helper()

	req := httptest.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	return req, rr
}
