package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// HTTPMiddleware returns an http.Handler middleware that records request metrics.
// It records: http_requests_total, http_request_duration_seconds, and http_response_size_bytes.
// Metrics are registered to the default registry.
func HTTPMiddleware(handler http.Handler, labels ...string) http.Handler {
	return HTTPMiddlewareWithRegistry(Default(), handler, labels...)
}

// HTTPMiddlewareWithRegistry returns an http.Handler middleware that records request metrics
// to the specified registry. Use this in tests to avoid duplicate metric registration.
func HTTPMiddlewareWithRegistry(reg *Registry, handler http.Handler, labels ...string) http.Handler {
	requestsTotal := NewCounter("http_requests_total", "Total HTTP requests", append([]string{"method", "status"}, labels...))
	reg.MustRegister(requestsTotal.vec)

	duration := NewHistogram("http_request_duration_seconds", "HTTP request duration", append([]string{"method", "status"}, labels...), nil)
	reg.MustRegister(duration.vec)

	size := NewHistogram("http_response_size_bytes", "HTTP response size", append([]string{"method", "status"}, labels...), prometheus.ExponentialBuckets(100, 10, 8))
	reg.MustRegister(size.vec)

	return &middleware{
		handler:       handler,
		requestsTotal: requestsTotal,
		duration:      duration,
		size:          size,
		staticLabels:  labels,
	}
}

type middleware struct {
	handler       http.Handler
	requestsTotal *Counter
	duration      *Histogram
	size          *Histogram
	staticLabels  []string
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
	m.handler.ServeHTTP(rw, r)

	labelValues := append([]string{r.Method, http.StatusText(rw.status)}, m.staticLabels...)
	m.requestsTotal.Inc(labelValues...)
	m.duration.Observe(time.Since(start).Seconds(), labelValues...)
	if rw.written > 0 {
		m.size.Observe(float64(rw.written), labelValues...)
	}
}

// responseWriter wraps http.ResponseWriter to capture status and size.
type responseWriter struct {
	http.ResponseWriter
	status  int
	written int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += n
	return n, err
}
