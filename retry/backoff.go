package retry

import (
	"math"
	"math/rand"
	"time"
)

// BackoffStrategy calculates delay between retry attempts.
type BackoffStrategy interface {
	Delay(attempt int) time.Duration
}

// exponentialBackoff implements exponential backoff with jitter.
type exponentialBackoff struct {
	baseDelay   time.Duration
	maxDelay    time.Duration
	multiplier  float64
	jitter      bool
	jitterRatio float64
}

// ExponentialOption configures exponential backoff.
type ExponentialOption func(*exponentialBackoff)

// BaseDelay sets the initial delay.
func BaseDelay(d time.Duration) ExponentialOption {
	return func(e *exponentialBackoff) {
		e.baseDelay = d
	}
}

// MaxDelay sets the maximum delay.
func MaxDelay(d time.Duration) ExponentialOption {
	return func(e *exponentialBackoff) {
		e.maxDelay = d
	}
}

// Multiplier sets the backoff multiplier.
func Multiplier(m float64) ExponentialOption {
	return func(e *exponentialBackoff) {
		e.multiplier = m
	}
}

// WithJitter enables jitter to avoid thundering herd.
func WithJitter() ExponentialOption {
	return func(e *exponentialBackoff) {
		e.jitter = true
	}
}

// JitterRatio sets the jitter ratio (default 0.25).
func JitterRatio(r float64) ExponentialOption {
	return func(e *exponentialBackoff) {
		e.jitterRatio = r
		e.jitter = true
	}
}

// NewExponentialBackoff creates an exponential backoff strategy.
func NewExponentialBackoff(opts ...ExponentialOption) *exponentialBackoff {
	e := &exponentialBackoff{
		baseDelay:   100 * time.Millisecond,
		maxDelay:    30 * time.Second,
		multiplier:  2,
		jitter:      true,
		jitterRatio: 0.25,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Delay calculates the delay for the given attempt.
func (e *exponentialBackoff) Delay(attempt int) time.Duration {
	delay := float64(e.baseDelay) * math.Pow(e.multiplier, float64(attempt))
	if delay > float64(e.maxDelay) {
		delay = float64(e.maxDelay)
	}

	if e.jitter {
		jitterRange := delay * e.jitterRatio
		delay -= jitterRange
		delay += 2 * jitterRange * rand.Float64()
	}

	return time.Duration(delay)
}

// linearBackoff implements linear backoff.
type linearBackoff struct {
	baseDelay time.Duration
	increment time.Duration
	maxDelay  time.Duration
}

// LinearOption configures linear backoff.
type LinearOption func(*linearBackoff)

// LinearBaseDelay sets the initial delay.
func LinearBaseDelay(d time.Duration) LinearOption {
	return func(l *linearBackoff) {
		l.baseDelay = d
	}
}

// LinearIncrement sets the delay increment per attempt.
func LinearIncrement(d time.Duration) LinearOption {
	return func(l *linearBackoff) {
		l.increment = d
	}
}

// LinearMaxDelay sets the maximum delay.
func LinearMaxDelay(d time.Duration) LinearOption {
	return func(l *linearBackoff) {
		l.maxDelay = d
	}
}

// NewLinearBackoff creates a linear backoff strategy.
func NewLinearBackoff(opts ...LinearOption) *linearBackoff {
	l := &linearBackoff{
		baseDelay: 100 * time.Millisecond,
		increment: 100 * time.Millisecond,
		maxDelay:  30 * time.Second,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// Delay calculates the delay for the given attempt.
func (l *linearBackoff) Delay(attempt int) time.Duration {
	delay := l.baseDelay + time.Duration(attempt)*l.increment
	if delay > l.maxDelay {
		delay = l.maxDelay
	}
	return delay
}

// constantBackoff implements constant delay between retries.
type constantBackoff struct {
	delay time.Duration
}

// NewConstantBackoff creates a constant backoff strategy.
func NewConstantBackoff(d time.Duration) *constantBackoff {
	return &constantBackoff{delay: d}
}

// Delay returns the constant delay.
func (c *constantBackoff) Delay(_ int) time.Duration {
	return c.delay
}
