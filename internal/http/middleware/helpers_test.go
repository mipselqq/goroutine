package middleware_test

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"goroutine/internal/domain"
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

type MockAuth struct {
	VerifyTokenFunc func(ctx context.Context, token string) (domain.UserID, error)
}

func (m *MockAuth) VerifyToken(ctx context.Context, token string) (domain.UserID, error) {
	return m.VerifyTokenFunc(ctx, token)
}
