package cron

import (
	"context"
	"time"
)

// SchedulerSession provides job execution context.
// Pattern: sarama.ConsumerGroupSession
//
// Each job execution receives a fresh session with context about the current execution.
type SchedulerSession interface {
	// Context returns the execution context.
	// Use this for cancellation, deadlines, and request-scoped data.
	Context() context.Context

	// Scheduler returns the scheduler name running this job.
	Scheduler() string

	// Job returns the job name being executed.
	Job() string

	// LastExecution returns when this job last ran.
	// For the first execution, this will be the zero time.
	LastExecution() time.Time

	// ScheduledExecution returns when this execution was scheduled.
	ScheduledExecution() time.Time
}

// schedulerSession implements SchedulerSession
type schedulerSession struct {
	ctx                context.Context
	schedulerName      string
	jobName            string
	lastExecution      time.Time
	scheduledExecution time.Time
}

// NewSchedulerSession creates a new session for testing or internal use
func NewSchedulerSession(
	ctx context.Context,
	schedulerName string,
	jobName string,
	lastExecution time.Time,
	scheduledExecution time.Time,
) SchedulerSession {
	return &schedulerSession{
		ctx:                ctx,
		schedulerName:      schedulerName,
		jobName:            jobName,
		lastExecution:      lastExecution,
		scheduledExecution: scheduledExecution,
	}
}

// Context implements SchedulerSession
func (s *schedulerSession) Context() context.Context {
	return s.ctx
}

// Scheduler implements SchedulerSession
func (s *schedulerSession) Scheduler() string {
	return s.schedulerName
}

// Job implements SchedulerSession
func (s *schedulerSession) Job() string {
	return s.jobName
}

// LastExecution implements SchedulerSession
func (s *schedulerSession) LastExecution() time.Time {
	return s.lastExecution
}

// ScheduledExecution implements SchedulerSession
func (s *schedulerSession) ScheduledExecution() time.Time {
	return s.scheduledExecution
}
