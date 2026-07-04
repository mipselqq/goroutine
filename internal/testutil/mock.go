package testutil

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

type MockTelegramAPI struct {
	Server *httptest.Server

	LastChatID int64
	LastText   string
	Called     bool
}

func NewMockTelegramAPI(t *testing.T, statusCode int) *MockTelegramAPI {
	t.Helper()

	m := &MockTelegramAPI{}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/sendMessage") {
			t.Errorf("mock Telegram API: unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		q := r.URL.Query()
		m.LastChatID, _ = strconv.ParseInt(q.Get("chat_id"), 10, 64)
		m.LastText = q.Get("text")
		m.Called = true

		w.WriteHeader(statusCode)
	}))

	return m
}

func (m *MockTelegramAPI) URL() string {
	return m.Server.URL
}

func (m *MockTelegramAPI) Close() {
	m.Server.Close()
}
