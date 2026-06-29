# govern/cron

Import: `github.com/haipham22/govern/cron`

Cron scheduler on gocron v2 with graceful lifecycle. Implements `graceful.Service`.

## Use When

- App schedules background jobs (cleanup, batch, periodic sync).

## Quick Start

```go
import (
    "github.com/haipham22/govern/cron"
    "github.com/haipham22/govern/graceful"
    "github.com/haipham22/govern/log"
)

logger := log.New()
scheduler, cleanup, _ := cron.New(cron.WithLogger(logger))
defer cleanup()

_, _ = scheduler.DurationJob(5*time.Minute, func() {
    logger.Info("running cleanup")
})

graceful.Run(ctx, logger, 30*time.Second, scheduler)
```

## Options

| Option | Description | Default |
|---|---|---|
| `WithLogger(l)` | Zap logger | `log.Default()` |
| `WithLocation(loc)` | Timezone | `time.Local` |
| `WithStopTimeout(d)` | Shutdown timeout | 30s |

## Job Types

| Method | Description |
|---|---|
| `DurationJob(d, fn, args...)` | Fixed interval |
| `CronJob(expr, withSecs, fn)` | Cron expression |
| `DailyJob(n, atTimes, fn)` | Every N days at times |
| `WeeklyJob(...)` | Weekly on weekdays |
| `MonthlyJob(...)` | Monthly on days |
| `OneTimeJob(...)` | Run once at time |
| `RandomDurationJob(min, max, fn)` | Random interval |

Daily example:

```go
import gocronv2 "github.com/go-co-op/gocron/v2"

scheduler.DailyJob(1, gocronv2.NewAtTimes(gocronv2.NewAtTime(9, 0, 0)), myFunc)
```

## Rules

- ✅ Wire scheduler through `graceful.Run` so jobs drain on shutdown.
- ✅ Capture `(scheduler, cleanup, err)` from `cron.New`; defer cleanup.
- ✅ Keep job functions idempotent where possible.
- ✅ Log job start/success/failure with job name.
- ❌ Do not run external API calls inside DB transactions from jobs.
- ❌ Do not use raw gocron without graceful lifecycle.

## Avoid

- Raw `robfig/cron` or gocron wired without shutdown drain (jobs killed mid-run).
- Long-running jobs that exceed shutdown timeout without checkpointing.

## Reference

Source: [`cron/`](../../../../../../../cron/). Uses `github.com/go-co-op/gocron/v2`.
