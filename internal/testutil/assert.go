package testutil

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

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
