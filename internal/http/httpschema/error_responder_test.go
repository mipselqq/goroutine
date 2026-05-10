package httpschema_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"goroutine/internal/http/httpschema"
	"goroutine/internal/testutil"
)

// Other methods are tested in handler tests indirectly

func TestErrorResponder_InternalError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		err          error
		wantHTTPCode int
		wantBody     map[string]any
		wantLevel    string
	}{
		{
			name:         "wrapped context.Canceled returns 499 and logs DEBUG",
			err:          fmt.Errorf("task service: create get board: %w", context.Canceled),
			wantHTTPCode: 499, // ClientClosedRequest, non-standard, used by nginx
			wantBody:     nil,
			wantLevel:    "DEBUG",
		},
		{
			name:         "wrapped context.DeadlineExceeded returns 408 and logs DEBUG",
			err:          fmt.Errorf("board repo: get by id: %w", context.DeadlineExceeded),
			wantHTTPCode: http.StatusRequestTimeout,
			wantBody:     nil,
			wantLevel:    "DEBUG",
		},
		{
			name:         "ordinary error returns 500 and logs ERROR",
			err:          fmt.Errorf("db connection refused"),
			wantHTTPCode: http.StatusInternalServerError,
			wantBody: map[string]any{
				"code":      "INTERNAL_SERVER_ERROR",
				"message":   "Internal server error",
				"timestamp": testutil.FixedTimeNowStr(),
			},
			wantLevel: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger, buf := testutil.NewBufJsonLogger(t, slog.LevelDebug)
			er := httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr)

			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			rr := httptest.NewRecorder()

			er.InternalError(rr, req, tt.err)

			testutil.AssertStatusCode(t, rr, tt.wantHTTPCode)

			if tt.wantBody == nil {
				if rr.Body.Len() != 0 {
					t.Fatalf("got body %q, want empty", rr.Body.String())
				}
			} else {
				testutil.AssertResponseBody(t, rr, tt.wantBody)
			}

			assertLogLevel(t, buf, tt.wantLevel)
		})
	}
}

func assertLogLevel(t *testing.T, buf *bytes.Buffer, wantLevel string) {
	t.Helper()

	var entry struct {
		Level string `json:"level"`
	}
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log entry: %v", err)
	}
	if entry.Level != wantLevel {
		t.Errorf("got log level %q, want %q", entry.Level, wantLevel)
	}
}
