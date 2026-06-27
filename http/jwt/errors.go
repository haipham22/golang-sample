package jwt

import (
	"strings"

	"github.com/haipham22/govern/errors"
)

var (
	// Configuration errors.
	ErrSecretRequired         = errors.NewCode(errors.CodeInvalid, "JWT secret is required")
	ErrSigningMethodRequired  = errors.NewCode(errors.CodeInvalid, "JWT signing method is required")
	ErrInvalidAccessDuration  = errors.NewCode(errors.CodeInvalid, "JWT access duration must be positive")
	ErrInvalidRefreshDuration = errors.NewCode(errors.CodeInvalid, "JWT refresh duration must be positive")

	// Token errors.
	ErrTokenMissing            = errors.NewCode(errors.CodeUnauthorized, "authorization token is required")
	ErrTokenInvalid            = errors.NewCode(errors.CodeUnauthorized, "invalid authorization token")
	ErrTokenExpired            = errors.NewCode(errors.CodeUnauthorized, "authorization token has expired")
	ErrTokenMalformed          = errors.NewCode(errors.CodeUnauthorized, "malformed authorization token")
	ErrSignatureInvalid        = errors.NewCode(errors.CodeUnauthorized, "invalid token signature")
	ErrUnexpectedSigningMethod = errors.NewCode(errors.CodeUnauthorized, "unexpected signing method")
)

// HandleJWTError converts JWT library errors to our errors.
func HandleJWTError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "token is expired"):
		return ErrTokenExpired
	case strings.Contains(errStr, "signature is invalid"):
		return ErrSignatureInvalid
	case strings.Contains(errStr, "malformed"):
		return ErrTokenMalformed
	default:
		// Return a new error with code instead of wrapping
		return errors.NewCode(errors.CodeUnauthorized, "invalid authorization token")
	}
}
