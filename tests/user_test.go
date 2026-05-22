//go:build e2e

package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"goroutine/internal/testutil"
)

func TestUser_TelegramLinkSuccess(t *testing.T) {
	httpClient, ts, pool := Prelude(t)

	t.Run("Telegram link full flow", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		ac := CreateUserAndAuthenticateClient(t, httpClient, ts.URL)

		resp := ac.Do(t, http.MethodPost, "/v1/users/me/telegram/link", nil)
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("got status %d, want %d", resp.StatusCode, http.StatusOK)
		}

		var linkBody struct {
			Token string `json:"token"`
		}
		err := json.NewDecoder(resp.Body).Decode(&linkBody)
		if err != nil {
			t.Fatalf("Decode() error = %v", err)
		}

		if linkBody.Token == "" {
			t.Error("got empty token, want non-empty token")
		}
	})
}
