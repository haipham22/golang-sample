package apperrors

// Convenience constructors for common error categories. Each returns a typed
// *Error carrying the appropriate Code and a descriptive message. Wrap an
// underlying cause with Wrap/WrapCode when one exists.

// InvalidInput returns an error for malformed/invalid request input.
func InvalidInput(message string) *Error {
	return New(CodeInvalid, message)
}

// NotFound returns an error indicating the given resource was not found.
func NotFound(resource string) *Error {
	return New(CodeNotFound, resource+" not found")
}

// Unauthorized returns an error for failed authentication.
func Unauthorized(message string) *Error {
	return New(CodeUnauthorized, message)
}

// Forbidden returns an error for failed authorization.
func Forbidden(message string) *Error {
	return New(CodeForbidden, message)
}

// Conflict returns an error for a duplicate/conflicting resource.
func Conflict(resource string) *Error {
	return New(CodeConflict, resource+" already exists")
}

// Internal returns an error for unexpected internal failures.
func Internal(message string) *Error {
	return New(CodeInternal, message)
}
