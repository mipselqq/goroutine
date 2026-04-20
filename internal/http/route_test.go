package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	app "goroutine/internal/http"
	"goroutine/internal/http/handler"
	"goroutine/internal/http/httpschema"
	"goroutine/internal/http/middleware"
	"goroutine/internal/service"
	"goroutine/internal/testutil"
)

func TestNewRouter_Full(t *testing.T) {
	t.Parallel()

	logger := testutil.NewTestLogger(t)
	responder := httpschema.MustNewErrorResponder(logger, service.TimeNowRFC3339Millis)

	handlers := &handler.Handlers{
		Auth:    handler.NewAuth(logger, nil, responder),
		Health:  handler.NewHealth(logger),
		Boards:  handler.NewBoards(logger, nil, responder),
		Columns: handler.NewColumns(logger, nil, responder),
	}
	middlewares := &middleware.Middlewares{
		Metrics:   &spyMetricsMiddleware{},
		CORS:      &spyCorsMiddleware{},
		Auth:      &spyAuthMiddleware{},
		RequestID: &spyRequestIDMiddleware{},
	}

	router := app.NewRouter(handlers, middlewares)

	type entry struct {
		name   string
		method string
		path   string
	}

	UUIDv7 := func() string {
		return "018e1000-0000-7000-8000-000000000001"
	}

	tests := []struct {
		entry     entry
		metrics   bool
		cors      bool
		requestID bool
	}{
		{
			entry:   entry{"Register endpoint", http.MethodPost, "/v1/register"},
			metrics: true, cors: true, requestID: true,
		},
		{
			entry:   entry{"Login endpoint", http.MethodPost, "/v1/login"},
			metrics: true, cors: true, requestID: true,
		},
		{
			entry:   entry{"Health endpoint", http.MethodGet, "/v1/health"},
			metrics: true, cors: true, requestID: true,
		},
		{
			entry:   entry{"Boards list endpoint", http.MethodGet, "/v1/boards"},
			metrics: true, cors: true, requestID: true,
		},
		{
			entry:   entry{"Board by id endpoint", http.MethodGet, "/v1/boards/" + UUIDv7()},
			metrics: true, cors: true, requestID: true,
		},
		{
			entry:   entry{"UpdateByID board endpoint", http.MethodPatch, "/v1/boards/" + UUIDv7()},
			metrics: true, cors: true, requestID: true,
		},
		{
			entry:   entry{"Delete board endpoint", http.MethodDelete, "/v1/boards/" + UUIDv7()},
			metrics: true, cors: true, requestID: true,
		},
		{
			entry:   entry{"Create column endpoint", http.MethodPost, "/v1/boards/" + UUIDv7() + "/columns"},
			metrics: true, cors: true, requestID: true,
		},
		{
			entry:   entry{"List columns endpoint", http.MethodGet, "/v1/boards/" + UUIDv7() + "/columns"},
			metrics: true, cors: true, requestID: true,
		},
		{
			entry:   entry{"UpdateByID column endpoint", http.MethodPatch, "/v1/boards/" + UUIDv7() + "/columns/" + UUIDv7()},
			metrics: true, cors: true, requestID: true,
		},
		{
			entry:   entry{"Delete column endpoint", http.MethodDelete, "/v1/boards/" + UUIDv7() + "/columns/" + UUIDv7()},
			metrics: true, cors: true, requestID: true,
		},
		{
			entry:   entry{"Swagger endpoint", http.MethodGet, "/v1/swagger/index.html"},
			metrics: false, cors: true, requestID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.entry.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.entry.method, tt.entry.path, http.NoBody)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			if rr.Code == http.StatusNotFound {
				t.Errorf("Path %q %q not registered in router (got 404)", tt.entry.method, tt.entry.path)
			}

			hasMetrics := rr.Header().Get("X-Metrics-Tracked") == "true"
			if hasMetrics != tt.metrics {
				t.Errorf("Metrics middleware application mismatch for %q: got %v, want %v", tt.entry.path, hasMetrics, tt.metrics)
			}

			hasCors := rr.Header().Get("X-Cors-Tracked") == "true"
			if hasCors != tt.cors {
				t.Errorf("CORS middleware application mismatch for %q: got %v, want %v", tt.entry.path, hasCors, tt.cors)
			}

			hasReqID := rr.Header().Get("X-RequestId-Tracked") == "true"
			if hasReqID != tt.requestID {
				t.Errorf("RequestID middleware application mismatch for %q: got %v, want %v", tt.entry.path, hasReqID, tt.requestID)
			}
		})
	}

	t.Run("Non-existing endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/non-existing", http.NoBody)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		testutil.AssertStatusCode(t, rr, http.StatusNotFound)
	})
}

func TestNewAdminRouter(t *testing.T) {
	t.Parallel()

	router := app.NewAdminRouter()

	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	_, pattern := router.Handler(req)

	if pattern == "" {
		t.Errorf("Path GET /metrics not registered in admin router")
	}
}
