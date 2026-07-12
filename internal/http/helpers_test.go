package http_test

import (
	"net/http"
)

type spyMetricsMiddleware struct{}

func (s *spyMetricsMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Metrics-Tracked", "true")
		next.ServeHTTP(w, r)
	})
}

type spyCorsMiddleware struct{}

func (s *spyCorsMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Cors-Tracked", "true")
		next.ServeHTTP(w, r)
	})
}

type spyAuthMiddleware struct{}

func (s *spyAuthMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Auth-Tracked", "true")
		next.ServeHTTP(w, r)
	})
}

type spyRequestIDMiddleware struct{}

func (s *spyRequestIDMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RequestId-Tracked", "true")
		next.ServeHTTP(w, r)
	})
}

type spyTimeoutMiddleware struct{}

func (s *spyTimeoutMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Timeout-Tracked", "true")
		next.ServeHTTP(w, r)
	})
}
