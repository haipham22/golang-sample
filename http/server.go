package http

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/haipham22/govern/graceful"
	"github.com/haipham22/govern/log"
)

type Server interface {
	graceful.Service

	Server() *http.Server
	Listen() (net.Listener, error)
	Use(middleware ...Middleware)
}

// Middleware function type
type Middleware func(http.Handler) http.Handler

// server wraps an http.server with graceful shutdown capabilities.
type server struct {
	server          *http.Server
	shutdownTimeout time.Duration
	logger          *zap.SugaredLogger
	manager         *graceful.Manager
	middlewares     []Middleware
}

// ServerOption configures a server.
type ServerOption func(*server)

// NewServer creates a new server with graceful shutdown support.
func NewServer(addr string, handler http.Handler, opts ...ServerOption) Server {
	s := &server{
		server: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       60 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
		},
		shutdownTimeout: 30 * time.Second,
		logger:          log.Default(),
		middlewares:     []Middleware{},
	}

	for _, opt := range opts {
		opt(s)
	}

	s.manager = graceful.NewManager(context.Background())

	return s
}

// Use adds middleware to the server.
func (s *server) Use(middleware ...Middleware) {
	s.middlewares = append(s.middlewares, middleware...)
}

// buildHandler builds the final handler with middleware chain.
func (s *server) buildHandler() http.Handler {
	handler := s.server.Handler
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		handler = s.middlewares[i](handler)
	}
	return handler
}

// Start begins serving and blocks until the server is shut down gracefully.
func (s *server) Start(_ context.Context) error {
	// Build handler with middleware
	s.server.Handler = s.buildHandler()

	// Register server shutdown as cleanup
	s.manager.Defer(func(ctx context.Context) error {
		s.logger.Info("Shutting down HTTP server")
		return s.server.Shutdown(ctx)
	})

	// Start a server in a managed goroutine
	s.manager.Go(func(ctx context.Context) error {
		s.logger.Infow("server starting", "addr", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Errorw("server error", "error", err)
			return err
		}
		return nil
	})

	// Wait for shutdown signal
	if err := s.manager.Wait(); err != nil {
		return err
	}

	return s.manager.Shutdown(s.shutdownTimeout)
}

// Shutdown triggers a graceful shutdown programmatically.
func (s *server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *server) Server() *http.Server {
	return s.server
}

// Listen returns the listener for testing purposes.
func (s *server) Listen() (net.Listener, error) {
	return net.Listen("tcp", s.server.Addr)
}
