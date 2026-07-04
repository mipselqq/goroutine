package telegram_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"goroutine/internal/driver/telegram"
	"goroutine/internal/testutil"
)

type sendMessageQuery struct {
	ChatID int64
	Text   string
}

func TestClient_SendMessage(t *testing.T) {
	token := testutil.ValidTelegramToken()
	chatID := testutil.ValidTelegramChatID()
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
				ChatID: chatID.Int64(),
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
			mock := testutil.NewMockTelegramAPI(t, tt.statusCode)
			defer mock.Close()

			client := telegram.NewClient(mock.URL(), token)
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
				got := sendMessageQuery{
					ChatID: mock.LastChatID,
					Text:   mock.LastText,
				}
				diff := cmp.Diff(*tt.wantQuery, got)
				if diff != "" {
					t.Errorf("query params mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
