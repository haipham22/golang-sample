package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequestID adds unique request ID to each request.
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			ctx := context.WithValue(r.Context(), RequestIDKeyVal, requestID)
			w.Header().Set("X-Request-ID", requestID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
