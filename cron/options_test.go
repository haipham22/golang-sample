package cron

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/haipham22/govern/log"
)

func TestOptions(t *testing.T) {
	t.Run("WithLogger sets custom logger", func(t *testing.T) {
		customLogger := log.New()
		scheduler, cleanup, err := New(WithLogger(customLogger))
		require.NoError(t, err)
		require.NotNil(t, scheduler)
		defer cleanup()

		assert.Same(t, customLogger, scheduler.logger)
	})

	t.Run("WithLocation sets custom location", func(t *testing.T) {
		loc := time.FixedZone("UTC", 0)
		scheduler, cleanup, err := New(WithLocation(loc))
		require.NoError(t, err)
		require.NotNil(t, scheduler)
		defer cleanup()

		assert.Equal(t, loc, scheduler.location)
	})

	t.Run("WithStopTimeout sets custom timeout", func(t *testing.T) {
		timeout := 10 * time.Second
		scheduler, cleanup, err := New(WithStopTimeout(timeout))
		require.NoError(t, err)
		require.NotNil(t, scheduler)
		defer cleanup()

		assert.Equal(t, timeout, scheduler.stopTimeout)
	})

	t.Run("multiple options work together", func(t *testing.T) {
		customLogger := log.New()
		loc := time.FixedZone("UTC", 0)
		timeout := 5 * time.Second

		scheduler, cleanup, err := New(
			WithLogger(customLogger),
			WithLocation(loc),
			WithStopTimeout(timeout),
		)
		require.NoError(t, err)
		require.NotNil(t, scheduler)
		defer cleanup()

		assert.Same(t, customLogger, scheduler.logger)
		assert.Equal(t, loc, scheduler.location)
		assert.Equal(t, timeout, scheduler.stopTimeout)
	})
}
