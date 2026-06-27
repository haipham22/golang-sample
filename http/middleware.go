package http

import (
	"go.uber.org/zap"

	"github.com/haipham22/govern/http/jwt"
	"github.com/haipham22/govern/http/middleware"
)

// WithJWT adds JWT authentication middleware.
func WithJWT(config *jwt.MiddlewareConfig) ServerOption {
	return func(s *server) {
		s.Use(jwt.Middleware(config))
	}
}

// WithLogging adds request/response logging.
func WithLogging(logger *zap.SugaredLogger) ServerOption {
	return func(s *server) {
		s.Use(Middleware(middleware.Logging(logger)))
	}
}

// WithRecovery adds panic recovery.
func WithRecovery(logger *zap.SugaredLogger) ServerOption {
	return func(s *server) {
		s.Use(Middleware(middleware.Recovery(logger)))
	}
}

// WithCORS adds CORS support.
func WithCORS(config *middleware.CORSConfig) ServerOption {
	return func(s *server) {
		s.Use(Middleware(middleware.CORS(config)))
	}
}

// WithRequestID adds request ID generation.
func WithRequestID() ServerOption {
	return func(s *server) {
		s.Use(Middleware(middleware.RequestID()))
	}
}
