package testutil

import (
	"bytes"
	"encoding/json"
	"mime"
	"net/http"
	"net/http/httptest"
	"reflect"
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

func MarshalJSONBody(t *testing.T, body any) string {
	t.Helper()

	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	return string(b)
}

func NewJSONRequestAndRecorder(t *testing.T, method, url string, body any) (*http.Request, *httptest.ResponseRecorder) {
	t.Helper()

	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(method, url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	return req, rr
}

func AssertStatusCode(t *testing.T, rr *httptest.ResponseRecorder, expectedStatusCode int) {
	t.Helper()

	if rr.Code != expectedStatusCode {
		t.Errorf("Expected status code %d, got %d", expectedStatusCode, rr.Code)
	}
}

func AssertResponseBody(t *testing.T, rr *httptest.ResponseRecorder, expected any) {
	t.Helper()

	if expected == nil {
		// Middlewares may not return
		return
	}

	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("Failed to marshal expected body: %v", err)
	}

	actualBody := rr.Body.Bytes()

	var actualDecoded, expectedDecoded any
	if err := json.Unmarshal(actualBody, &actualDecoded); err != nil {
		t.Fatalf("Failed to unmarshal actual body: %v\nBody: %q", err, string(actualBody))
	}
	if err := json.Unmarshal(expectedJSON, &expectedDecoded); err != nil {
		t.Fatalf("Failed to unmarshal expected body: %v\nBody: %q", err, string(expectedJSON))
	}

	if !reflect.DeepEqual(actualDecoded, expectedDecoded) {
		t.Errorf("Response body mismatch\nExpected: %s\nGot: %s", string(expectedJSON), string(actualBody))
	}
}
