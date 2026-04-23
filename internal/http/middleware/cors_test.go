package middleware_test

import (
	"maps"
	"net/http"
	"strings"
	"testing"

	"goroutine/internal/config"
	"goroutine/internal/http/middleware"
	"goroutine/internal/testutil"
)

const (
	goodSite    = "http://good-site.com"
	awesomeSite = "http://awesome-site.com"
	evilSite    = "http://evil-site.com"
)

func TestCors_Wrap(t *testing.T) {
	originsStr := goodSite + "," + awesomeSite
	allowedOrigins := config.ParseAllowedOrigins(originsStr)
	originsWithAny := config.ParseAllowedOrigins(originsStr + ",*")

	filledGoodCORSHeaders := map[string]string{
		"Access-Control-Allow-Origin":      goodSite,
		"Access-Control-Allow-Methods":     "DELETE, GET, OPTIONS, PATCH, POST, PUT",
		"Access-Control-Allow-Headers":     "Content-Type, Authorization",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Max-Age":           "86400",
		"Vary":                             "Origin",
	}

	filledEvilCORSHeaders := maps.Clone(filledGoodCORSHeaders)
	filledEvilCORSHeaders["Access-Control-Allow-Origin"] = evilSite

	emptyCORSHeaders := map[string]string{
		"Access-Control-Allow-Origin":      "",
		"Access-Control-Allow-Methods":     "",
		"Access-Control-Allow-Headers":     "",
		"Access-Control-Allow-Credentials": "",
		"Access-Control-Max-Age":           "",
		"Vary":                             "Origin",
	}
	t.Parallel()

	tests := []struct {
		name           string
		method         string
		allowedOrigins map[string]struct{}
		reqHeaders     map[string]string
		wantStatus     int
		wantHeaders    map[string]string
	}{
		{
			name:           "No Origin Header",
			method:         "GET",
			allowedOrigins: allowedOrigins,
			reqHeaders:     map[string]string{
				// Not a browser
			},
			wantStatus:  http.StatusTeapot,
			wantHeaders: map[string]string{},
		},
		{
			name:           "Origin Header is not allowed",
			method:         "GET",
			allowedOrigins: allowedOrigins,
			reqHeaders: map[string]string{
				"Origin": evilSite,
			},
			wantStatus:  http.StatusForbidden,
			wantHeaders: emptyCORSHeaders,
		},
		{
			name:           "Origin Header is allowed",
			method:         "GET",
			allowedOrigins: allowedOrigins,
			reqHeaders: map[string]string{
				"Origin": goodSite,
			},
			wantStatus:  http.StatusNoContent,
			wantHeaders: filledGoodCORSHeaders,
		},
		{
			name:           "Preflight rejected",
			method:         "OPTIONS",
			allowedOrigins: allowedOrigins,
			reqHeaders: map[string]string{
				"Origin": evilSite,
			},
			wantStatus:  http.StatusForbidden,
			wantHeaders: emptyCORSHeaders,
		},
		{
			name:           "Preflight allowed",
			method:         "OPTIONS",
			allowedOrigins: allowedOrigins,
			reqHeaders: map[string]string{
				"Origin": goodSite,
			},
			wantStatus:  http.StatusNoContent,
			wantHeaders: filledGoodCORSHeaders,
		},
		{
			name:           "Any origin allowed",
			method:         "GET",
			allowedOrigins: originsWithAny,
			reqHeaders: map[string]string{
				"Origin": evilSite,
			},
			wantStatus:  http.StatusTeapot,
			wantHeaders: filledEvilCORSHeaders,
		},
		{
			name:           "No allowed origins",
			method:         "GET",
			allowedOrigins: map[string]struct{}{},
			reqHeaders: map[string]string{
				"Origin": goodSite,
			},
			wantStatus:  http.StatusForbidden,
			wantHeaders: emptyCORSHeaders,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.wantStatus)
			})

			mw := middleware.NewCORS(testutil.NewTestLogger(t), tt.allowedOrigins)

			req, rr := testutil.NewJSONRequestAndRecorder(t, tt.method, "/", "")
			for k, v := range tt.reqHeaders {
				req.Header.Set(k, v)
			}

			wrapped := mw.Wrap(handler)
			wrapped.ServeHTTP(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantStatus)
			for k, v := range tt.wantHeaders {
				if rr.Header().Get(k) != v {
					t.Errorf("got header %q=%q, want %q", k, rr.Header().Get(k), v)
				}
			}
		})
	}
}

func TestCors_WarnsAnyOriginAllowed(t *testing.T) {
	t.Parallel()

	logger, buf := testutil.NewBufJsonLogger(t)

	_ = middleware.NewCORS(logger, config.ParseAllowedOrigins(goodSite+",*,"+awesomeSite))

	if !strings.Contains(buf.String(), "too permissive") {
		t.Errorf("got log output %q, want mention of %q", buf.String(), "too permissive")
	}
}
