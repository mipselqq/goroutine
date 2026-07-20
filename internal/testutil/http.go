package testutil

import (
	"bytes"
	"encoding/json"
	"mime"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func AssertContentType(t *testing.T, rr *httptest.ResponseRecorder, want string) {
	t.Helper()

	contentType := rr.Header().Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatalf("mime.ParseMediaType(%q) error = %v", contentType, err)
	}
	if mediaType != want {
		t.Errorf("got content type %q, want %q", mediaType, want)
	}
}

func NewJSONRequestAndRecorder(t *testing.T, method, url string, body any) (*http.Request, *httptest.ResponseRecorder) {
	t.Helper()

	var resultingBytes []byte
	switch value := body.(type) {
	case json.RawMessage:
		resultingBytes = value
	default:
		var err error
		resultingBytes, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal(request body) error = %v", err)
		}
	}

	req := httptest.NewRequest(method, url, bytes.NewBuffer(resultingBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	return req, rr
}

func AssertStatusCode(t *testing.T, rr *httptest.ResponseRecorder, want int) {
	t.Helper()

	if rr.Code != want {
		t.Errorf("got status %d, want %d", rr.Code, want)
	}
}

func AssertResponseBody(t *testing.T, rr *httptest.ResponseRecorder, want any) {
	t.Helper()

	if want == nil {
		if rr.Body.Len() != 0 {
			t.Fatalf("got body %q, want empty", rr.Body.String())
		}
		return
	}

	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("json.Marshal(want) error = %v", err)
	}

	gotBody := rr.Body.Bytes()

	var gotDecoded, wantDecoded any
	err = json.Unmarshal(gotBody, &gotDecoded)
	if err != nil {
		t.Fatalf("json.Unmarshal(response body) error = %v\nBody: %q", err, string(gotBody))
	}
	err = json.Unmarshal(wantJSON, &wantDecoded)
	if err != nil {
		t.Fatalf("json.Unmarshal(want JSON) error = %v\nWant JSON: %q", err, string(wantJSON))
	}

	if diff := cmp.Diff(wantDecoded, gotDecoded); diff != "" {
		t.Errorf("got response body mismatch (-want +got):\n%s", diff)
	}
}
