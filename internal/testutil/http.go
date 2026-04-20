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

func AssertContentType(t *testing.T, rr *httptest.ResponseRecorder, wantMediaType string) {
	t.Helper()

	contentType := rr.Header().Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatalf("mime.ParseMediaType(%q) error = %v", contentType, err)
	}
	if mediaType != wantMediaType {
		t.Errorf("got content type %q, want %q", mediaType, wantMediaType)
	}
}

func MarshalJSONBody(t *testing.T, body any) string {
	t.Helper()

	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	return string(b)
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

func AssertStatusCode(t *testing.T, rr *httptest.ResponseRecorder, wantStatusCode int) {
	t.Helper()

	if rr.Code != wantStatusCode {
		t.Errorf("got status %d, want %d", rr.Code, wantStatusCode)
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
	if err := json.Unmarshal(gotBody, &gotDecoded); err != nil {
		t.Fatalf("json.Unmarshal(response body) error = %v\nBody: %q", err, string(gotBody))
	}
	if err := json.Unmarshal(wantJSON, &wantDecoded); err != nil {
		t.Fatalf("json.Unmarshal(want JSON) error = %v\nWant JSON: %q", err, string(wantJSON))
	}

	if !reflect.DeepEqual(gotDecoded, wantDecoded) {
		t.Errorf("got response body %s, want %s", string(gotBody), string(wantJSON))
	}
}
