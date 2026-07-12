package middleware_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"goroutine/internal/http/middleware"
	"goroutine/internal/testutil"
)

func TestTimeout(t *testing.T) {
	var gotCtx context.Context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { gotCtx = r.Context() })

	wantTimeout := 500 * time.Millisecond
	mw := middleware.NewTimeout(wantTimeout)

	req, rr := testutil.NewJSONRequestAndRecorder(t, http.MethodGet, "/", http.NoBody)
	mw.Wrap(handler).ServeHTTP(rr, req)

	deadline, ok := gotCtx.Deadline()
	if !ok {
		t.Fatal("deadline not set")
	}

	timeout := time.Until(deadline)
	threshold := 30 * time.Millisecond

	if timeout > wantTimeout || timeout < wantTimeout-threshold {
		t.Errorf("want deadline %vms, got %vms", wantTimeout, timeout)
	}
}
