// Package job wires govern/cron JobHandlers behind a small Scheduler façade.
// It registers sample periodic jobs (e.g. a cleanup job) on a govern/cron
// Scheduler and exposes Start/Shutdown so a cobra command (workerd) can run
// them via graceful.Run.
package job

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	governcron "github.com/haipham22/govern/cron"
)

// CleanupInterval is how often the sample cleanup job runs.
const CleanupInterval = 5 * time.Minute

// Scheduler owns a govern/cron Scheduler and its registered jobs. The zero
// value is NOT usable — construct with New.
type Scheduler struct {
	log       *zap.SugaredLogger
	scheduler *governcron.Scheduler
	cleanup   func()
}

// New builds a govern/cron Scheduler, registers the sample jobs, and returns
// a Scheduler wrapping it together with a cleanup func the caller must defer.
func New(log *zap.SugaredLogger) (*Scheduler, error) {
	if log == nil {
		return nil, fmt.Errorf("job: nil logger")
	}
	s, cleanup, err := governcron.New(governcron.WithLogger(log))
	if err != nil {
		return nil, fmt.Errorf("job: create cron scheduler: %w", err)
	}

	sched := &Scheduler{log: log, scheduler: s, cleanup: cleanup}

	// Register the sample cleanup job. DurationJob takes a func + args; we use
	// a closure that captures the scheduler logger.
	if _, err := s.DurationJob(CleanupInterval, sched.runCleanup); err != nil {
		cleanup()
		return nil, fmt.Errorf("job: register cleanup job: %w", err)
	}
	return sched, nil
}

// Start implements graceful.Service by delegating to the govern scheduler.
func (s *Scheduler) Start(ctx context.Context) error {
	return s.scheduler.Start(ctx)
}

// Shutdown implements graceful.Service.
func (s *Scheduler) Shutdown(ctx context.Context) error {
	return s.scheduler.Shutdown(ctx)
}

// Cleanup releases scheduler resources. Safe to call multiple times.
func (s *Scheduler) Cleanup() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

// runCleanup is the sample periodic job body. A real implementation would
// delete stale sessions, expire tokens, etc. Kept dependency-free so the
// template stays self-contained.
func (s *Scheduler) runCleanup(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-ctx.Done():
		return
	default:
	}
	s.log.Infow("cleanup job tick", "ts", time.Now().UTC())
}
