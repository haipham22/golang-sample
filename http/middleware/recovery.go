package middleware

import (
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
)

// Recovery returns panic recovery middleware.
func Recovery(logger *zap.SugaredLogger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Errorw("Recovered from panic",
						"error", err,
						"stack", debug.Stack(),
						"request_id", r.Context().Value(RequestIDKeyVal),
						"method", r.Method,
						"path", r.URL.Path,
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
