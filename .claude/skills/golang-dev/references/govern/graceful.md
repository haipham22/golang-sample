# govern/graceful

Import: `github.com/haipham22/govern/graceful`

Graceful shutdown + goroutine/cleanup coordination. Signal-aware (SIGINT/SIGTERM).

## Components

| Component | Role |
|---|---|
| `Service` | Interface: `Start(ctx) error` + `Shutdown(ctx) error`. Lifecycle contract. |
| `Run()` | Top helper: start services, wait signal/error, shut down. Common case. |
| `Manager` | Low-level coordinator for manual goroutine/cleanup control. |
| `WorkerGroup` | Bounded concurrency pool with graceful drain. |

## Use When

- App needs signal-aware shutdown.
- Long-lived components (HTTP server, worker, consumer) need lifecycle.
- Coordinated goroutines with cleanup hooks.

## Run — Common Case

```go
import (
    "github.com/haipham22/govern/graceful"
)

return graceful.Run(ctx, logger, 10*time.Second, httpServer, workerSvc)
```

Internally: `NewManager` → `Defer(svc.Shutdown)` per service → `Go(svc.Start)` per service → `Wait` → `Shutdown(timeout)` (LIFO).

## Manager — Manual Control

```go
m := graceful.NewManager(nil, graceful.WithLogger(logger))

// Managed goroutine; must respect ctx.Done()
m.Go(func(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return nil
        case <-ticker.C:
            doWork(ctx)
        }
    }
})

// Cleanup hook (LIFO)
m.Defer(func(ctx context.Context) error {
    return resource.Close()
})

if err := m.Wait(); err != nil { return err }
return m.Shutdown(10 * time.Second)
```

### Manager API

| Method | Description |
|---|---|
| `NewManager(parent, opts...)` | Signal-aware ctx; `parent=nil` → Background |
| `Context()` | Root ctx, canceled on shutdown |
| `Go(fn)` | Managed goroutine; fn must respect `ctx.Done()` |
| `Defer(c)` | Register cleanup hook (LIFO) |
| `Wait()` | Block until signal/cancel or first goroutine error |
| `Shutdown(timeout)` | Cancel → wait goroutines → run cleanups (deadline ctx) |
| `InitiateShutdown()` | Idempotent manual shutdown trigger |

Options: `WithFailFast(v)` (default **true**), `WithLogger(logger)`.

## Service Interface

```go
type Service interface {
    Start(ctx context.Context) error
    Shutdown(ctx context.Context) error
}
```

Or adapt two functions:

```go
svc := graceful.FromFunc(startFn, shutdownFn)
```

## WorkerGroup — Bounded Concurrency

```go
m := graceful.NewManager(ctx)
wg := graceful.NewWorkerGroup(10) // max 10 concurrent

m.Go(func(ctx context.Context) error {
    for _, job := range jobs {
        if !wg.TryGo(ctx, func(ctx context.Context) { process(ctx, job) }) {
            break // shutdown started — stop intake
        }
    }
    return nil
})

m.Defer(func(ctx context.Context) error {
    return wg.Drain(5 * time.Second)
})
```

| Method | Description |
|---|---|
| `NewWorkerGroup(concurrency)` | `concurrency <= 0` → defaults to 1 |
| `TryGo(ctx, fn)` | Start job; **false** = shutdown started, stop intake |
| `Drain(timeout)` | Wait in-flight jobs; `DeadlineExceeded` on timeout |

## Rules

- ✅ Every managed goroutine MUST select on `ctx.Done()`.
- ✅ `Start` returns `nil` on clean `ctx.Done()`; non-nil only on fatal error.
- ✅ `Shutdown` idempotent + respects ctx deadline.
- ✅ Register cleanups in construction order → LIFO teardown.
- ✅ Always `Drain` worker pools in `Defer`/`Shutdown`.
- ✅ Return `nil` (not `ctx.Err()`) on normal cancellation — avoid spurious fail-fast.
- ❌ Never ignore `ctx.Done()` (shutdown timeout, goroutine leak).
- ❌ Never add second `signal.Notify` racing the Manager.
- ❌ Never `os.Exit` inside services (skips cleanups).
- ❌ Never register cleanup after shutdown starts.

## Avoid

- Hand-rolled `signal.Notify` + `sync.WaitGroup`.
- Unmanaged goroutines that ignore `ctx.Done()`.
- Treating cancellation as error → spurious fail-fast.

## Timeout Guidance

Match `Shutdown(timeout)` to slowest graceful drain (DB pool close, in-flight requests). `timeout=0` means wait forever for goroutines — do not use in production.

## Reference

Source: [`graceful/`](../../../../../../../graceful/) — `runner.go`, `service.go`, `manager.go`, `worker-group.go`.
