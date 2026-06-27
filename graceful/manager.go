package graceful

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/haipham22/govern/log"
)

// Cleanup is a shutdown hook. It receives a context with deadline.
type Cleanup func(ctx context.Context) error

type Option func(*Manager)

// WithFailFast: if any managed goroutine returns error, begin shutdown immediately.
func WithFailFast(v bool) Option {
	return func(m *Manager) { m.failFast = v }
}

// WithLogger sets a custom logger for the Manager.
func WithLogger(logger *zap.SugaredLogger) Option {
	return func(m *Manager) { m.logger = logger }
}

// Manager coordinates goroutines + cleanup during shutdown.
type Manager struct {
	ctx    context.Context
	cancel context.CancelFunc
	stop   context.CancelFunc // stop signal.NotifyContext

	wg       sync.WaitGroup
	errOnce  sync.Once
	errCh    chan error
	failFast bool
	logger   *zap.SugaredLogger

	mu       sync.Mutex
	cleanups []Cleanup
	once     sync.Once
}

// NewManager New creates a Manager with a signal-aware context.
// Default signals: SIGINT, SIGTERM.
func NewManager(parent context.Context, opts ...Option) *Manager {
	if parent == nil {
		parent = context.Background()
	}
	sigCtx, stop := signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(sigCtx)

	m := &Manager{
		ctx:      ctx,
		cancel:   cancel,
		stop:     stop,
		errCh:    make(chan error, 1),
		failFast: true,
		logger:   log.Default(),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// Context returns the root context canceled on shutdown.
func (m *Manager) Context() context.Context {
	return m.ctx
}

// InitiateShutdown cancels the root context exactly once and stops signal notifier.
func (m *Manager) InitiateShutdown() {
	m.once.Do(func() {
		m.logger.Debug("Initiating shutdown")
		if m.stop != nil {
			m.stop()
		}
		m.cancel()
	})
}

// Go runs fn in a managed goroutine.
// fn MUST respect ctx.Done().
func (m *Manager) Go(fn func(ctx context.Context) error) {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		if err := fn(m.ctx); err != nil && !errors.Is(err, context.Canceled) {
			m.errOnce.Do(func() {
				m.logger.Errorw("Goroutine error", "error", err)
				m.errCh <- err
			})
			if m.failFast {
				m.InitiateShutdown()
			}
		}
	}()
}

// Defer registers a cleanup hook. Runs in LIFO order.
func (m *Manager) Defer(c Cleanup) {
	m.mu.Lock()
	m.cleanups = append(m.cleanups, c)
	m.mu.Unlock()
}

// Wait blocks until context canceled (signal/manual) or first goroutine error arrives.
// Returns first goroutine error (if any).
func (m *Manager) Wait() error {
	// Prioritize error over context cancellation
	select {
	case err := <-m.errCh:
		m.logger.Debug("Goroutine error received")
		return err
	case <-m.ctx.Done():
		m.logger.Debug("Context canceled, stopping wait")
	}

	// Check if error arrived after context cancellation
	select {
	case err := <-m.errCh:
		return err
	default:
		return nil
	}
}

// Shutdown performs graceful shutdown in 3 steps:
//  1. cancels context (idempotent)
//  2. waits for goroutines (up to timeout)
//  3. runs cleanup hooks with deadline context
func (m *Manager) Shutdown(timeout time.Duration) error {
	m.InitiateShutdown()

	// wait goroutines
	waitDone := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(waitDone)
	}()

	var waitErr error
	if timeout > 0 {
		select {
		case <-waitDone:
			m.logger.Debug("All goroutines stopped")
		case <-time.After(timeout):
			m.logger.Warnw("Shutdown timeout waiting for goroutines")
			waitErr = context.DeadlineExceeded
		}
	} else {
		<-waitDone
		m.logger.Debug("All goroutines stopped")
	}

	// cleanup ctx
	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// run cleanups LIFO
	m.mu.Lock()
	cleanups := make([]Cleanup, len(m.cleanups))
	copy(cleanups, m.cleanups)
	m.mu.Unlock()

	var cleanupErr error
	for i := len(cleanups) - 1; i >= 0; i-- {
		if err := cleanups[i](ctx); err != nil && cleanupErr == nil {
			m.logger.Errorw("Cleanup error", "error", err)
			cleanupErr = err
		}
	}

	if cleanupErr != nil {
		return cleanupErr
	}
	return waitErr
}

// ExitOnSignal: waits, shuts down, then exits (0 on success, 1 on error).
func ExitOnSignal(m *Manager, timeout time.Duration) {
	err := m.Wait()
	shErr := m.Shutdown(timeout)
	if err == nil {
		err = shErr
	}
	if err != nil {
		_, _ = os.Stderr.WriteString("shutdown error: " + err.Error() + "\n")
		os.Exit(1)
	}
	os.Exit(0)
}
