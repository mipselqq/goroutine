package middleware_test

import (
	"net/http"
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

			mw := middleware.MustNewRequestID(testutil.NewTestLogger(t), generateStaticID)

			req, rr := testutil.NewJSONRequestAndRecorder(t, http.MethodGet, "/", http.NoBody)
			for k, v := range tt.requestHeaders {
				req.Header.Set(k, v)
			}

			wrapped := mw.Wrap(handler)
			wrapped.ServeHTTP(rr, req)

			testutil.AssertStatusCode(t, rr, mockStatusCode)

			if tt.expectedRequestID != "" {
				headerID := rr.Header().Get("X-Request-ID")
				if headerID != tt.expectedRequestID {
					t.Errorf("expected response request ID %q, got %q", tt.expectedRequestID, headerID)
				}
			}
		})
	}
}
