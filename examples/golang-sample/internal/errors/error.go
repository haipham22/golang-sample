package apperrors

import (
	stderrors "errors"
	"fmt"
)

// Error is a typed error carrying a Code and an optional underlying cause.
// A nil *Error must never be constructed; use the constructors below.
type Error struct {
	// Code categorizes the error for HTTP status mapping and logging.
	Code Code
	// Err is the wrapped underlying error; may be nil for sentinel errors.
	Err error
	// message is an optional human-readable detail. When empty, Error() falls
	// back to the wrapped error or the code itself.
	message string
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	switch {
	case e.message != "" && e.Err != nil:
		return fmt.Sprintf("%s: %v", e.message, e.Err)
	case e.message != "":
		return e.message
	case e.Err != nil:
		return e.Err.Error()
	default:
		return string(e.Code)
	}
}

// Unwrap returns the underlying error so stderrors.Is and stderrors.As traverse
// the chain.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Message returns the optional human-readable detail (may be empty).
func (e *Error) Message() string {
	if e == nil {
		return ""
	}
	return e.message
}

// New creates an *Error with a code and message (no wrapped cause).
func New(code Code, message string) *Error {
	return &Error{Code: code, message: message}
}

// NewCode is the govern/errors-compatible constructor returning an error
// interface.
func NewCode(code Code, message string) error {
	return New(code, message)
}

// Wrap creates an *Error that wraps an underlying error with a code.
func Wrap(code Code, err error) *Error {
	return &Error{Code: code, Err: err}
}

// WrapCode is the govern/errors-compatible wrapper returning an error interface.
// If err is nil it returns nil.
func WrapCode(code Code, err error) error {
	if err == nil {
		return nil
	}
	return Wrap(code, err)
}

// GetCode extracts the Code from an error chain. Returns the code and true if
// an *Error is present anywhere in the chain; otherwise ("", false).
func GetCode(err error) (Code, bool) {
	var e *Error
	if stderrors.As(err, &e) {
		return e.Code, true
	}
	return "", false
}

// IsCode reports whether err (or any error in its chain) carries the given code.
func IsCode(err error, code Code) bool {
	c, ok := GetCode(err)
	if !ok {
		return false
	}
	return c == code
}

// Sentinel errors for common cases. These are *Error values so stderrors.Is and
// GetCode recognize them across a wrapped chain.
var (
	ErrInternal     = New(CodeInternal, "internal error")
	ErrInvalid      = New(CodeInvalid, "invalid input")
	ErrNotFound     = New(CodeNotFound, "resource not found")
	ErrUnauthorized = New(CodeUnauthorized, "unauthorized")
)
