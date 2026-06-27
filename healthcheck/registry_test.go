package healthcheck

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRegistry_Register(t *testing.T) {
	r := New()

	r.Register("check1", func(ctx context.Context) error {
		return nil
	})

	r.Register("check2", func(ctx context.Context) error {
		return errors.New("failed")
	}, WithTimeout(2*time.Second))

	// Should panic on duplicate name
	assert.Panics(t, func() {
		r.Register("check1", func(ctx context.Context) error {
			return nil
		})
	})
}

func TestRegistry_Unregister(t *testing.T) {
	r := New()

	r.Register("check1", func(ctx context.Context) error {
		return nil
	})

	r.Unregister("check1")

	// Should allow re-register after unregister
	assert.NotPanics(t, func() {
		r.Register("check1", func(ctx context.Context) error {
			return nil
		})
	})
}

// Validation helper functions for table-driven tests
type responseValidator func(*testing.T, *Response)

func validateAllChecksPassing(checks ...string) responseValidator {
	return func(t *testing.T, resp *Response) {
		t.Helper()
		assert.Len(t, resp.Checks, len(checks))
		for _, check := range checks {
			assert.Equal(t, StatusPassing, resp.Checks[check].Status)
		}
	}
}

func validateMixedStatus(checksPassing []string, checksFailing map[string]string) responseValidator {
	return func(t *testing.T, resp *Response) {
		t.Helper()
		totalChecks := len(checksPassing) + len(checksFailing)
		assert.Len(t, resp.Checks, totalChecks)

		for _, check := range checksPassing {
			assert.Equal(t, StatusPassing, resp.Checks[check].Status)
		}

		for check, msg := range checksFailing {
			assert.Equal(t, StatusFailing, resp.Checks[check].Status)
			assert.Equal(t, msg, resp.Checks[check].Message)
		}
	}
}

func validateSingleCheckFailing(status Status, message string) responseValidator {
	return func(t *testing.T, resp *Response) {
		t.Helper()
		assert.Len(t, resp.Checks, 1)
		assert.Equal(t, status, resp.Checks["slow"].Status)
		assert.Equal(t, message, resp.Checks["slow"].Message)
	}
}

func validatePanicCheck(checkName string) responseValidator {
	return func(t *testing.T, resp *Response) {
		t.Helper()
		assert.Equal(t, StatusFailing, resp.Checks[checkName].Status)
		assert.Contains(t, resp.Checks[checkName].Message, "panic")
	}
}

func validateOnlyOverallStatus(status Status) responseValidator {
	return func(t *testing.T, resp *Response) {
		t.Helper()
		assert.Equal(t, status, resp.Status)
	}
}

func TestRegistry_Run(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*Registry)
		wantStatus Status
		validate   responseValidator
		useCancel  bool
	}{
		{
			name:       "all checks passing",
			setup:      setupMultipleChecks(nil),
			wantStatus: StatusPassing,
			validate:   validateAllChecksPassing("db", "cache"),
		},
		{
			name:       "one check failing",
			setup:      setupMultipleChecks(map[string]string{"cache": "connection failed"}),
			wantStatus: StatusFailing,
			validate:   validateMixedStatus([]string{"db"}, map[string]string{"cache": "connection failed"}),
		},
		{
			name:       "check timeout",
			setup:      setupSlowCheck(10*time.Second, 100*time.Millisecond),
			wantStatus: StatusFailing,
			validate:   validateSingleCheckFailing(StatusFailing, "timeout"),
		},
		{
			name:       "context canceled during run",
			setup:      setupSlowCheck(5*time.Second, 0),
			wantStatus: StatusFailing,
			validate:   validateOnlyOverallStatus(StatusFailing),
			useCancel:  true,
		},
		{
			name:       "panic with DisablePanic option",
			setup:      setupPanicCheck(true),
			wantStatus: StatusFailing,
			validate:   validatePanicCheck("panic"),
		},
		{
			name:       "panic without DisablePanic option",
			setup:      setupPanicCheck(false),
			wantStatus: StatusFailing,
			validate:   validatePanicCheck("panic"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New()
			tt.setup(r)

			var resp *Response
			if tt.useCancel {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(50 * time.Millisecond)
					cancel()
				}()
				resp = r.Run(ctx)
			} else {
				resp = r.Run(context.Background())
			}

			assert.Equal(t, tt.wantStatus, resp.Status)
			if tt.validate != nil {
				tt.validate(t, resp)
			}
		})
	}
}

// Setup helper functions for test cases
func setupMultipleChecks(failingChecks map[string]string) func(*Registry) {
	return func(r *Registry) {
		r.Register("db", func(ctx context.Context) error {
			return nil
		})
		if failingChecks != nil {
			for name, msg := range failingChecks {
				r.Register(name, func(ctx context.Context) error {
					return errors.New(msg)
				})
			}
		} else {
			r.Register("cache", func(ctx context.Context) error {
				return nil
			})
		}
	}
}

func setupSlowCheck(sleepDuration, timeout time.Duration) func(*Registry) {
	return func(r *Registry) {
		check := func(ctx context.Context) error {
			select {
			case <-time.After(sleepDuration):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		if timeout > 0 {
			r.Register("slow", check, WithTimeout(timeout))
		} else {
			r.Register("slow", check)
		}
	}
}

func setupPanicCheck(disablePanic bool) func(*Registry) {
	return func(r *Registry) {
		check := func(ctx context.Context) error {
			panic("test panic")
		}

		if disablePanic {
			r.Register("panic", check, DisablePanic())
		} else {
			r.Register("panic", check)
		}
	}
}

func TestRegistry_Run_TimeoutOption(t *testing.T) {
	r := New()

	r.Register("fast", func(ctx context.Context) error {
		return nil
	}, WithTimeout(100*time.Millisecond))

	r.Register("slow", func(ctx context.Context) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	}, WithTimeout(100*time.Millisecond))

	resp := r.Run(context.Background())
	assert.Equal(t, StatusPassing, resp.Checks["fast"].Status)
	assert.Equal(t, StatusFailing, resp.Checks["slow"].Status)
	assert.Equal(t, "timeout", resp.Checks["slow"].Message)
}

func TestRegistry_Run_Empty(t *testing.T) {
	r := New()

	resp := r.Run(context.Background())
	assert.Equal(t, StatusPassing, resp.Status)
	assert.Empty(t, resp.Checks)
}

func TestCheckResult_Duration(t *testing.T) {
	r := New()

	r.Register("delayed", func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	resp := r.Run(context.Background())
	assert.True(t, resp.Checks["delayed"].Duration >= 50*time.Millisecond)
}
