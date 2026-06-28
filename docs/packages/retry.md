# retry

Exponential backoff retry with configurable policies.

## Overview

The `retry` package provides retry logic with exponential backoff, configurable policies, and support for context cancellation.

## Key Types

### Policy

```go
type Policy struct {
    maxAttempts int
    maxDuration time.Duration
    backoff     BackoffStrategy
}
```

### BackoffStrategy

```go
type BackoffStrategy interface {
    Delay(attempt int) time.Duration
}
```

### Function Types

```go
type Func func() error
type FuncWithContext func(ctx context.Context) error
```

## Key Functions

### Policy Creation

```go
// Create new policy with defaults
func NewPolicy(opts ...Option) *Policy

// Execute function with retry
func (p *Policy) Do(fn Func) error

// Execute function with retry and context
func (p *Policy) DoWithContext(ctx context.Context, fn FuncWithContext) error
```

### Convenience Functions

```go
// Retry with default policy
func Do(fn Func, opts ...Option) error

// Retry with context using default policy
func DoWithContext(ctx context.Context, fn FuncWithContext, opts ...Option) error
```

### Options

```go
// Set maximum retry attempts
func MaxAttempts(n int) Option

// Set maximum total duration
func MaxDuration(d time.Duration) Option

// Set backoff strategy
func Backoff(strategy BackoffStrategy) Option
```

### Backoff Strategies

```go
// Create exponential backoff with jitter
func NewExponentialBackoff() BackoffStrategy

// Create fixed delay backoff
func NewFixedBackoff(delay time.Duration) BackoffStrategy

// Create linear backoff
func NewLinearBackoff(delay time.Duration, increment time.Duration) BackoffStrategy
```

### Predicate Functions

```go
// Mark error as non-retryable
func MarkNonRetryable(err error)

// Check if error is non-retryable
func IsNonRetryable(err error) bool

// Wrap error to make it non-retryable
func NonRetryable(err error) error
```

## Defaults

```go
Default MaxAttempts: 3
Default MaxDuration: 1 minute
Default Backoff: Exponential with jitter
```

## Usage

### Basic Retry

```go
import "github.com/haipham22/govern/retry"

// Simple retry with default policy
err := retry.Do(func() error {
    return callAPI()
})

if err != nil {
    log.Errorf("All retry attempts failed: %v", err)
}
```

### With Custom Options

```go
// Custom retry policy
err := retry.Do(func() error {
    return callAPI()
},
    retry.MaxAttempts(5),
    retry.MaxDuration(2*time.Minute),
)

if err != nil {
    log.Errorf("Failed after 5 attempts: %v", err)
}
```

### With Context

```go
// Retry with context cancellation
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := retry.DoWithContext(ctx, func(ctx context.Context) error {
    return callAPIWithContext(ctx)
})

if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        log.Error("Operation timed out")
    } else {
        log.Errorf("Operation failed: %v", err)
    }
}
```

### With Custom Backoff

```go
// Exponential backoff with jitter
policy := retry.NewPolicy(
    retry.Backoff(retry.NewExponentialBackoff()),
)

// Fixed delay backoff
policy := retry.NewPolicy(
    retry.Backoff(retry.NewFixedBackoff(time.Second)),
)

// Linear backoff
policy := retry.NewPolicy(
    retry.Backoff(retry.NewLinearBackoff(time.Second, 500*time.Millisecond)),
)

err := policy.Do(func() error {
    return callAPI()
})
```

### Non-Retryable Errors

```go
import "github.com/haipham22/govern/retry"

err := retry.Do(func() error {
    resp, err := http.Get("https://api.example.com")
    if err != nil {
        return err // Retryable
    }
    defer resp.Body.Close()

    if resp.StatusCode == 404 {
        // Don't retry on 404
        return retry.NonRetryable(fmt.Errorf("resource not found"))
    }

    if resp.StatusCode >= 500 {
        return fmt.Errorf("server error") // Retryable
    }

    return nil
})
```

### Marking Errors Non-Retryable

```go
var ErrNotFound = errors.New("not found")

err := retry.Do(func() error {
    resource, err := fetchResource(id)
    if errors.Is(err, ErrNotFound) {
        // Mark this error as non-retryable
        retry.MarkNonRetryable(err)
        return err
    }
    return err
})
```

### Custom Policy

