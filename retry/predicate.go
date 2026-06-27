package retry

import (
	"context"
	"errors"
)

// IsRetryable returns true if the error should trigger a retry.
type IsRetryable func(error) bool

// AlwaysRetry always returns true.
func AlwaysRetry(error) bool {
	return true
}

// NeverRetry always returns false.
func NeverRetry(error) bool {
	return false
}

// RetrySpecificErrors returns a function that checks if error matches.
func RetrySpecificErrors(errs ...error) IsRetryable {
	return func(err error) bool {
		for _, e := range errs {
			if err == e {
				return true
			}
		}
		return false
	}
}

// RetryIf wraps a FuncWithContext to only retry on specific errors.
func RetryIf(fn FuncWithContext, check IsRetryable) FuncWithContext {
	return func(ctx context.Context) error {
		err := fn(ctx)
		if err != nil && check(err) {
			return err
		}
		if err != nil {
			return &nonRetryableError{err: err}
		}
		return nil
	}
}

// nonRetryableError wraps errors that should stop retries.
type nonRetryableError struct {
	err error
}

func (e *nonRetryableError) Error() string {
	return "non-retryable: " + e.err.Error()
}

func (e *nonRetryableError) Unwrap() error {
	return e.err
}

// IsNonRetryable checks if an error is non-retryable.
func IsNonRetryable(err error) bool {
	var nre *nonRetryableError
	return errors.As(err, &nre)
}
