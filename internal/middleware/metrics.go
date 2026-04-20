package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests by method, path, and status.",
		},
		[]string{"method", "path", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	activeConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_active_connections",
			Help: "Current number of active HTTP connections.",
		},
	)
)

func init() {
	prometheus.MustRegister(requestsTotal, requestDuration, activeConnections)
}

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		activeConnections.Inc()
		defer activeConnections.Dec()

		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(rw.status)

		requestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}
