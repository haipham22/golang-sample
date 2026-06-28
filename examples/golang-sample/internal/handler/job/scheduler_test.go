package job

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNew_ConstructsScheduler(t *testing.T) {
	s, err := New(zap.NewNop().Sugar())
	require.NoError(t, err)
	require.NotNil(t, s)
	t.Cleanup(s.Cleanup)
}

func TestNew_NilLogger(t *testing.T) {
	_, err := New(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil logger")
}

// TestScheduler_GracefulServiceLifecycle exercises Start/Shutdown once to
// confirm the Scheduler satisfies graceful.Service and the gocron scheduler
// transitions cleanly. Start is non-blocking per govern/cron.
func TestScheduler_GracefulServiceLifecycle(t *testing.T) {
	s, err := New(zap.NewNop().Sugar())
	require.NoError(t, err)
	t.Cleanup(s.Cleanup)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	require.NoError(t, s.Start(ctx))
	// Double-start must error (govern/cron guards with CompareAndSwap).
	err = s.Start(ctx)
	require.Error(t, err)

	require.NoError(t, s.Shutdown(ctx))
}

func TestScheduler_runCleanup_NoPanic(t *testing.T) {
	s, err := New(zap.NewNop().Sugar())
	require.NoError(t, err)
	t.Cleanup(s.Cleanup)

	// runCleanup must be safe with both real and background contexts.
	assert.NotPanics(t, func() { s.runCleanup(context.Background()) })
	assert.NotPanics(t, func() { s.runCleanup(nil) })
}

func TestScheduler_Cleanup_Idempotent(t *testing.T) {
	s, err := New(zap.NewNop().Sugar())
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		s.Cleanup()
		s.Cleanup() // second call must not panic
	})
}
