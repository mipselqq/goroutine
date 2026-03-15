package handler_test

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
	RegisterFunc func(ctx context.Context, email domain.Email, password domain.Password) error
	LoginFunc    func(ctx context.Context, email domain.Email, password domain.Password) (string, error)
}

func (m *MockAuth) Register(ctx context.Context, email domain.Email, password domain.Password) error {
	return m.RegisterFunc(ctx, email, password)
}

func (m *MockAuth) Login(ctx context.Context, email domain.Email, password domain.Password) (string, error) {
	return m.LoginFunc(ctx, email, password)
}

type MockBoards struct {
	CreateFunc func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error)
}

func (m *MockBoards) Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
	return m.CreateFunc(ctx, ownerID, name, description)
}
