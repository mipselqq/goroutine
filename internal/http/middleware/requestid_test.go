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

	// TODO: implement this side effects check in other middlewares
	headersToModify := []string{"X-Request-ID"}
	mockStatusCode := http.StatusTeapot
	mockRequestID := "mock-request-id"
	requestRequestID := "pre-set-request-id"

	generateStaticID := func() string {
		return mockRequestID
	}

	tests := []struct {
		name              string
		requestHeaders    map[string]string
		expectedRequestID string
	}{
		{
			name:              "No input header",
			requestHeaders:    map[string]string{},
			expectedRequestID: mockRequestID,
		},
		{
			name: "Valid input header",
			requestHeaders: map[string]string{
				"X-Request-Id": requestRequestID,
			},
			expectedRequestID: requestRequestID,
		},
		{
			name: "Empty input header",
			requestHeaders: map[string]string{
				"X-Request-Id": "",
			},
			expectedRequestID: mockRequestID,
		},
		{
			name: "Whitespace input header",
			requestHeaders: map[string]string{
				"X-Request-Id": "   ",
			},
			expectedRequestID: mockRequestID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectedRequestID != "" {
					reqID, ok := r.Context().Value(httpschema.ContextKeyRequestID).(string)
					if !ok {
						t.Errorf("RequestID missing in context")
					}
					if reqID != tt.expectedRequestID {
						t.Errorf("expected request ID in context %q, got %q", tt.expectedRequestID, reqID)
					}
				}
				w.WriteHeader(mockStatusCode)
			})

			mw := middleware.NewRequestID(testutil.NewTestLogger(t), generateStaticID)

			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			for k, v := range tt.requestHeaders {
				req.Header.Set(k, v)
			}

			wrapped := mw.Wrap(handler)
			rr := httptest.NewRecorder()

			wrapped.ServeHTTP(rr, req)

			if rr.Code != mockStatusCode {
				t.Errorf("Middleware changed status code, expected %d, got %d", mockStatusCode, rr.Code)
			}

			testutil.EnsureNoUnexpectedHeadersModified(t, rr, headersToModify)

			if tt.expectedRequestID != "" {
				headerID := rr.Header().Get("X-Request-Id")
				if headerID != tt.expectedRequestID {
					t.Errorf("expected response request ID %q, got %q", tt.expectedRequestID, headerID)
				}
			}
		})
	}
}
