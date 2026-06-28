# Graceful Rules

**Rules for graceful shutdown and goroutine management via `govern/graceful` (`github.com/haipham22/govern/graceful`).**

---

## Overview

The `graceful` package coordinates goroutines, cleanup hooks, and signal-aware shutdown. Four components:

| Component | Role |
|-----------|------|
| **`Service`** | Interface: `Start(ctx) error` + `Shutdown(ctx) error` — the lifecycle contract |
| **`Run()`** | Top-level helper — starts `...Service`, waits, shuts down (the common case) |
| **`Manager`** | Low-level coordinator when you need manual goroutine/cleanup control |
| **`WorkerGroup`** | Bounded concurrency pool with graceful drain |

Source: [`graceful/`](../../graceful/) (`runner.go`, `service.go`, `manager.go`, `worker-group.go`).

**Core rules:**
- ✅ Prefer `Run(ctx, log, timeout, services...)` for the typical app
- ✅ Every managed goroutine MUST respect `ctx.Done()`
- ✅ Implement `Service` for anything with a lifecycle (HTTP server, worker, consumer)
- ✅ `Shutdown` must be idempotent and fast
- ✅ Register cleanup via `Defer` — runs LIFO (last registered, first run)
- ❌ Never add your own `signal.Notify` — the Manager already traps SIGINT/SIGTERM
- ❌ Never block in `Start`/`Shutdown` without checking `ctx.Done()`
- ❌ Never call `os.Exit` directly — let `Run`/`Manager` or `ExitOnSignal` handle it

---

## The `Service` Interface

**Everything with a lifecycle implements `Service`:**

```go
type Service interface {
    Start(ctx context.Context) error    // blocks until ctx done or fatal error
    Shutdown(ctx context.Context) error // idempotent, respects ctx deadline
}
```

**Implement it directly**, or adapt functions with `FromFunc`:

```go
// Struct implementation
type workerService struct{ wg *sync.WaitGroup }
func (w *workerService) Start(ctx context.Context) error    { /* loop on ctx.Done() */ }
func (w *workerService) Shutdown(ctx context.Context) error { /* stop + drain */ }

// Or adapt two functions
svc := graceful.FromFunc(
    func(ctx context.Context) error { return runConsumer(ctx) },
    func(ctx context.Context) error { return closeConsumer(ctx) },
)
```

**Rules:**
- ✅ `Start` blocks (runs the work loop); returns `nil` on clean ctx cancel, non-nil on fatal error
- ✅ `Shutdown` is idempotent — may be called multiple times
- ✅ `Shutdown` respects `ctx` deadline — don't exceed it
- ✅ `govern/http.Server` already satisfies `Service` (pass it straight to `Run`)
- ❌ Never return a non-nil error from `Start` on normal `ctx.Done()` (it would trigger fail-fast)
- ❌ Never do work in `Shutdown` that ignores the passed `ctx`

---

## `Run()` — the Common Case

**`Run` is the high-level entry: create Manager, start each Service, wait for signal/error, shut down. Use it from the composition root.**

```go
// cmd/serverd.go — RunE hands the server (a Service) to Run
return govern.Run(ctx, log, 10*time.Second, httpServer)
```

Internally `Run` does:
1. `NewManager(ctx)` — signal-aware context
2. `m.Defer(svc.Shutdown)` for each service → shutdown hooks registered
3. `m.Go(svc.Start)` for each service → started concurrently
4. `m.Wait()` — block on signal or first error
5. `m.Shutdown(timeout)` — LIFO cleanup with deadline

**Rules:**
- ✅ Pass every long-lived component as a `Service` to `Run`
- ✅ One `Run` call per process (at the composition root)
- ✅ Pass a realistic `shutdownTimeout` (enough to drain in-flight work)
- ❌ Don't wrap `Run` in your own Manager/Wait/Shutdown — it owns the full lifecycle
- 🔗 See [cobra-cli.md](cobra-cli.md) → *Composition Root in RunE* and [dependency-injection.md](dependency-injection.md)

---

## `Manager` — Manual Control

**Reach for `Manager` only when `Run` doesn't fit** (ad-hoc goroutines, dynamic cleanup, custom wait logic):

