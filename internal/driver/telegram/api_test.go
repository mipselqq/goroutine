package telegram_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"

	"goroutine/internal/driver/telegram"
	"goroutine/internal/testutil"
)

type sendMessageQuery struct {
	ChatID int64
	Text   string
}

func sendMessagePath(token string) string {
	return fmt.Sprintf("POST /bot%s/sendMessage", token)
}

func TestAPIClient_SendMessage(t *testing.T) {
	token := testutil.ValidTelegramToken()
	message := testutil.ValidTelegramMessage()

	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
		wantQuery  *sendMessageQuery
	}{
		{
			name:       "Success",
			statusCode: http.StatusOK,
			wantErr:    false,
			wantQuery: &sendMessageQuery{
				ChatID: testutil.ValidTelegramChatID().Int64(),
				Text:   message.String(),
			},
		},
		{
			name:       "Non-OK status",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured sendMessageQuery
			mux := http.NewServeMux()
			mux.HandleFunc(sendMessagePath(token.RevealSecret()), func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				captured.ChatID, _ = strconv.ParseInt(q.Get("chat_id"), 10, 64)
				captured.Text = q.Get("text")
				w.WriteHeader(tt.statusCode)
			})

			ts := httptest.NewServer(mux)
			defer ts.Close()

			client := telegram.NewAPIClient(ts.URL, token)
			err := client.SendMessage(
				context.Background(),
				testutil.ValidTelegramChatID().Int64(),
				message,
			)

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("SendMessage() error = %v", err)
			}

			if tt.wantQuery != nil {
				if diff := cmp.Diff(*tt.wantQuery, captured); diff != "" {
					t.Errorf("query params mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
