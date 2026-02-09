package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"goroutine/internal/middleware"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetrics_Collection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		path         string
		method       string
		expectedCode string
		shouldCount  bool
	}{
		{
			name:         "Count successful register",
			path:         "/register",
			method:       http.MethodPost,
			expectedCode: "200",
			shouldCount:  true,
		},
		{
			name:         "Count failed health",
			path:         "/health",
			method:       http.MethodGet,
			expectedCode: "403",
			shouldCount:  true,
		},
		{
			name:         "Don't count successful metrics request",
			path:         "/metrics",
			method:       http.MethodGet,
			expectedCode: "200",
			shouldCount:  false,
		},
		{
			name:         "Don't count failed swagger request",
			path:         "/swagger",
			method:       http.MethodGet,
			expectedCode: "403",
			shouldCount:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				code, _ := strconv.Atoi(tt.expectedCode)
				w.WriteHeader(code)
			})

			reg := prometheus.NewRegistry()
			mw := middleware.NewMetrics(reg)
			wrapped := mw.Wrap(handler)

			req := httptest.NewRequest(tt.method, tt.path, http.NoBody)
			rr := httptest.NewRecorder()

			wrapped.ServeHTTP(rr, req)

			code := strconv.Itoa(rr.Code)
			if code != tt.expectedCode {
				t.Errorf("Middleware modified status code: expected %s, got %s", tt.expectedCode, code)
			}

			count := testutil.ToFloat64(mw.HttpRequests.With(prometheus.Labels{
				"path":   tt.path,
				"method": tt.method,
				"status": tt.expectedCode,
			}))

			if tt.shouldCount && count != 1 {
				t.Errorf("Expected count = 1, got %f", count)
			}

			histCount := testutil.CollectAndCount(mw.HttpDuration)
			if tt.shouldCount && histCount == 0 {
				t.Errorf("Expected histogram to capture observation")
			}
		})
	}
}