```go
m := graceful.NewManager(parentCtx, graceful.WithLogger(log))

// Managed goroutine — error triggers shutdown (fail-fast default)
m.Go(func(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return nil
        default:
            doWork(ctx)
        }
    }
})

// Cleanup hook — runs LIFO during Shutdown
m.Defer(func(ctx context.Context) error {
    return resource.Close()
})

if err := m.Wait(); err != nil { return err }
return m.Shutdown(10 * time.Second)
```

**API:**

| Method | Description |
|--------|-------------|
| `NewManager(parent, opts...)` | Signal-aware ctx (SIGINT/SIGTERM); `parent=nil` → Background |
| `Context()` | Root ctx, canceled on shutdown |
| `Go(fn)` | Managed goroutine; fn must respect `ctx.Done()` |
| `Defer(c)` | Register cleanup hook (LIFO) |
| `Wait()` | Block until signal/manual cancel or first goroutine error |
| `Shutdown(timeout)` | Cancel → wait goroutines → run cleanups (deadline ctx) |
| `InitiateShutdown()` | Idempotent manual shutdown trigger (`sync.Once`) |

**Options:** `WithFailFast(v)` (default **true**), `WithLogger(logger)`.

---

## Goroutines MUST Respect `ctx.Done()`

**A managed goroutine that ignores `ctx.Done()` blocks shutdown until the timeout.** Always select on the context:

```go
// GOOD — exits when ctx canceled
m.Go(func(ctx context.Context) error {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return nil
        case <-ticker.C:
            doWork(ctx)
        }
    }
})

// BAD — never exits; Shutdown times out and leaks the goroutine
m.Go(func(ctx context.Context) error {
    for { doWork() }   // ignores ctx
})
```

**Rules:**
- ✅ Use `ctx` from `Manager.Context()` (or the arg passed to `Start`/`Go`)
- ✅ Check `ctx.Done()` in every loop and before blocking ops
- ✅ Propagate `ctx` into HTTP clients, DB queries, downstream calls
- ❌ Never create a detached `context.Background()` for managed work
- 🔗 See [golang-context-concurrency.md](golang-context-concurrency.md)

---

## Cleanup Hooks (LIFO)

**`Defer` hooks run in reverse registration order — last registered, first run** (matches "last opened, first closed"):

```go
m := graceful.NewManager(ctx)
m.Defer(closeDB)      // registered 1st → runs 3rd
m.Defer(closeCache)   // registered 2nd → runs 2nd
m.Defer(closeLog)     // registered 3rd → runs 1st
```

**Rules:**
- ✅ Register cleanups in construction order — LIFO gives correct teardown
- ✅ Each cleanup gets a deadline ctx; respect it
- ✅ First cleanup error is returned (others still run)
- ❌ Never register cleanup after `Shutdown` has started
- 🔗 Mirrors the composition-root cleanup order in [dependency-injection.md](dependency-injection.md)

---

## Fail-Fast & Error Handling

**By default (`WithFailFast(true)`), the first non-`context.Canceled` goroutine error triggers shutdown.**

- `context.Canceled` is **ignored** (normal shutdown, not an error)
- The first real error is captured once (`sync.Once`) and returned by `Wait`
- Disable for self-healing workers: `NewManager(ctx, graceful.WithFailFast(false))`

**Rules:**
- ✅ Return `nil` from `Start`/`Go` on clean `ctx.Done()` — don't treat cancellation as failure
- ✅ Return a wrapped error only on genuine failure (so fail-fast engages)
- ✅ Use `WithFailFast(false)` only if a failed goroutine shouldn't take down the app

---

## `WorkerGroup` — Bounded Concurrency

**Pool with a max-concurrency semaphore; stops intake on shutdown; drains in-flight jobs.**

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

