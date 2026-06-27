package retry

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPolicy_Do(t *testing.T) {
	tests := []struct {
		name        string
		policy      *Policy
		fnFactory   func(*int) func() error
		wantErr     bool
		errMsg      string
		wantCalls   int
		description string
	}{
		{
			name:   "success after retries",
			policy: NewPolicy(),
			fnFactory: func(calls *int) func() error {
				return func() error {
					(*calls)++
					if *calls < 2 {
						return errors.New("temporary error")
					}
					return nil
				}
			},
			wantErr:     false,
			wantCalls:   2,
			description: "should succeed on second attempt",
		},
		{
			name:   "max attempts exceeded",
			policy: NewPolicy(MaxAttempts(3)),
			fnFactory: func(calls *int) func() error {
				return func() error {
					(*calls)++
					return errors.New("always fails")
				}
			},
			wantErr:     true,
			errMsg:      "max retries exceeded",
			wantCalls:   3,
			description: "should fail after max attempts",
		},
		{
			name:   "zero attempts",
			policy: NewPolicy(MaxAttempts(1)),
			fnFactory: func(calls *int) func() error {
				return func() error {
					(*calls)++
					return errors.New("always fails")
				}
			},
			wantErr:     true,
			wantCalls:   1,
			description: "should fail on single attempt",
		},
		{
			name:   "last error returned",
			policy: NewPolicy(),
			fnFactory: func(calls *int) func() error {
				return func() error {
					(*calls)++
					return errors.New("final error")
				}
			},
			wantErr:     true,
			errMsg:      "final error",
			wantCalls:   3,
			description: "should include last error in message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			fn := tt.fnFactory(&calls)

			err := tt.policy.Do(fn)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantCalls, calls)
		})
	}
}

func TestPolicy_Do_MaxDuration(t *testing.T) {
	tests := []struct {
		name        string
		policy      *Policy
		fnFactory   func(*int) func() error
		wantErr     bool
		errMsg      string
		maxCalls    int
		description string
	}{
		{
			name:   "deadline exceeded",
			policy: NewPolicy(MaxDuration(100 * time.Millisecond)),
			fnFactory: func(calls *int) func() error {
				return func() error {
					(*calls)++
					time.Sleep(10 * time.Millisecond)
					return errors.New("always fails")
				}
			},
			wantErr:     true,
			errMsg:      "deadline exceeded",
			maxCalls:    2,
			description: "should stop when max duration exceeded",
		},
		{
			name:   "max duration with fast fail",
			policy: NewPolicy(MaxDuration(50*time.Millisecond), Backoff(NewConstantBackoff(30*time.Millisecond))),
			fnFactory: func(calls *int) func() error {
				return func() error {
					(*calls)++
					time.Sleep(10 * time.Millisecond)
					return errors.New("always fails")
				}
			},
			wantErr:     true,
			maxCalls:    2,
			description: "should hit max duration before max attempts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			fn := tt.fnFactory(&calls)

			err := tt.policy.Do(fn)

			assert.Error(t, err)
			if tt.errMsg != "" {
				assert.Contains(t, err.Error(), tt.errMsg)
			}
			assert.LessOrEqual(t, calls, tt.maxCalls)
		})
	}
}

func TestPolicy_Do_BackingOff(t *testing.T) {
	tests := []struct {
		name          string
		policy        *Policy
		fnFactory     func(*int) func() error
		wantErr       bool
		wantCalls     int
		minElapsed    time.Duration
		maxElapsed    time.Duration
		expectedDelay time.Duration
		description   string
	}{
		{
			name:   "backoff option applied",
			policy: NewPolicy(MaxAttempts(5), MaxDuration(time.Hour), Backoff(NewConstantBackoff(10*time.Millisecond))),
			fnFactory: func(calls *int) func() error {
				return func() error {
					(*calls)++
					if *calls < 3 {
						return errors.New("temp")
					}
					return nil
				}
			},
			wantErr:    false,
			wantCalls:  3,
			minElapsed: 15 * time.Millisecond,
			maxElapsed: 100 * time.Millisecond,
		},
		{
			name:   "fast failure with backoff",
			policy: NewPolicy(MaxAttempts(5), MaxDuration(time.Hour), Backoff(NewConstantBackoff(10*time.Millisecond))),
			fnFactory: func(calls *int) func() error {
				return func() error {
					(*calls)++
					return errors.New("fail fast")
				}
			},
			wantErr:    true,
			wantCalls:  5,
			minElapsed: 40 * time.Millisecond,
		},
		{
			name:   "long running success",
			policy: NewPolicy(MaxAttempts(1)),
			fnFactory: func(calls *int) func() error {
				return func() error {
					time.Sleep(50 * time.Millisecond)
					return nil
				}
			},
			wantErr:    false,
			minElapsed: 50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			start := time.Now()
			fn := tt.fnFactory(&calls)

			err := tt.policy.Do(fn)
			elapsed := time.Since(start)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantCalls > 0 {
				assert.Equal(t, tt.wantCalls, calls)
			}

			if tt.minElapsed > 0 {
				assert.True(t, elapsed >= tt.minElapsed, "elapsed %v should be >= %v", elapsed, tt.minElapsed)
			}
			if tt.maxElapsed > 0 {
				assert.True(t, elapsed < tt.maxElapsed, "elapsed %v should be < %v", elapsed, tt.maxElapsed)
			}
		})
	}
}

