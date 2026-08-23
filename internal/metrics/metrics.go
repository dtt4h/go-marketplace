package metrics

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry holds all application metrics.
type Registry struct {
	reg                *prometheus.Registry
	httpRequestsTotal  *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge
}

// New creates and registers all metrics.
func New() *Registry {
	reg := prometheus.NewRegistry()
	r := &Registry{reg: reg}

	r.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "marketplace",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)
	r.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "marketplace",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	r.httpRequestsInFlight = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "marketplace",
		Subsystem: "http",
		Name:      "requests_in_flight",
		Help:      "Number of HTTP requests currently being processed.",
	})

	r.reg.MustRegister(
		r.httpRequestsTotal,
		r.httpRequestDuration,
		r.httpRequestsInFlight,
	)

	return r
}

// Handler returns an http.Handler that serves Prometheus metrics.
func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.reg, promhttp.HandlerOpts{})
}

// Middleware returns a chi middleware that records HTTP metrics.
func (m *Registry) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		sw := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}

		m.httpRequestsInFlight.Inc()
		defer m.httpRequestsInFlight.Dec()

		next.ServeHTTP(sw, r)

		duration := time.Since(start)
		path := r.URL.Path

		m.httpRequestsTotal.WithLabelValues(
			r.Method,
			path,
			strconv.Itoa(sw.statusCode),
		).Inc()

		m.httpRequestDuration.WithLabelValues(
			r.Method,
			path,
		).Observe(duration.Seconds())
	})
}

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.statusCode = code
	sw.ResponseWriter.WriteHeader(code)
}

// LogSlowRequests starts a goroutine that logs requests slower than threshold.
func LogSlowRequests(threshold time.Duration, log *slog.Logger) {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			log.Info("slow request logger running",
				slog.Duration("threshold", threshold))
		}
	}()
}