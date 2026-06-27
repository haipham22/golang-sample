package retry

import (
	"context"
	"fmt"
	"time"
)

// Func is a function that can be retried.
type Func func() error

// FuncWithContext is a function that accepts context and can be retried.
type FuncWithContext func(ctx context.Context) error

// Policy defines retry behavior.
type Policy struct {
	maxAttempts int
	maxDuration time.Duration
	backoff     BackoffStrategy
}

// Option configures a retry Policy.
type Option func(*Policy)

// MaxAttempts sets the maximum number of retry attempts.
func MaxAttempts(n int) Option {
	return func(p *Policy) {
		p.maxAttempts = n
	}
}

// MaxDuration sets the maximum total duration for retries.
func MaxDuration(d time.Duration) Option {
	return func(p *Policy) {
		p.maxDuration = d
	}
}

// Backoff sets the backoff strategy.
func Backoff(strategy BackoffStrategy) Option {
	return func(p *Policy) {
		p.backoff = strategy
	}
}

// NewPolicy creates a new retry policy with defaults:
// - maxAttempts: 3
// - maxDuration: 1 minute
// - backoff: exponential with jitter
func NewPolicy(opts ...Option) *Policy {
	p := &Policy{
		maxAttempts: 3,
		maxDuration: time.Minute,
		backoff:     NewExponentialBackoff(),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Do executes fn with retry according to the policy.
func (p *Policy) Do(fn Func) error {
	ctx := context.Background()
	return p.DoWithContext(ctx, func(ctx context.Context) error {
		return fn()
	})
}

// DoWithContext executes fn with retry according to the policy.
func (p *Policy) DoWithContext(ctx context.Context, fn FuncWithContext) error {
	deadline := time.Now().Add(p.maxDuration)
	var lastErr error

	for attempt := 0; attempt < p.maxAttempts; attempt++ {
		if ctx.Err() != nil {
			return fmt.Errorf("retry canceled: %w", ctx.Err())
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("retry deadline exceeded: %w", lastErr)
		}

		lastErr = fn(ctx)
		if lastErr == nil {
			return nil
		}

		// Stop retrying if the error is marked as non-retryable
		if IsNonRetryable(lastErr) {
			return lastErr
		}

		if attempt < p.maxAttempts-1 {
			delay := p.backoff.Delay(attempt)
			if time.Now().Add(delay).After(deadline) {
				delay = time.Until(deadline)
			}
			if delay > 0 {
				select {
				case <-time.After(delay):
				case <-ctx.Done():
					return fmt.Errorf("retry canceled during backoff: %w", ctx.Err())
				}
			}
		}
	}

	return fmt.Errorf("max retries exceeded (%d attempts): %w", p.maxAttempts, lastErr)
}

// Do is a convenience function using the default policy.
func Do(fn Func, opts ...Option) error {
	return NewPolicy(opts...).Do(fn)
}

// DoWithContext is a convenience function using the default policy.
func DoWithContext(ctx context.Context, fn FuncWithContext, opts ...Option) error {
	return NewPolicy(opts...).DoWithContext(ctx, fn)
}
