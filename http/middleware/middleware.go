package middleware

import "net/http"

// Middleware function type.
type Middleware func(http.Handler) http.Handler
