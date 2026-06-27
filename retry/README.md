# retry

Flexible retry logic with pluggable backoff strategies.

## Features

- Configurable max attempts and duration
- Exponential, linear, and constant backoff
- Jitter support to avoid thundering herd
- Context cancellation support
- Retry predicates for conditional retry

## Usage

```go
import "github.com/haipham22/govern/retry"

// Simple retry
err := retry.Do(func() error {
    return callAPI()
})

// With options
err := retry.Do(func() error {
    return callAPI()
}, retry.MaxAttempts(5), retry.MaxDuration(time.Minute))

// With context
err := retry.DoWithContext(ctx, func(ctx context.Context) error {
    return callAPIWithContext(ctx)
})
```

## Backoff Strategies

```go
// Exponential with jitter (default)
retry.Backoff(retry.NewExponentialBackoff())

// Linear backoff
retry.Backoff(retry.NewLinearBackoff(
    retry.LinearBaseDelay(100*time.Millisecond),
    retry.LinearIncrement(50*time.Millisecond),
))

// Constant backoff
retry.Backoff(retry.NewConstantBackoff(100*time.Millisecond))
```

## Conditional Retry

```go
// Only retry specific errors
isRetryable := func(err error) bool {
    return IsTemporaryError(err)
}

fn := retry.RetryIf(func(ctx context.Context) error {
    return riskyOperation()
}, isRetryable)

retry.Do(fn)
```
