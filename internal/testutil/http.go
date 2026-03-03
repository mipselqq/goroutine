package testutil

import (
	"net/http/httptest"
	"testing"
)

func EnsureNoUnexpectedHeadersModified(t *testing.T, rr *httptest.ResponseRecorder, headersToModify []string) {
	for _, header := range headersToModify {
		if _, exists := rr.Header()[header]; exists {
			t.Errorf("Middleware should not modify the header %q, but it was found in response", header)
		}
	}
}
