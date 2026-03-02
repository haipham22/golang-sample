package model

import "errors"

// Domain-specific errors for User entity
var (
	ErrUsernameRequired = errors.New("username is required")
	ErrEmailRequired    = errors.New("email is required")
	ErrUsernameTooShort = errors.New("username must be at least 3 characters")
	ErrUsernameTooLong  = errors.New("username must be at most 50 characters")
	ErrInvalidEmail     = errors.New("invalid email format")
)