// Always Drain in a cleanup hook so inflight jobs finish before exit
m.Defer(func(ctx context.Context) error {
    return wg.Drain(5 * time.Second) // DeadlineExceeded if not done in time
})
```

| Method | Description |
|--------|-------------|
| `NewWorkerGroup(concurrency)` | `concurrency <= 0` → defaults to 1 |
| `TryGo(ctx, fn)` | Start job; returns **false** if shutdown started (stop intake) |
| `Drain(timeout)` | Wait for in-flight jobs; `context.DeadlineExceeded` on timeout |

**Rules:**
- ✅ Check `TryGo`'s return — `false` means stop submitting
- ✅ Always `Drain` in a `Defer` hook (or the Service's `Shutdown`)
- ❌ Never submit jobs after ctx cancel (TryGo handles this — respect its return)

---

## Signal Handling

**The Manager traps `SIGINT` + `SIGTERM` via `signal.NotifyContext`. Don't duplicate this.**

```go
// GOOD — Manager owns signal handling; RunE just builds the ctx it passes in
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()
return govern.Run(ctx, log, 10*time.Second, server)
```

**Rules:**
- ✅ Let the Manager handle SIGINT/SIGTERM
- ✅ For programmatic shutdown, call `m.InitiateShutdown()` (idempotent)
- ❌ Never add a second `signal.Notify` that races with the Manager

---

## Shutdown Timeout

**`Shutdown(timeout)` bounds both goroutine wait and cleanup.** Pick a value large enough to drain real work, small enough to not hang deployments:

- Goroutines get up to `timeout` to return from `ctx.Done()`
- Cleanup hooks share a deadline context of `timeout`
- On timeout: `Wait` returns `context.DeadlineExceeded`; cleanups still run

**Rules:**
- ✅ Match timeout to your slowest graceful drain (DB pool close, in-flight requests)
- ✅ Keep `Shutdown` work under the timeout
- ❌ Never set `timeout=0` in production (it means *wait forever* for goroutines)

---

## Best Practices & Pitfalls

**✅ DO:**
- Model every long-lived component as a `Service`
- Respect `ctx.Done()` everywhere — the whole package depends on it
- Register cleanup in construction order (LIFO teardown)
- Drain worker pools in a `Defer`/`Shutdown`

**❌ DON'T:**
- Ignore `ctx.Done()` in a managed goroutine (shutdown timeout, goroutine leak)
- Return non-nil on normal cancellation (triggers spurious fail-fast)
- Call `os.Exit` inside services (skips cleanups)
- Register cleanup after shutdown starts
- Re-implement signal handling the Manager already does

**Pitfalls:**
```go
// BAD — ignores ctx; Shutdown hangs until timeout
m.Go(func(ctx context.Context) error { for { work() } })

// BAD — treats cancellation as error → fail-fast fires on every normal shutdown
func (s *svc) Start(ctx context.Context) error {
    <-ctx.Done()
    return ctx.Err()                       // return nil instead
}

// BAD — WorkerGroup never drained → inflight jobs abandoned
wg := graceful.NewWorkerGroup(10)
// missing: m.Defer(func(ctx) error { return wg.Drain(5*time.Second) })

// BAD — races the Manager's signal handling
signal.Notify(sigCh, syscall.SIGTERM)      // Manager already does this
```

---

## Quick Reference

```go
// Common case — composition root
return graceful.Run(ctx, log, 10*time.Second, httpServer, workerSvc)

// Service via functions
svc := graceful.FromFunc(startFn, shutdownFn)

// Manual Manager
m := graceful.NewManager(ctx, graceful.WithLogger(log), graceful.WithFailFast(false))
m.Go(func(ctx context.Context) error { /* select on ctx.Done() */ return nil })
m.Defer(func(ctx context.Context) error { return resource.Close() })
err := m.Wait(); if err != nil { return err }
return m.Shutdown(10 * time.Second)

// Bounded workers
wg := graceful.NewWorkerGroup(10)
wg.TryGo(ctx, func(ctx context.Context) { work(ctx) })
m.Defer(func(ctx context.Context) error { return wg.Drain(5 * time.Second) })
```

| Concern | Rule |
|---------|------|
| Typical app | `Run(ctx, log, timeout, services...)` |
| Lifecycle unit | implement `Service` (Start + Shutdown) |
| Goroutines | always select on `ctx.Done()` |
| Cleanups | `Defer`, run LIFO |
| Fail-fast | default on; `context.Canceled` ignored |
| Signals | Manager traps SIGINT/SIGTERM — don't duplicate |
| Bounded work | `WorkerGroup` + `Drain` in cleanup |

---

## References

- Package source: [`graceful/`](../../graceful/) — `runner.go`, `service.go`, `manager.go`, `worker-group.go`
- [graceful/README.md](../../graceful/README.md) — full API table
- [cobra-cli.md](cobra-cli.md) → *Composition Root in RunE* (`govern.Run` call site)
- [dependency-injection.md](dependency-injection.md) → composition root & cleanup
- [golang-context-concurrency.md](golang-context-concurrency.md) → context & goroutine patterns
