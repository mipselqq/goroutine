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

const UUIDv7 = "018e1000-0000-7000-8000-000000000001"

func TestNewRouter_Full(t *testing.T) {
	t.Parallel()

	logger := testutil.NewTestLogger(t)
	responder := httpschema.MustNewErrorResponder(logger, service.TimeNowRFC3339Millis)

	handlers := &handler.Handlers{
		Auth:    handler.NewAuth(logger, nil, responder),
		Health:  handler.NewHealth(logger),
		Boards:  handler.NewBoards(logger, nil, responder),
		Columns: handler.NewColumns(logger, nil, responder),
		Tasks:   handler.NewTasks(logger, nil, responder),
	}
	middlewares := &middleware.Middlewares{
		Metrics:   &spyMetricsMiddleware{},
		CORS:      &spyCorsMiddleware{},
		Auth:      &spyAuthMiddleware{},
		RequestID: &spyRequestIDMiddleware{},
	}

	router := app.NewRouter(handlers, middlewares, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	type entry struct {
		name   string
		method string
		path   string
	}

	tests := []struct {
		entry     entry
		auth      bool
		metrics   bool
		cors      bool
		requestID bool
	}{
		{
			entry: entry{"Register", http.MethodPost, "/v1/register"},
			auth:  false, metrics: true, cors: true, requestID: true,
		},
		{
			entry: entry{"Login", http.MethodPost, "/v1/login"},
			auth:  false, metrics: true, cors: true, requestID: true,
		},
		{
			entry: entry{"Health", http.MethodGet, "/v1/health"},
			auth:  false, metrics: true, cors: true, requestID: true,
		},
		{
			entry: entry{"Boards list", http.MethodGet, "/v1/boards"},
			auth:  true, metrics: true, cors: true, requestID: true,
		},
		{
			entry: entry{"Board by id", http.MethodGet, "/v1/boards/" + UUIDv7},
			auth:  true, metrics: true, cors: true, requestID: true,
		},
		{
			entry: entry{"Board aggregate by id", http.MethodGet, "/v1/boards/" + UUIDv7 + "/aggregate"},
			auth:  true, metrics: true, cors: true, requestID: true,
		},
		{
			entry: entry{"UpdateByID board", http.MethodPatch, "/v1/boards/" + UUIDv7},
			auth:  true, metrics: true, cors: true, requestID: true,
		},
		{
			entry: entry{"Delete board", http.MethodDelete, "/v1/boards/" + UUIDv7},
			auth:  true, metrics: true, cors: true, requestID: true,
		},
		{
			entry: entry{"Create column", http.MethodPost, "/v1/boards/" + UUIDv7 + "/columns"},
			auth:  true, metrics: true, cors: true, requestID: true,
		},
		{
			entry: entry{"List columns", http.MethodGet, "/v1/boards/" + UUIDv7 + "/columns"},
			auth:  true, metrics: true, cors: true, requestID: true,
		},
		{
			entry: entry{"UpdateByID column", http.MethodPatch, "/v1/boards/" + UUIDv7 + "/columns/" + UUIDv7},
			auth:  true, metrics: true, cors: true, requestID: true,
		},
		{
			entry: entry{"Move column", http.MethodPut, "/v1/boards/" + UUIDv7 + "/columns/" + UUIDv7 + "/position"},
			auth:  true, metrics: true, cors: true, requestID: true,
		},
		{
			entry: entry{"Delete column", http.MethodDelete, "/v1/boards/" + UUIDv7 + "/columns/" + UUIDv7},
			auth:  true, metrics: true, cors: true, requestID: true,
		},
		{
			entry: entry{"Create task", http.MethodPost, "/v1/boards/" + UUIDv7 + "/columns/" + UUIDv7 + "/tasks"},
			auth:  true, metrics: true, cors: true, requestID: true,
		},
		{
			entry: entry{"List tasks", http.MethodGet, "/v1/boards/" + UUIDv7 + "/columns/" + UUIDv7 + "/tasks"},
			auth:  true, metrics: true, cors: true, requestID: true,
		},
		{
			entry: entry{"UpdateByID task", http.MethodPatch, "/v1/boards/" + UUIDv7 + "/columns/" + UUIDv7 + "/tasks/" + UUIDv7},
			auth:  true, metrics: true, cors: true, requestID: true,
		},
		{
			entry: entry{"Move task", http.MethodPut, "/v1/boards/" + UUIDv7 + "/columns/" + UUIDv7 + "/tasks/" + UUIDv7 + "/position"},
			auth:  true, metrics: true, cors: true, requestID: true,
		},
		{
			entry: entry{"Delete task", http.MethodDelete, "/v1/boards/" + UUIDv7 + "/columns/" + UUIDv7 + "/tasks/" + UUIDv7},
			auth:  true, metrics: true, cors: true, requestID: true,
		},
		{
			entry: entry{"Swagger", http.MethodGet, "/v1/swagger/index.html"},
			auth:  false, metrics: false, cors: true, requestID: true,
		},
		{
			entry: entry{"Telegram webhook", http.MethodPost, "/webhook/telegram"},
			auth:  false, metrics: true, cors: true, requestID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.entry.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.entry.method, tt.entry.path, http.NoBody)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			if rr.Code == http.StatusNotFound {
				t.Errorf("got 404 for %s %q, want registered route", tt.entry.method, tt.entry.path)
			}

			hasAuth := rr.Header().Get("X-Auth-Tracked") == "true"
			if hasAuth != tt.auth {
				t.Errorf("got auth middleware=%v for %q, want %v", hasAuth, tt.entry.path, tt.auth)
			}

			hasMetrics := rr.Header().Get("X-Metrics-Tracked") == "true"
			if hasMetrics != tt.metrics {
				t.Errorf("got metrics middleware=%v for %q, want %v", hasMetrics, tt.entry.path, tt.metrics)
			}

			hasCors := rr.Header().Get("X-Cors-Tracked") == "true"
			if hasCors != tt.cors {
				t.Errorf("got CORS middleware=%v for %q, want %v", hasCors, tt.entry.path, tt.cors)
			}

			hasReqID := rr.Header().Get("X-RequestId-Tracked") == "true"
			if hasReqID != tt.requestID {
				t.Errorf("got request ID middleware=%v for %q, want %v", hasReqID, tt.entry.path, tt.requestID)
			}
		})
	}

	t.Run("Non-existing endpoint", func(t *testing.T) {
		t.Parallel()

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
		t.Errorf("got empty pattern for GET /metrics, want registered route")
	}
}