func TestPolicy_DoWithContext(t *testing.T) {
	tests := []struct {
		name        string
		setupCtx    func() (context.Context, context.CancelFunc)
		testFunc    func(*testing.T, *Policy, context.Context, context.CancelFunc, *TestContextDoWithParams)
		wantErr     bool
		errMsg      string
		wantCalls   int
		maxCalls    int
		description string
	}{
		{
			name: "context canceled during retry",
			setupCtx: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			testFunc: func(t *testing.T, p *Policy, ctx context.Context, cancel context.CancelFunc, params *TestContextDoWithParams) {
				calls := int32(0)
				done := make(chan error)
				go func() {
					done <- p.DoWithContext(ctx, func(ctx context.Context) error {
						if atomic.AddInt32(&calls, 1) == 2 {
							cancel()
						}
						time.Sleep(10 * time.Millisecond)
						return errors.New("fails")
					})
				}()

				err := <-done
				assert.Error(t, err)
				assert.Contains(t, err.Error(), params.errMsg)
				assert.True(t, atomic.LoadInt32(&calls) <= 3)
			},
			wantErr:  true,
			errMsg:   "canceled",
			maxCalls: 3,
		},
		{
			name: "context already canceled",
			setupCtx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, cancel
			},
			testFunc: func(t *testing.T, p *Policy, ctx context.Context, cancel context.CancelFunc, params *TestContextDoWithParams) {
				err := p.DoWithContext(ctx, func(ctx context.Context) error {
					t.Fatal("function should not be called")
					return nil
				})
				assert.Error(t, err)
				assert.Contains(t, err.Error(), params.errMsg)
			},
			wantErr:     true,
			errMsg:      "canceled",
			wantCalls:   0,
			description: "should fail immediately if context canceled",
		},
		{
			name: "passing context to function",
			setupCtx: func() (context.Context, context.CancelFunc) {
				return context.Background(), func() {}
			},
			testFunc: func(t *testing.T, p *Policy, ctx context.Context, cancel context.CancelFunc, params *TestContextDoWithParams) {
				calls := 0
				err := p.DoWithContext(ctx, func(ctx context.Context) error {
					calls++
					assert.NotNil(t, ctx)
					if calls < 2 {
						return errors.New("temp")
					}
					return nil
				})
				assert.NoError(t, err)
				assert.Equal(t, params.wantCalls, calls)
			},
			wantErr:   false,
			wantCalls: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPolicy()
			ctx, cancel := tt.setupCtx()
			defer cancel()

			params := &TestContextDoWithParams{
				wantErr:   tt.wantErr,
				errMsg:    tt.errMsg,
				wantCalls: tt.wantCalls,
				maxCalls:  tt.maxCalls,
			}

			tt.testFunc(t, p, ctx, cancel, params)
		})
	}
}

type TestContextDoWithParams struct {
	wantErr   bool
	errMsg    string
	wantCalls int
	maxCalls  int
}

