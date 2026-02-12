package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	HttpRequests *prometheus.CounterVec
	HttpDuration *prometheus.HistogramVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := Metrics{
		HttpRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"path", "method", "status"},
		),
		HttpDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"path", "method"},
		),
	}
	reg.MustRegister(m.HttpRequests, m.HttpDuration)

	return &m
}

type StatusSpyWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (w *StatusSpyWriter) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (m *Metrics) Wrap(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/metrics" || path == "/swagger" {
			next.ServeHTTP(w, r)
			return
		}

		startTime := time.Now()
		spy := &StatusSpyWriter{ResponseWriter: w, StatusCode: http.StatusOK}

		next.ServeHTTP(spy, r)

		duration := time.Since(startTime).Seconds()
		status := spy.StatusCode
		method := r.Method

		m.HttpRequests.WithLabelValues(path, method, strconv.Itoa(status)).Inc()
		m.HttpDuration.WithLabelValues(path, method).Observe(duration)
	})
}
