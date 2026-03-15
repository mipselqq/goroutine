package testutil

import "testing"

// TODO: remove this code function or create local env/helpers_test.go
// and use it there to avoid uncontrolled global testutil growth
func UnsetEnv(t *testing.T, keys ...string) {
	for _, key := range keys {
		t.Setenv(key, "")
	}
}