func TestExponentialBackoff(t *testing.T) {
	tests := []struct {
		name        string
		backoff     *exponentialBackoff
		attempts    []int
		minDelays   []time.Duration
		maxDelays   []time.Duration
		exactDelays []time.Duration
		description string
	}{
		{
			name: "with jitter",
			backoff: NewExponentialBackoff(
				BaseDelay(10*time.Millisecond),
				Multiplier(2),
				WithJitter(),
			),
			attempts:  []int{0, 1, 2},
			minDelays: []time.Duration{7 * time.Millisecond, 15 * time.Millisecond, 30 * time.Millisecond},
			maxDelays: []time.Duration{13 * time.Millisecond, 25 * time.Millisecond, 50 * time.Millisecond},
		},
		{
			name: "with max delay and jitter",
			backoff: NewExponentialBackoff(
				BaseDelay(10*time.Millisecond),
				Multiplier(10),
				MaxDelay(50*time.Millisecond),
			),
			attempts:  []int{0, 10},
			minDelays: []time.Duration{7 * time.Millisecond, 37 * time.Millisecond},
			maxDelays: []time.Duration{13 * time.Millisecond, 63 * time.Millisecond},
		},
		{
			name: "without jitter",
			backoff: NewExponentialBackoff(
				BaseDelay(10*time.Millisecond),
				Multiplier(2),
			),
			attempts:    []int{0, 1, 2},
			exactDelays: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 40 * time.Millisecond},
		},
		{
			name: "custom jitter ratio",
			backoff: NewExponentialBackoff(
				BaseDelay(100*time.Millisecond),
				JitterRatio(0.5),
			),
			attempts:  []int{1},
			minDelays: []time.Duration{100 * time.Millisecond},
			maxDelays: []time.Duration{300 * time.Millisecond},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "without jitter" {
				tt.backoff.jitter = false
			}

			for i, attempt := range tt.attempts {
				delay := tt.backoff.Delay(attempt)

				if len(tt.exactDelays) > 0 {
					assert.Equal(t, tt.exactDelays[i], delay)
				} else {
					assert.True(t, delay >= tt.minDelays[i] && delay <= tt.maxDelays[i],
						"attempt %d: delay %v should be between %v and %v", attempt, delay, tt.minDelays[i], tt.maxDelays[i])
				}
			}
		})
	}
}

func TestLinearBackoff(t *testing.T) {
	tests := []struct {
		name        string
		backoff     *linearBackoff
		attempts    []int
		expected    []time.Duration
		description string
	}{
		{
			name: "basic linear increment",
			backoff: NewLinearBackoff(
				LinearBaseDelay(10*time.Millisecond),
				LinearIncrement(5*time.Millisecond),
			),
			attempts: []int{0, 1, 2},
			expected: []time.Duration{10 * time.Millisecond, 15 * time.Millisecond, 20 * time.Millisecond},
		},
		{
			name: "with max delay",
			backoff: NewLinearBackoff(
				LinearBaseDelay(10*time.Millisecond),
				LinearIncrement(100*time.Millisecond),
				LinearMaxDelay(50*time.Millisecond),
			),
			attempts: []int{0, 10},
			expected: []time.Duration{10 * time.Millisecond, 50 * time.Millisecond},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, attempt := range tt.attempts {
				delay := tt.backoff.Delay(attempt)
				assert.Equal(t, tt.expected[i], delay, "attempt %d", attempt)
			}
		})
	}
}

func TestConstantBackoff(t *testing.T) {
	tests := []struct {
		name        string
		delay       time.Duration
		attempts    []int
		expected    time.Duration
		description string
	}{
		{
			name:     "constant delay",
			delay:    50 * time.Millisecond,
			attempts: []int{0, 10},
			expected: 50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConstantBackoff(tt.delay)
			for _, attempt := range tt.attempts {
				assert.Equal(t, tt.expected, c.Delay(attempt), "attempt %d", attempt)
			}
		})
	}
}

func TestConvenience_Functions(t *testing.T) {
	tests := []struct {
		name        string
		testFunc    func(*testing.T, *ConvenienceTestParams)
		wantErr     bool
		wantCalls   int
		description string
	}{
		{
			name: "Do convenience function",
			testFunc: func(t *testing.T, p *ConvenienceTestParams) {
				calls := 0
				err := Do(func() error {
					calls++
					if calls < 3 {
						return errors.New("temp")
					}
					return nil
				})
				assert.NoError(t, err)
				assert.Equal(t, p.wantCalls, calls)
			},
			wantErr:     false,
			wantCalls:   3,
			description: "should retry with default policy",
		},
		{
			name: "DoWithContext convenience function",
			testFunc: func(t *testing.T, p *ConvenienceTestParams) {
				ctx := context.Background()
				calls := 0
				err := DoWithContext(ctx, func(ctx context.Context) error {
					calls++
					if calls < 2 {
						return errors.New("temp")
					}
					return nil
				})
				assert.NoError(t, err)
				assert.Equal(t, p.wantCalls, calls)
			},
			wantErr:     false,
			wantCalls:   2,
			description: "should retry with context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &ConvenienceTestParams{
				wantErr:   tt.wantErr,
				wantCalls: tt.wantCalls,
			}
			tt.testFunc(t, params)
		})
	}
}

type ConvenienceTestParams struct {
	wantErr   bool
	wantCalls int
}

