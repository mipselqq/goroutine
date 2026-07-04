//go:build e2e

// Package tests contains end-to-end happy path tests.
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
	p := Prelude(t)

	t.Run("Full auth flow: register then login", func(t *testing.T) {
		testutil.TruncateAllTables(t, p.Pool)

		ac := CreateUserAndAuthenticateClient(t, p.HTTPClient, p.Server.URL)

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
		err := json.NewDecoder(whoamiResp.Body).Decode(&whoamiData)
		if err != nil {
			t.Fatalf("Whoami response Decode() error = %v", err)
		}

		_, err = uuid.Parse(whoamiData.UID)
		if err != nil {
			t.Errorf("uuid.Parse(%q) error = %v, want nil", whoamiData.UID, err)
		}
	})
}
