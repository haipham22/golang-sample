package errors

import "net/http"

var (
	ErrInternalServerError = New(http.StatusInternalServerError, "Internal server error", "internal_server_error")
	ErrNotFound            = New(http.StatusNotFound, "Resource not found", "not_found")
	ErrInvalidSchema       = New(http.StatusBadRequest, "Invalid schema", "invalid_schema")

	ErrUserAlreadyExist  = New(http.StatusBadRequest, "ERR_USER_ALREADY_EXIST", "AUTH_001")
	ErrEmailAlreadyExist = New(http.StatusBadRequest, "ERR_EMAIL_ALREADY_EXIST", "AUTH_002")
)
