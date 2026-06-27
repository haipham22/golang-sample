package errors

import (
	"errors"
	"fmt"

	pkgErrors "github.com/pkg/errors"
)

// New creates a new error with the given message.
func New(message string) error {
	return errors.New(message)
}

// Errorf formats according to a format specifier and returns the error.
func Errorf(format string, args ...interface{}) error {
	return pkgErrors.Errorf(format, args...)
}

// Is reports whether err is equal to target.
func Is(err, target error) bool {
	return pkgErrors.Is(err, target)
}

// As finds the first error in err's chain that matches target.
func As(err error, target interface{}) bool {
	return pkgErrors.As(err, target)
}

// Unwrap returns the result of calling Unwrap method on err.
func Unwrap(err error) error {
	return pkgErrors.Unwrap(err)
}

// Join joins errors into a single error.
func Join(errs ...error) error {
	return errors.Join(errs...)
}

// ErrorCode represents an error code type.
type ErrorCode string

const (
	CodeInternal      ErrorCode = "INTERNAL"
	CodeInvalid       ErrorCode = "INVALID"
	CodeNotFound      ErrorCode = "NOT_FOUND"
	CodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	CodeUnauthorized  ErrorCode = "UNAUTHORIZED"
	CodeForbidden     ErrorCode = "FORBIDDEN"
	CodeConflict      ErrorCode = "CONFLICT"
	CodeRateLimit     ErrorCode = "RATE_LIMIT"
)

// ErrorWithCode is an error that includes a code.
type ErrorWithCode struct {
	Code ErrorCode
	Err  error
}

// Error implements the error interface.
func (e *ErrorWithCode) Error() string {
	if e.Err == nil {
		return string(e.Code)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Err.Error())
}

// Unwrap returns the underlying error.
func (e *ErrorWithCode) Unwrap() error {
	return e.Err
}

// NewCode creates a new error with a code.
func NewCode(code ErrorCode, message string) error {
	return &ErrorWithCode{Code: code, Err: New(message)}
}

// WrapCode wraps an error with a code.
func WrapCode(code ErrorCode, err error) error {
	if err == nil {
		return nil
	}
	return &ErrorWithCode{Code: code, Err: err}
}

// GetCode extracts the code from an error if present.
func GetCode(err error) (ErrorCode, bool) {
	var e *ErrorWithCode
	if errors.As(err, &e) {
		return e.Code, true
	}
	return "", false
}

// IsCode checks if an error has a specific code.
func IsCode(err error, code ErrorCode) bool {
	c, ok := GetCode(err)
	return ok && c == code
}

// Common predefined errors.
var (
	ErrInternal     = &ErrorWithCode{Code: CodeInternal, Err: New("internal error")}
	ErrInvalid      = &ErrorWithCode{Code: CodeInvalid, Err: New("invalid input")}
	ErrNotFound     = &ErrorWithCode{Code: CodeNotFound, Err: New("resource not found")}
	ErrUnauthorized = &ErrorWithCode{Code: CodeUnauthorized, Err: New("unauthorized")}
)
