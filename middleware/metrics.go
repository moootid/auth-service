package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "auth_service_http_request_duration_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"path", "method", "status_code"})

	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_service_http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"path", "method", "status_code"})
)

// responseWriter is a wrapper for http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// MetricsMiddleware measures the duration and counts the total number of HTTP requests.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)
		
		// Serve the request
		next.ServeHTTP(rw, r)

		// Get the route path template, e.g., /auth/profile/{id}
		route := mux.CurrentRoute(r)
		path := "unknown"
		if route != nil {
			path, _ = route.GetPathTemplate()
		}

		statusCode := strconv.Itoa(rw.statusCode)
		duration := time.Since(start).Seconds()

		// Record metrics
		httpDuration.WithLabelValues(path, r.Method, statusCode).Observe(duration)
		httpRequestsTotal.WithLabelValues(path, r.Method, statusCode).Inc()
	})
}