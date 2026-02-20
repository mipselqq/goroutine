package testutil

import "testing"

func UnsetEnv(t *testing.T, keys ...string) {
	for _, key := range keys {
		t.Setenv(key, "")
	}
}
