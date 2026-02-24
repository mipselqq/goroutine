package app_test

import "net/http"

type spyMetricsMiddleware struct{}

func (s *spyMetricsMiddleware) Wrap(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Metrics-Tracked", "true")
		next.ServeHTTP(w, r)
	}
}

type spyCorsMiddleware struct{}

func (s *spyCorsMiddleware) Wrap(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Cors-Tracked", "true")
		next.ServeHTTP(w, r)
	}
}
