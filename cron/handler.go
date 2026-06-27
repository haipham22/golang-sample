package cron

import (
	"context"
)

// JobHandler instances handle job execution with lifecycle hooks.
// Pattern: sarama.ConsumerGroupHandler
//
// Jobs using this interface have three lifecycle phases:
// 1. Setup - Run before job execution begins (initialize resources, validate state)
// 2. Execute - Run repeatedly on schedule (main job logic)
// 3. Cleanup - Run after job execution stops (release resources, flush buffers)
//
// PLEASE NOTE that handlers may be called from multiple goroutines concurrently
// if the scheduler allows overlapping job executions. Ensure that all state is
// safely protected against race conditions.
type JobHandler interface {
	// Setup is run before job execution begins.
	// Use this to initialize resources, validate state, prepare connections, etc.
	// Called once when the scheduler starts or when the job is first registered.
	//
	// Pattern: sarama.ConsumerGroupHandler.Setup()
	Setup(session SchedulerSession) error

	// Execute is called for each job execution (repeated on schedule).
	// This is where the main job logic goes.
	// Called repeatedly according to the job's schedule.
	//
	// Pattern: sarama.ConsumerGroupHandler.ConsumeClaim()
	Execute(session SchedulerSession) error

	// Cleanup is run after job execution stops.
	// Use this to release resources, close connections, flush buffers, etc.
	// Called once when the scheduler stops or when the job is removed.
	//
	// Pattern: sarama.ConsumerGroupHandler.Cleanup()
	Cleanup(session SchedulerSession) error
}

// JobHandlerFunc adapts a function to the JobHandler interface.
// Setup and Cleanup are no-ops, only Execute runs the provided function.
//
// Use this for simple jobs that don't need resource initialization/cleanup.
//
// Example:
//
//	handler := cron.JobHandlerFunc(func(ctx context.Context) error {
//	    log.Info("Running job")
//	    return doTask(ctx)
//	})
type JobHandlerFunc func(ctx context.Context) error

// Setup is a no-op for JobHandlerFunc
func (f JobHandlerFunc) Setup(session SchedulerSession) error {
	return nil
}

// Execute calls the underlying function
func (f JobHandlerFunc) Execute(session SchedulerSession) error {
	return f(session.Context())
}

// Cleanup is a no-op for JobHandlerFunc
func (f JobHandlerFunc) Cleanup(session SchedulerSession) error {
	return nil
}
