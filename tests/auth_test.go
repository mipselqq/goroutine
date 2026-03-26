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

		// Register, login, append token to each request
		ac := CrateUserAndAuthenticateClient(t, httpClient, ts.URL)

		parts := strings.Split(ac.Token, ".")
		if len(parts) != 3 {
			t.Fatal("Got invalid JWT token")
		}

		whoamiResp := ac.Do(t, http.MethodGet, "/v1/whoami", nil)
		defer func() {
			_ = whoamiResp.Body.Close()
		}()

		if whoamiResp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", whoamiResp.StatusCode)
		}

		var whoamiData struct {
			UID string `json:"uid"`
		}
		if err := json.NewDecoder(whoamiResp.Body).Decode(&whoamiData); err != nil {
			t.Fatalf("Failed to decode whoami response: %v", err)
		}

		if _, err := uuid.Parse(whoamiData.UID); err != nil {
			t.Errorf("Expected valid UUID user ID, got %q: %v", whoamiData.UID, err)
		}
	})
}
