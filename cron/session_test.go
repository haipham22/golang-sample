package cron

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSchedulerSession(t *testing.T) {
	t.Run("NewSchedulerSession creates valid session", func(t *testing.T) {
		ctx := context.Background()
		schedulerName := "test-scheduler"
		jobName := "test-job"
		lastExecution := time.Now().Add(-1 * time.Hour)
		scheduledExecution := time.Now()

		session := NewSchedulerSession(ctx, schedulerName, jobName, lastExecution, scheduledExecution)

		assert.NotNil(t, session)
	})

	t.Run("Context returns the provided context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), contextKey("key"), "value")
		session := NewSchedulerSession(ctx, "test", "job1", time.Time{}, time.Now())

		assert.Equal(t, "value", session.Context().Value(contextKey("key")))
	})

	t.Run("Scheduler returns the scheduler name", func(t *testing.T) {
		session := NewSchedulerSession(context.Background(), "my-scheduler", "job1", time.Time{}, time.Now())

		assert.Equal(t, "my-scheduler", session.Scheduler())
	})

	t.Run("Job returns the job name", func(t *testing.T) {
		session := NewSchedulerSession(context.Background(), "test", "my-job", time.Time{}, time.Now())

		assert.Equal(t, "my-job", session.Job())
	})

	t.Run("LastExecution returns the last execution time", func(t *testing.T) {
		lastExec := time.Now().Add(-1 * time.Hour)
		session := NewSchedulerSession(context.Background(), "test", "job1", lastExec, time.Now())

		assert.Equal(t, lastExec, session.LastExecution())
	})

	t.Run("LastExecution returns zero time for first execution", func(t *testing.T) {
		session := NewSchedulerSession(context.Background(), "test", "job1", time.Time{}, time.Now())

		assert.True(t, session.LastExecution().IsZero())
	})

	t.Run("ScheduledExecution returns the scheduled execution time", func(t *testing.T) {
		scheduled := time.Now()
		session := NewSchedulerSession(context.Background(), "test", "job1", time.Time{}, scheduled)

		assert.Equal(t, scheduled, session.ScheduledExecution())
	})

	t.Run("multiple sessions have independent values", func(t *testing.T) {
		ctx1 := context.WithValue(context.Background(), contextKey("key"), "value1")
		ctx2 := context.WithValue(context.Background(), contextKey("key"), "value2")

		session1 := NewSchedulerSession(ctx1, "scheduler1", "job1", time.Now().Add(-2*time.Hour), time.Now())
		session2 := NewSchedulerSession(ctx2, "scheduler2", "job2", time.Now().Add(-1*time.Hour), time.Now().Add(1*time.Hour))

		assert.Equal(t, "value1", session1.Context().Value(contextKey("key")))
		assert.Equal(t, "value2", session2.Context().Value(contextKey("key")))
		assert.Equal(t, "scheduler1", session1.Scheduler())
		assert.Equal(t, "scheduler2", session2.Scheduler())
		assert.Equal(t, "job1", session1.Job())
		assert.Equal(t, "job2", session2.Job())
	})
}
