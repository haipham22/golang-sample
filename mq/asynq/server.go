package asynq

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Server implements graceful.Service for asynq workers.
// It wraps asynq.Server with govern patterns for lifecycle management.
type Server struct {
	server    *asynq.Server
	mux       *TaskMux
	logger    *zap.SugaredLogger
	config    *Config
	mu        sync.RWMutex
	started   atomic.Bool
	shutdown  atomic.Bool
}

var _ gracefulService = (*Server)(nil)

// gracefulService is an interface to avoid import cycle with graceful package.
type gracefulService interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// NewServer creates a task queue server with govern patterns.
//
// The returned cleanup function closes the server and should be called
// when the server is no longer needed (typically via defer).
//
// Example:
//
//	mux := asynq.NewTaskMux()
//	mux.HandleFunc("email:send", handleEmail)
//
//	server, cleanup, _ := asynq.NewServer(redisClient, mux,
//	    asynq.WithConcurrency(10),
//	    asynq.WithLogger(logger),
//	)
//	defer cleanup()
//
//	// Use with graceful.Run
//	graceful.Run(ctx, logger, 30*time.Second, server)
func NewServer(redisClient redis.UniversalClient, mux *TaskMux, opts ...Option) (*Server, func(), error) {
	if mux == nil {
		return nil, nil, fmt.Errorf("mux cannot be nil")
	}

	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	// Convert redis.UniversalClient to asynq.RedisClientOpt
	asynqOpt := asynqRedisClientOpt(redisClient)

	// Build asynq config
	asynqCfg := asynq.Config{
		Concurrency: cfg.Concurrency,
		Queues:      cfg.Queues,
	}

	server := asynq.NewServer(asynqOpt, asynqCfg)

	s := &Server{
		server:   server,
		mux:      mux,
		logger:   zap.NewNop().Sugar(),
		config:   cfg,
		started:  atomic.Bool{},
		shutdown: atomic.Bool{},
	}

	cleanup := func() {
		if err := s.Close(); err != nil && s.logger != nil {
			s.logger.Errorf("Failed to close asynq server: %v", err)
		}
	}

	return s, cleanup, nil
}

// Start begins processing tasks.
// Implements graceful.Service.
// Blocks until Shutdown is called or an error occurs.
func (s *Server) Start(ctx context.Context) error {
	if !s.started.CompareAndSwap(false, true) {
		return fmt.Errorf("server already started")
	}

	s.logger.Infof("Starting asynq server: concurrency=%d queues=%v",
		s.config.Concurrency, s.config.Queues)

	// Run the server - this blocks until shutdown
	if err := s.server.Run(s.handlerAdapter()); err != nil {
		return fmt.Errorf("asynq server error: %w", err)
	}

	return nil
}

// Shutdown stops processing tasks gracefully.
// Implements graceful.Service.
// Waits for in-flight tasks to complete up to the configured timeout.
func (s *Server) Shutdown(ctx context.Context) error {
	if !s.shutdown.CompareAndSwap(false, true) {
		return fmt.Errorf("server already shutdown")
	}

	s.logger.Infof("Shutting down asynq server: timeout=%v", s.config.ShutdownTimeout)

	// asynq.Server.Shutdown() doesn't take context and returns no error
	// Start shutdown in background
	done := make(chan struct{})
	go func() {
		s.server.Shutdown()
		close(done)
	}()

	// Wait for shutdown or timeout
	select {
	case <-done:
		// Shutdown complete
	case <-time.After(s.config.ShutdownTimeout):
		return fmt.Errorf("asynq server shutdown timeout")
	}

	s.logger.Info("Asynq server shutdown complete")
	return nil
}

// Close closes the server connection.
// This is an alias for Shutdown with background context.
func (s *Server) Close() error {
	ctx := context.Background()
	return s.Shutdown(ctx)
}

// handlerAdapter converts the server's mux to asynq.Handler.
// TaskMux already implements asynq.Handler via ServeAsynqHandler.
func (s *Server) handlerAdapter() asynq.Handler {
	// TaskMux implements asynq.Handler
	return s.mux
}

// SetLogger sets the logger for the server.
func (s *Server) SetLogger(logger *zap.SugaredLogger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger = logger
}

// IsStarted returns true if the server has been started.
func (s *Server) IsStarted() bool {
	return s.started.Load()
}

// IsShutdown returns true if the server has been shutdown.
func (s *Server) IsShutdown() bool {
	return s.shutdown.Load()
}
