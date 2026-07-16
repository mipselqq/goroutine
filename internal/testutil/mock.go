package testutil

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

type MockTelegramAPI struct {
	t *testing.T
	*httptest.Server

	LastChatID    int64
	LastMessage   string
	messagesCount int
}

func NewMockTelegramAPI(t *testing.T, statusCode int) *MockTelegramAPI {
	t.Helper()

	m := &MockTelegramAPI{t: t}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/sendMessage") {
			t.Errorf("mock Telegram API: unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		chatID, err := strconv.ParseInt(r.FormValue("chat_id"), 10, 64)
		if err != nil {
			t.Errorf("mock Telegram API: failed to parse chat_id: %v", err)
		}
		m.LastChatID = chatID
		m.LastMessage = r.FormValue("text")
		m.messagesCount++

		w.WriteHeader(statusCode)
	}))

	return m
}

func (m *MockTelegramAPI) AssertMessagesCount(count int) {
	m.t.Helper()

	if m.messagesCount != count {
		m.t.Errorf("mock Telegram API: expected %d messages, got %d", count, m.messagesCount)
	}
}
