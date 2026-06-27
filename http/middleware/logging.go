package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestIDKey is context key for request ID.
type RequestIDKey string

// RequestIDKeyVal is the context key value.
const RequestIDKeyVal RequestIDKey = "request_id"

// Logging returns request/response logging middleware.
func Logging(logger *zap.SugaredLogger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Add request ID to context
			ctx := context.WithValue(r.Context(), RequestIDKeyVal, requestID)
			r = r.WithContext(ctx)

			// Wrap response writer
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			// Log request
			logger.Infow("Incoming request",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)

			// Call next
			next.ServeHTTP(rw, r)

			// Log response
			logger.Infow("Request completed",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"duration", time.Since(start),
				"bytes_written", rw.bytesWritten,
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status.
type responseWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int
	wroteHeader  bool
}

func (rw *responseWriter) WriteHeader(status int) {
	if !rw.wroteHeader {
		rw.status = status
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(status)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}
