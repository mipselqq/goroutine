//go:build e2e

package tests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"goroutine/internal/testutil"

	"github.com/google/uuid"
)

func TestAuth_HappyPath(t *testing.T) {
	httpClient, ts, pool := Prelude(t)

	t.Run("Full auth flow: register then login", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "users")

		ac := CreateUserAndAuthenticateClient(t, httpClient, ts.URL)

		parts := strings.Split(ac.Token, ".")
		if len(parts) != 3 {
			t.Fatalf("got %d JWT segments, want 3", len(parts))
		}

		whoamiResp := ac.Do(t, http.MethodGet, "/v1/whoami", nil)
		defer func() {
			_ = whoamiResp.Body.Close()
		}()

		if whoamiResp.StatusCode != http.StatusOK {
			t.Fatalf("got status %d, want %d", whoamiResp.StatusCode, http.StatusOK)
		}

		var whoamiData struct {
			UID string `json:"uid"`
		}
		if err := json.NewDecoder(whoamiResp.Body).Decode(&whoamiData); err != nil {
			t.Fatalf("Whoami response Decode() error = %v", err)
		}

		if _, err := uuid.Parse(whoamiData.UID); err != nil {
			t.Errorf("uuid.Parse(%q) error = %v, want nil", whoamiData.UID, err)
		}
	})
}
