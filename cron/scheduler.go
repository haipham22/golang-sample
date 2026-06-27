package cron

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"

	"github.com/haipham22/govern/log"
)

// Scheduler wraps gocron.Scheduler and implements graceful.Service
type Scheduler struct {
	scheduler   gocron.Scheduler
	logger      *zap.SugaredLogger
	location    *time.Location
	stopTimeout time.Duration
	started     atomic.Bool
	shutdown    atomic.Bool
}

// New creates a new cron scheduler with given options
func New(opts ...Option) (*Scheduler, func(), error) {
	cfg := &Config{
		Logger:      log.Default(),
		Location:    time.Local,
		StopTimeout: 30 * time.Second,
	}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, nil, err
		}
	}

	// Create gocron scheduler with location
	gocronOpts := []gocron.SchedulerOption{
		gocron.WithLocation(cfg.Location),
	}

	s, _ := gocron.NewScheduler(gocronOpts...)

	scheduler := &Scheduler{
		scheduler:   s,
		logger:      cfg.Logger,
		location:    cfg.Location,
		stopTimeout: cfg.StopTimeout,
	}

	cleanup := func() {
		// Cleanup function for resource cleanup if needed
	}

	return scheduler, cleanup, nil
}

// Start begins the scheduler (non-blocking)
func (s *Scheduler) Start(_ context.Context) error {
	if !s.started.CompareAndSwap(false, true) {
		return ErrSchedulerAlreadyStarted
	}

	s.logger.Info("Starting cron scheduler")

	// Start the gocron scheduler
	s.scheduler.Start()

	s.logger.Info("Cron scheduler started")

	return nil
}

// Shutdown gracefully stops the scheduler
func (s *Scheduler) Shutdown(ctx context.Context) error {
	if !s.started.Load() {
		return ErrSchedulerNotStarted
	}

	if !s.shutdown.CompareAndSwap(false, true) {
		return ErrSchedulerAlreadyShutdown
	}

	s.logger.Info("Shutting down cron scheduler")

	// Shutdown gocron scheduler
	err := s.scheduler.Shutdown()
	if err != nil {
		s.logger.Errorf("Error shutting down scheduler: %v", err)
		return err
	}

	s.logger.Info("Cron scheduler shutdown complete")

	return nil
}

// DurationJob creates a duration-based job
func (s *Scheduler) DurationJob(d time.Duration, fn any, args ...any) (gocron.Job, error) {
	job, err := s.scheduler.NewJob(
		gocron.DurationJob(d),
		gocron.NewTask(fn, args...),
	)
	if err != nil {
		s.logger.Errorf("Error creating duration job: %v", err)
		return nil, err
	}
	return job, nil
}
