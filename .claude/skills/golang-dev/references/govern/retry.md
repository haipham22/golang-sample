# govern/retry

Import: `github.com/haipham22/govern/retry`

Flexible retry with pluggable backoff (exponential, linear, constant), jitter, context, retry predicates.

## Use When

- App retries flaky external calls (HTTP, DB connection, queue).

## Basic

```go
import "github.com/haipham22/govern/retry"

err := retry.Do(func() error { return callAPI() })
```

## With Options

```go
err := retry.Do(func() error { return callAPI() },
    retry.MaxAttempts(5),
    retry.MaxDuration(time.Minute),
    retry.Backoff(retry.NewExponentialBackoff()),
)
```

## With Context

```go
err := retry.DoWithContext(ctx, func(ctx context.Context) error {
    return callAPIWithContext(ctx)
})
```

## Backoff Strategies

```go
retry.Backoff(retry.NewExponentialBackoff()) // default, with jitter

retry.Backoff(retry.NewLinearBackoff(
    retry.LinearBaseDelay(100*time.Millisecond),
    retry.LinearIncrement(50*time.Millisecond),
))

retry.Backoff(retry.NewConstantBackoff(100*time.Millisecond))
```

## Conditional Retry

```go
isRetryable := func(err error) bool { return IsTemporary(err) }
fn := retry.RetryIf(func(ctx context.Context) error {
    return riskyOperation()
}, isRetryable)
retry.Do(fn)
```

## Rules

- ✅ Retry only transient errors (timeouts, connection, 5xx, deadlock).
- ✅ Use jitter (default in exponential) to avoid thundering herd.
- ✅ Respect context cancellation.
- ✅ Cap attempts + total duration.
- ❌ Never retry permanent errors (validation, auth, `CodeInvalid`, `CodeNotFound`).
- ❌ Do not retry inside a DB transaction.

## Avoid

- Hand-rolled retry loops (no jitter, no context).
- `cenkalti/backoff` when govern/retry covers same ground.

## Reference

Source: [`retry/`](../../../../../../../retry/).
