package errors

import (
	"fmt"

	errorPkg "github.com/pkg/errors"
)

type AppError struct {
	BaseErr   error
	Message   string   `json:"message"`
	ErrorCode string   `json:"error_code"`
	HTTPCode  int      `json:"http_code"`
	RequestID *string  `json:"request_id,omitempty"`
	Issues    []string `json:"issues,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.BaseErr
}

func (e *AppError) WithRequestID(rID string) *AppError {
	e.RequestID = &rID
	return e
}

func New(httpCode int, message string, errorCode string) *AppError {
	return &AppError{
		BaseErr:   fmt.Errorf("error: %s", message),
		HTTPCode:  httpCode,
		Message:   message,
		ErrorCode: errorCode,
	}
}

func Wrapf(err error, format string, args ...interface{}) error {
	return errorPkg.Wrapf(err, format, args...)
}

func As(err error, target interface{}) bool {
	return errorPkg.As(err, target)
}

func Is(err, target error) bool {
	return errorPkg.Is(err, target)
}
