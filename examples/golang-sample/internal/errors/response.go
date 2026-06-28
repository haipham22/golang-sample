package apperrors

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

// Resolve maps an error to an HTTP status and a standard Response body.
// status defaults to 500 when the error carries no apperrors Code.
//
// For CodeInvalid the caller may enrich body.Errors with field-level details
// (e.g. from a validator.ValidationError) and override body.Msg accordingly.
func Resolve(err error, path, requestID string) (status int, body Response) {
	code, ok := GetCode(err)
	if !ok {
		return 500, Response{
			Msg:       "Internal Server Error",
			Error:     "Internal Server Error",
			Path:      path,
			RequestID: requestID,
		}
	}

	msg := code.ClientMessage()
	return code.HTTPStatus(), Response{
		Msg:       msg,
		Error:     msg,
		Path:      path,
		RequestID: requestID,
	}
}