func TestRetryIf(t *testing.T) {
	tests := []struct {
		name        string
		testFunc    func(*testing.T, *Policy, *RetryIfTestParams)
		wantErr     bool
		wantCalls   int
		description string
	}{
		{
			name: "retry on specific errors",
			testFunc: func(t *testing.T, p *Policy, params *RetryIfTestParams) {
				retryableErr := errors.New("retryable")
				nonRetryableErr := errors.New("non-retryable")
				calls := 0

				check := func(err error) bool {
					return err == retryableErr
				}

				fn := RetryIf(func(ctx context.Context) error {
					calls++
					if calls == 1 {
						return retryableErr
					}
					return nonRetryableErr
				}, check)

				err := p.DoWithContext(context.Background(), fn)
				assert.Error(t, err)
				assert.Equal(t, params.wantCalls, calls)
			},
			wantErr:     true,
			wantCalls:   2,
			description: "should stop on non-retryable error",
		},
		{
			name: "stop on non-retryable error",
			testFunc: func(t *testing.T, p *Policy, params *RetryIfTestParams) {
				calls := 0
				check := func(err error) bool {
					return err.Error() != "stop"
				}

				fn := RetryIf(func(ctx context.Context) error {
					calls++
					if calls == 3 {
						return errors.New("stop")
					}
					return errors.New("continue")
				}, check)

				err := p.DoWithContext(context.Background(), fn)
				assert.Error(t, err)
				assert.Equal(t, params.wantCalls, calls)
			},
			wantErr:     true,
			wantCalls:   3,
			description: "should stop at specific error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPolicy()
			params := &RetryIfTestParams{
				wantErr:   tt.wantErr,
				wantCalls: tt.wantCalls,
			}
			tt.testFunc(t, p, params)
		})
	}
}

type RetryIfTestParams struct {
	wantErr   bool
	wantCalls int
}

func TestRetrySpecificErrors(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")

	tests := []struct {
		name     string
		errors   []error
		testErr  error
		wantBool bool
	}{
		{
			name:     "matching first error",
			errors:   []error{err1, err2},
			testErr:  err1,
			wantBool: true,
		},
		{
			name:     "matching second error",
			errors:   []error{err1, err2},
			testErr:  err2,
			wantBool: true,
		},
		{
			name:     "non-matching error",
			errors:   []error{err1, err2},
			testErr:  err3,
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := RetrySpecificErrors(tt.errors...)
			assert.Equal(t, tt.wantBool, check(tt.testErr))
		})
	}
}

func TestNonRetryableError(t *testing.T) {
	baseErr := errors.New("base error")
	wrapped := &nonRetryableError{err: baseErr}

	tests := []struct {
		name      string
		err       error
		contains  string
		unwrapped error
	}{
		{
			name:     "error message",
			err:      wrapped,
			contains: "non-retryable",
		},
		{
			name:      "unwrapped error",
			err:       wrapped,
			unwrapped: baseErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.contains != "" {
				assert.Contains(t, tt.err.Error(), tt.contains)
			}
			if tt.unwrapped != nil {
				wrapped, ok := tt.err.(*nonRetryableError)
				assert.True(t, ok)
				assert.Equal(t, tt.unwrapped, wrapped.Unwrap())
			}
		})
	}
}

func TestBackoff_Strategies(t *testing.T) {
	strategies := []struct {
		name     BackoffStrategy
		expected []time.Duration
		jittered bool
	}{
		{NewConstantBackoff(50 * time.Millisecond), []time.Duration{50 * time.Millisecond, 50 * time.Millisecond, 50 * time.Millisecond}, false},
		{NewLinearBackoff(LinearBaseDelay(10*time.Millisecond), LinearIncrement(10*time.Millisecond)), []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond}, false},
		{NewExponentialBackoff(BaseDelay(10*time.Millisecond), Multiplier(2), MaxDelay(100*time.Millisecond)), []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 40 * time.Millisecond}, true},
	}

	for _, s := range strategies {
		t.Run(fmt.Sprintf("%T", s.name), func(t *testing.T) {
			for i, expected := range s.expected {
				delay := s.name.Delay(i)
				if s.jittered {
					assert.InDelta(t, float64(expected), float64(delay), float64(expected)*0.3)
				} else {
					assert.Equal(t, expected, delay)
				}
			}
		})
	}
}

func TestDefaultPolicy(t *testing.T) {
	p := NewPolicy()

	tests := []struct {
		name      string
		field     string
		expected  interface{}
		actual    interface{}
		checkType bool
		typeName  string
	}{
		{
			name:     "max attempts default",
			field:    "maxAttempts",
			expected: 3,
			actual:   p.maxAttempts,
		},
		{
			name:     "max duration default",
			field:    "maxDuration",
			expected: time.Minute,
			actual:   p.maxDuration,
		},
		{
			name:      "backoff type",
			field:     "backoff",
			checkType: true,
			typeName:  "*retry.exponentialBackoff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.checkType {
				assert.IsType(t, &exponentialBackoff{}, p.backoff)
			} else {
				assert.Equal(t, tt.expected, tt.actual)
			}
		})
	}
}
