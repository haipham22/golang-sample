package apperrors

import "errors"

// FieldError describes a single field-level validation failure.
type FieldError struct {
	Property string `json:"property" example:"email"`
	Msg      string `json:"msg" example:"email is required"`
}

// Response is the standard JSON error body returned by the HTTP error handler.
// The shape (msg, error, path) matches the legacy response format for backward
// compatibility; request_id and errors are optional.
type Response struct {
	Msg       string       `json:"msg"`
	Error     string       `json:"error"`
	Path      string       `json:"path,omitempty"`
	RequestID string       `json:"request_id,omitempty"`
	Errors    []FieldError `json:"errors,omitempty"`
}

// ClientMessage returns a sanitized, client-safe message for a code. Internal
// errors never leak details; 4xx codes return a concise generic message.
func (c Code) ClientMessage() string {
	switch c {
	case CodeInvalid:
		return "invalid request parameters"
	case CodeNotFound:
		return "Resource not found"
	case CodeUnauthorized:
		return "Unauthorized"
	case CodeForbidden:
		return "Forbidden"
	case CodeConflict, CodeAlreadyExists:
		return "Resource already exists"
	case CodeRateLimit:
		return "Too many requests"
	default:
		return "Internal Server Error"
	}
}

// NewBody builds the standard JSON error body.
func NewBody(message, path, requestID string) Response {
	return Response{
		Msg:       message,
		Error:     message,
		Path:      path,
		RequestID: requestID,
	}
}

// Resolve maps an error to an HTTP status and a standard Response body.
// status defaults to 500 when the error carries no apperrors Code.
func Resolve(err error, path, requestID string) (status int, body Response) {
	code, ok := GetCode(err)
	if !ok {
		return 500, NewBody("Internal Server Error", path, requestID)
	}

	msg := code.ClientMessage()
	body = NewBody(msg, path, requestID)

	var appErr *Error
	if code == CodeInvalid && errors.As(err, &appErr) && len(appErr.Errors) > 0 {
		body.Msg = appErr.Errors[0].Msg
		body.Error = appErr.Errors[0].Msg
		body.Errors = appErr.Errors
	}

	return code.HTTPStatus(), body
}
