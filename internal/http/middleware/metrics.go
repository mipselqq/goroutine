package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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

func (m *Metrics) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		abstractPath := AbstractMetricPath(r.URL.Path)

		startTime := time.Now()
		spy := &StatusSpyWriter{ResponseWriter: w, StatusCode: http.StatusOK}

		next.ServeHTTP(spy, r)

		duration := time.Since(startTime).Seconds()
		responseStatus := spy.StatusCode
		method := r.Method

		m.HttpRequests.WithLabelValues(abstractPath, method, strconv.Itoa(responseStatus)).Inc()
		m.HttpDuration.WithLabelValues(abstractPath, method).Observe(duration)
	})
}

func AbstractMetricPath(rawPath string) string {
	segments := strings.Split(rawPath, "/")
	for i, segment := range segments {
		if segment == "" {
			continue
		}

		if _, err := uuid.Parse(segment); err == nil {
			segments[i] = "{id}"
		}
	}

	return strings.Join(segments, "/")
}
