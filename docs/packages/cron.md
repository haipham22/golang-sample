# cron

Cron scheduler with graceful shutdown and job lifecycle management.

## Overview

The `cron` package provides a scheduled job manager built on `go-co-op/gocron/v2` with Govern patterns for graceful shutdown and lifecycle management.

## Key Types

### Scheduler

```go
type Scheduler struct {
    scheduler   gocron.Scheduler
    logger      *zap.SugaredLogger
    location    *time.Location
    stopTimeout time.Duration
}
```

### JobHandler

```go
type JobHandler interface {
    Setup(session SchedulerSession) error
    Execute(session SchedulerSession) error
    Cleanup(session SchedulerSession) error
}
```

## Key Functions

### Creating Scheduler

```go
// Create new scheduler with options
func New(opts ...Option) (*Scheduler, func(), error)
```

### Job Methods

```go
// Create duration-based job
func (s *Scheduler) DurationJob(d time.Duration, fn any, args ...any) (gocron.Job, error)

// Create cron expression job
func (s *Scheduler) CronJob(cronExpr string, fn any, args ...any) (gocron.Job, error)

// Create job with custom handler
func (s *Scheduler) Job(handler JobHandler) (gocron.Job, error)
```

### Lifecycle

```go
// Start scheduler (non-blocking)
func (s *Scheduler) Start(ctx context.Context) error

// Shutdown gracefully
func (s *Scheduler) Shutdown(ctx context.Context) error
```

## Usage

### Simple Job Function

```go
scheduler, cleanup, err := cron.New(
    cron.WithLogger(logger),
    cron.WithLocation(time.UTC),
)
defer cleanup()

// Simple function job
err = scheduler.DurationJob(
    10*time.Minute,
    func() {
        log.Info("Running cleanup task")
        // Cleanup logic
    },
)

scheduler.Start(context.Background())
```

### JobHandler with Lifecycle

```go
type DataSyncJob struct {
    db *gorm.DB
}

func (j *DataSyncJob) Setup(session cron.SchedulerSession) error {
    log.Info("Setting up data sync job")
    // Initialize connections, validate state
    return nil
}

func (j *DataSyncJob) Execute(session cron.SchedulerSession) error {
    log.Info("Executing data sync")
    // Main sync logic
    return j.syncData(session.Context())
}

func (j *DataSyncJob) Cleanup(session cron.SchedulerSession) error {
    log.Info("Cleaning up data sync job")
    // Release resources, close connections
    return nil
}

// Register job
job := &DataSyncJob{db: database}
scheduler.Job(job)
```

### Cron Expression

```go
// Run daily at 2 AM
scheduler.CronJob("0 2 * * *", func() {
    log.Info("Running daily backup")
    backup()
})

// Run every hour
scheduler.CronJob("0 * * * *", func() {
    processHourlyMetrics()
})
```

## JobHandler Function Adapter

For simple jobs that don't need resource management:

```go
handler := cron.JobHandlerFunc(func(ctx context.Context) error {
    log.Info("Running simple job")
    return doTask(ctx)
})

scheduler.DurationJob(5*time.Minute, handler)
```

## Options

```go
// Set custom logger
func WithLogger(logger *zap.SugaredLogger) Option

// Set timezone location
func WithLocation(loc *time.Location) Option

// Set shutdown timeout
func WithStopTimeout(timeout time.Duration) Option
```

## Integration with Graceful Shutdown

```go
import "github.com/haipham22/govern/graceful"

scheduler, cleanup, _ := cron.New()
defer cleanup()

// Use with graceful.Run
graceful.Run(ctx, logger, 30*time.Second, scheduler)
```

## Thread Safety

**IMPORTANT:** JobHandler instances may be called from multiple goroutines concurrently if the scheduler allows overlapping job executions. Ensure all state is safely protected against race conditions.

## References

- [cron/scheduler.go](../../cron/scheduler.go) - Scheduler implementation
- [cron/handler.go](../../cron/handler.go) - JobHandler interface
- [cron/options.go](../../cron/options.go) - Configuration options
