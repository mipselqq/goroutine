package middleware_test

import (
	"net/http"
	"strconv"
	"testing"

	"goroutine/internal/http/middleware"
	"goroutine/internal/testutil"

	"github.com/prometheus/client_golang/prometheus"
	promTestutil "github.com/prometheus/client_golang/prometheus/testutil"
)

func TestAbstractMetricPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		rawPath      string
		abstractPath string
	}{
		{name: "Empty", rawPath: "", abstractPath: ""},
		{name: "Root", rawPath: "/", abstractPath: "/"},
		{name: "Static segments preserved", rawPath: "/v1/register", abstractPath: "/v1/register"},
		{name: "UUID v7 segment abstracted", rawPath: "/v1/boards/018e1000-0000-7000-8000-000000000001", abstractPath: "/v1/boards/{id}"},
		{name: "Several UUID segments", rawPath: "/v1/a/550e8400-e29b-41d4-a716-446655440000/b/018e1000-0000-7000-8000-000000000001", abstractPath: "/v1/a/{id}/b/{id}"},
		{name: "Opaque slug not UUID", rawPath: "/v1/boards/not-a-uuid", abstractPath: "/v1/boards/not-a-uuid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			have := middleware.AbstractMetricPath(tt.rawPath)
			if have != tt.abstractPath {
				t.Errorf("AbstractMetricPath(%q) = %q, want %q", tt.rawPath, have, tt.abstractPath)
			}
		})
	}
}

func TestMetrics_Collection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		requestPath   string
		abstractPath  string
		method        string
		handlerStatus int
	}{
		{
			name:          "register 200",
			requestPath:   "/v1/register",
			abstractPath:  "/v1/register",
			method:        http.MethodPost,
			handlerStatus: http.StatusOK,
		},
		{
			name:          "health 403",
			requestPath:   "/v1/health",
			abstractPath:  "/v1/health",
			method:        http.MethodGet,
			handlerStatus: http.StatusForbidden,
		},
		{
			name:          "board by id path label uses placeholder",
			requestPath:   "/v1/boards/550e8400-e29b-41d4-a716-446655440000",
			abstractPath:  "/v1/boards/{id}",
			method:        http.MethodGet,
			handlerStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.handlerStatus)
			})

			reg := prometheus.NewRegistry()
			mw := middleware.NewMetrics(reg)
			wrapped := mw.Wrap(handler)

			req, rr := testutil.NewJSONRequestAndRecorder(t, tt.method, tt.requestPath, "")

			wrapped.ServeHTTP(rr, req)

			responseStatus := strconv.Itoa(rr.Code)
			handlerStatusStr := strconv.Itoa(tt.handlerStatus)
			if responseStatus != handlerStatusStr {
				t.Errorf("recorder status %q, want %q (handler wrote %d)", responseStatus, handlerStatusStr, tt.handlerStatus)
			}

			requestCount := promTestutil.ToFloat64(mw.HttpRequests.With(prometheus.Labels{
				"path":   tt.abstractPath,
				"method": tt.method,
				"status": handlerStatusStr,
			}))

			if requestCount != 1 {
				t.Errorf("http_requests_total for path=%q method=%q status=%q: got %f, want 1",
					tt.abstractPath,
					tt.method,
					handlerStatusStr,
					requestCount,
				)
			}

			histogramSamples := promTestutil.CollectAndCount(mw.HttpDuration)
			if histogramSamples == 0 {
				t.Errorf("http_request_duration_seconds: no observations")
			}
		})
	}
}
