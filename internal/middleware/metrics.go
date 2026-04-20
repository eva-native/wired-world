package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func NewMetrics(reg prometheus.Registerer) func(http.Handler) http.Handler {
	requestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests by method, path, and status.",
		},
		[]string{"method", "path", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	activeRequests := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of in-flight HTTP requests.",
		},
	)

	reg.MustRegister(requestsTotal, requestDuration, activeRequests)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			activeRequests.Inc()
			defer activeRequests.Dec()

			start := time.Now()
			rw := wrapWriter(w)
			next.ServeHTTP(rw, r)

			pattern := routePattern(r)
			requestsTotal.WithLabelValues(r.Method, pattern, strconv.Itoa(rw.status)).Inc()
			requestDuration.WithLabelValues(r.Method, pattern).Observe(time.Since(start).Seconds())
		})
	}
}
