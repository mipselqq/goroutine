package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"goroutine/internal/http/httpschema"
	"goroutine/internal/http/middleware"
	"goroutine/internal/testutil"
)

func TestRequestIDMiddleware(t *testing.T) {
	t.Parallel()

	mockStatusCode := http.StatusTeapot
	mockRequestID := "mock-request-id"
	requestRequestID := "pre-set-request-id"

	generateStaticID := func() string {
		return mockRequestID
	}

	tests := []struct {
		name           string
		requestHeaders map[string]string
		wantRequestID  string
	}{
		{
			name:           "No input header",
			requestHeaders: map[string]string{},
			wantRequestID:  mockRequestID,
		},
		{
			name: "Valid input header",
			requestHeaders: map[string]string{
				"X-Request-Id": requestRequestID,
			},
			wantRequestID: requestRequestID,
		},
		{
			name: "Empty input header",
			requestHeaders: map[string]string{
				"X-Request-Id": "",
			},
			wantRequestID: mockRequestID,
		},
		{
			name: "Whitespace input header",
			requestHeaders: map[string]string{
				"X-Request-Id": "   ",
			},
			wantRequestID: mockRequestID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.wantRequestID != "" {
					reqID, ok := r.Context().Value(httpschema.ContextKeyRequestID).(string)
					if !ok {
						t.Errorf("RequestID missing in context")
					}
					if reqID != tt.wantRequestID {
						t.Errorf("got request ID in context %q, want %q", reqID, tt.wantRequestID)
					}
				}
				w.WriteHeader(mockStatusCode)
			})

			mw := middleware.MustNewRequestID(testutil.NewLogger(t), generateStaticID)

			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			rr := httptest.NewRecorder()

			for k, v := range tt.requestHeaders {
				req.Header.Set(k, v)
			}

			wrapped := mw.Wrap(handler)
			wrapped.ServeHTTP(rr, req)

			testutil.AssertStatusCode(t, rr, mockStatusCode)

			if tt.wantRequestID != "" {
				headerID := rr.Header().Get("X-Request-ID")
				if headerID != tt.wantRequestID {
					t.Errorf("got response request ID %q, want %q", headerID, tt.wantRequestID)
				}
			}
		})
	}
}