```go
// Create custom policy
policy := retry.NewPolicy(
    retry.MaxAttempts(10),
    retry.MaxDuration(5*time.Minute),
    retry.Backoff(retry.NewExponentialBackoff()),
)

// Use custom policy
err := policy.Do(func() error {
    return processTask()
})
```

### Retry in Loop

```go
// Process multiple items with retry
items := []Item{item1, item2, item3}

for _, item := range items {
    err := retry.Do(func() error {
        return processItem(item)
    })

    if err != nil {
        log.Errorf("Failed to process item %d: %v", item.ID, err)
        // Continue with next item
    }
}
```

### Database Operations

```go
func (r *repository) CreateUser(ctx context.Context, user *User) error {
    return retry.DoWithContext(ctx, func(ctx context.Context) error {
        return r.db.WithContext(ctx).Create(user).Error
    })
}
```

### HTTP Client

```go
func (c *Client) FetchData(ctx context.Context, url string) ([]byte, error) {
    var data []byte
    
    err := retry.DoWithContext(ctx, func(ctx context.Context) error {
        req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
        if err != nil {
            return err
        }

        resp, err := c.httpClient.Do(req)
        if err != nil {
            return err
        }
        defer resp.Body.Close()

        if resp.StatusCode == http.StatusTooManyRequests {
            return fmt.Errorf("rate limited") // Retry
        }

        if resp.StatusCode >= 500 {
            return fmt.Errorf("server error") // Retry
        }

        if resp.StatusCode >= 400 {
            return retry.NonRetryable(fmt.Errorf("client error")) // Don't retry
        }

        data, err = io.ReadAll(resp.Body)
        return err
    })

    return data, err
}
```

## Backoff Strategies

### Exponential Backoff

```go
// Delay formula: base_delay * 2^attempt + jitter
// Attempt 0: ~100ms
// Attempt 1: ~200ms
// Attempt 2: ~400ms
// Attempt 3: ~800ms

strategy := retry.NewExponentialBackoff()
```

### Fixed Backoff

```go
// Constant delay between attempts
strategy := retry.NewFixedBackoff(time.Second)
// All attempts: 1 second delay
```

### Linear Backoff

```go
// Delay formula: base_delay + (attempt * increment)
// Attempt 0: 1s
// Attempt 1: 1.5s
// Attempt 2: 2s
// Attempt 3: 2.5s

strategy := retry.NewLinearBackoff(time.Second, 500*time.Millisecond)
```

### Custom Backoff

```go
type customBackoff struct{}

func (c *customBackoff) Delay(attempt int) time.Duration {
    // Custom delay calculation
    return time.Duration(attempt*attempt) * time.Second
}

policy := retry.NewPolicy(
    retry.Backoff(&customBackoff{}),
)
```

## Best Practices

1. **Set reasonable limits** - Use MaxAttempts and MaxDuration to prevent infinite retries
2. **Context cancellation** - Use DoWithContext for timeout support
3. **Non-retryable errors** - Mark errors that shouldn't be retried (404, validation errors)
4. **Backoff strategy** - Use exponential backoff for most cases
5. **Logging** - Log retry attempts for debugging
6. **Side effects** - Ensure operations are idempotent when retrying

## Common Patterns

### API Calls

```go
func callAPIWithRetry(url string) error {
    return retry.Do(func() error {
        resp, err := http.Get(url)
        if err != nil {
            return err // Retry network errors
        }
        defer resp.Body.Close()

        if resp.StatusCode >= 500 {
            return fmt.Errorf("server error: %d", resp.StatusCode) // Retry
        }

        if resp.StatusCode >= 400 {
            return retry.NonRetryable(fmt.Errorf("client error: %d", resp.StatusCode))
        }

        return nil
    })
}
```

### Database Queries

```go
func queryWithRetry(db *gorm.DB, user *User) error {
    return retry.Do(func() error {
        return db.First(user).Error
    })
}
```

### External Services

```go
func sendEmailWithRetry(to, subject, body string) error {
    return retry.Do(func() error {
        return emailService.Send(to, subject, body)
    })
}
```

## References

- [retry/policy.go](../../retry/policy.go) - Policy implementation
- [retry/backoff.go](../../retry/backoff.go) - Backoff strategies
- [retry/predicate.go](../../retry/predicate.go) - Error predicates
