package apperrors

import (
	"go.uber.org/zap"
)

// LogRequestError logs a request error with structured fields. Server errors
// (status >= 500) are logged at Error level including the underlying error;
// client errors are logged at Warn level. Conflict errors are always Warn and
// never include the raw error (it may leak existence information).
func LogRequestError(log *zap.SugaredLogger, err error, path string, status int) {
	if log == nil {
		return
	}

	code, hasCode := GetCode(err)
	fields := []any{
		"path", path,
		"status", status,
	}

	if hasCode && code == CodeConflict {
		log.Warnw("request error: conflict", fields...)
		return
	}

	if status >= 500 {
		fields = append(fields, "error", err)
		log.Errorw("request error", fields...)
		return
	}

	log.Warnw("request error", fields...)
}
