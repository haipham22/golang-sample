package http

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// WithTimeout sets server timeouts.
func WithTimeout(read, write, idle time.Duration) ServerOption {
	return func(s *server) {
		s.server.ReadTimeout = read
		s.server.WriteTimeout = write
		s.server.IdleTimeout = idle
	}
}

// WithReadHeaderTimeout sets read header timeout.
func WithReadHeaderTimeout(timeout time.Duration) ServerOption {
	return func(s *server) {
		s.server.ReadHeaderTimeout = timeout
	}
}

// WithShutdownTimeout sets the maximum time to wait for connections to close.
func WithShutdownTimeout(timeout time.Duration) ServerOption {
	return func(s *server) {
		s.shutdownTimeout = timeout
	}
}

// WithLogger sets a custom logger.
func WithLogger(logger *zap.SugaredLogger) ServerOption {
	return func(s *server) {
		s.logger = logger
	}
}

// WithMaxHeaderBytes sets max header bytes.
func WithMaxHeaderBytes(bytes int) ServerOption {
	return func(s *server) {
		s.server.MaxHeaderBytes = bytes
	}
}

// WithBaseContext sets base context.
func WithBaseContext(fn func(net.Listener) context.Context) ServerOption {
	return func(s *server) {
		s.server.BaseContext = fn
	}
}

// WithTLS configures TLS.
func WithTLS(certFile, keyFile string) ServerOption {
	return func(s *server) {
		s.server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		// Note: CertFile and KeyFile are not part of http.server struct
		// They're passed directly to ListenAndServeTLS
	}
}

// WithHTTPServerOptions sets additional http.server options.
func WithHTTPServerOptions(opts ...func(*http.Server)) ServerOption {
	return func(s *server) {
		for _, opt := range opts {
			opt(s.server)
		}
	}
}
