// Package apperrors provides typed application errors with error codes and HTTP
// status mapping. It is a lightweight, dependency-free replacement for
// github.com/haipham22/govern/errors used by this sample app.
//
// The API mirrors govern/errors (Code constants, WrapCode, NewCode, GetCode,
// IsCode, sentinel errors) so existing call sites migrate by changing only the
// import path and alias.
package apperrors

// Code is a stable, machine-readable error category.
type Code string

// Well-known error codes. String values are part of the API contract (logged,
// returned to clients) and must not change.
const (
	CodeInternal      Code = "INTERNAL"
	CodeInvalid       Code = "INVALID"
	CodeNotFound      Code = "NOT_FOUND"
	CodeAlreadyExists Code = "ALREADY_EXISTS"
	CodeUnauthorized  Code = "UNAUTHORIZED"
	CodeForbidden     Code = "FORBIDDEN"
	CodeConflict      Code = "CONFLICT"
	CodeRateLimit     Code = "RATE_LIMIT"
)

// HTTPStatus maps an error code to its canonical HTTP response status.
// Unknown codes default to 500 Internal Server Error.
func (c Code) HTTPStatus() int {
	switch c {
	case CodeInvalid:
		return 400 // http.StatusBadRequest
	case CodeUnauthorized:
		return 401 // http.StatusUnauthorized
	case CodeForbidden:
		return 403 // http.StatusForbidden
	case CodeNotFound:
		return 404 // http.StatusNotFound
	case CodeConflict, CodeAlreadyExists:
		return 409 // http.StatusConflict
	case CodeRateLimit:
		return 429 // http.StatusTooManyRequests
	default:
		return 500 // http.StatusInternalServerError
	}
}
