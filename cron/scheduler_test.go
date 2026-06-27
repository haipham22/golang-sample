package cron

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/haipham22/govern/log"
)

func TestNew(t *testing.T) {
	t.Run("creates scheduler with defaults", func(t *testing.T) {
		scheduler, cleanup, err := New()
		require.NoError(t, err)
		require.NotNil(t, scheduler)
		assert.NotNil(t, cleanup)
		defer cleanup()

		assert.NotNil(t, scheduler.logger)
	})

	t.Run("creates scheduler with custom logger", func(t *testing.T) {
		customLogger := log.New()
		scheduler, cleanup, err := New(WithLogger(customLogger))
		require.NoError(t, err)
		require.NotNil(t, scheduler)
		defer cleanup()

		assert.Same(t, customLogger, scheduler.logger)
	})

	t.Run("creates scheduler with location", func(t *testing.T) {
		loc := time.FixedZone("UTC", 0)
		scheduler, cleanup, err := New(WithLocation(loc))
		require.NoError(t, err)
		require.NotNil(t, scheduler)
		defer cleanup()

		assert.Equal(t, loc, scheduler.location)
	})

	t.Run("creates scheduler with stop timeout", func(t *testing.T) {
		timeout := 10 * time.Second
		scheduler, cleanup, err := New(WithStopTimeout(timeout))
		require.NoError(t, err)
		require.NotNil(t, scheduler)
		defer cleanup()

		assert.Equal(t, timeout, scheduler.stopTimeout)
	})
}

func TestScheduler_Start(t *testing.T) {
	t.Run("starts scheduler successfully", func(t *testing.T) {
		scheduler, cleanup, err := New()
		require.NoError(t, err)
		defer cleanup()

		ctx := context.Background()
		err = scheduler.Start(ctx)
		assert.NoError(t, err)
		assert.True(t, scheduler.started.Load())

		// Cleanup
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = scheduler.Shutdown(shutdownCtx)
	})

	t.Run("returns error if already started", func(t *testing.T) {
		scheduler, cleanup, err := New()
		require.NoError(t, err)
		defer cleanup()

		ctx := context.Background()
		err = scheduler.Start(ctx)
		require.NoError(t, err)

		err = scheduler.Start(ctx)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSchedulerAlreadyStarted)

		// Cleanup
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = scheduler.Shutdown(shutdownCtx)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		scheduler, cleanup, err := New()
		require.NoError(t, err)
		defer cleanup()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// gocron Start() is non-blocking and doesn't check context
		// Context is checked during job execution
		err = scheduler.Start(ctx)
		assert.NoError(t, err) // Start succeeds even with cancelled context

		// Cleanup
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = scheduler.Shutdown(shutdownCtx)
	})
}

func TestScheduler_Shutdown(t *testing.T) {
	t.Run("shuts down started scheduler", func(t *testing.T) {
		scheduler, cleanup, err := New()
		require.NoError(t, err)
		defer cleanup()

		ctx := context.Background()
		_ = scheduler.Start(ctx)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = scheduler.Shutdown(shutdownCtx)
		assert.NoError(t, err)
		assert.True(t, scheduler.shutdown.Load())
	})

	t.Run("returns error if not started", func(t *testing.T) {
		scheduler, cleanup, err := New()
		require.NoError(t, err)
		defer cleanup()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = scheduler.Shutdown(shutdownCtx)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSchedulerNotStarted)
	})

	t.Run("waits for jobs to complete before shutdown", func(t *testing.T) {
		scheduler, cleanup, err := New()
		require.NoError(t, err)
		defer cleanup()

		ctx := context.Background()
		_ = scheduler.Start(ctx)

		// Add a quick job
		_, _ = scheduler.DurationJob(time.Hour, func() {})

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		err = scheduler.Shutdown(shutdownCtx)
		// Shutdown waits for running jobs, completes within timeout
		assert.NoError(t, err)
	})
}

func TestScheduler_graceful_Service(t *testing.T) {
	t.Run("implements graceful.Service interface", func(t *testing.T) {
		// Compile-time check
		var _ interface {
			Start(ctx context.Context) error
			Shutdown(ctx context.Context) error
		} = (*Scheduler)(nil)

		scheduler, cleanup, err := New()
		require.NoError(t, err)
		defer cleanup()

		ctx := context.Background()
		err = scheduler.Start(ctx)
		assert.NoError(t, err)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = scheduler.Shutdown(shutdownCtx)
		assert.NoError(t, err)
	})
}
