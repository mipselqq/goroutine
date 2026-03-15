package http_test

import (
	"bytes"
	"net/http"
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

type spyMetricsMiddleware struct{}

func (s *spyMetricsMiddleware) Wrap(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Metrics-Tracked", "true")
		next.ServeHTTP(w, r)
	}
}

type spyCorsMiddleware struct{}

func (s *spyCorsMiddleware) Wrap(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Cors-Tracked", "true")
		next.ServeHTTP(w, r)
	}
}

type spyAuthMiddleware struct{}

func (s *spyAuthMiddleware) Wrap(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Auth-Tracked", "true")
		next.ServeHTTP(w, r)
	}
}

type spyRequestIDMiddleware struct{}

func (s *spyRequestIDMiddleware) Wrap(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RequestId-Tracked", "true")
		next.ServeHTTP(w, r)
	}
}
