package echo

import (
	"github.com/haipham22/govern/http/middleware"
)

// TrimStrings is a convenience wrapper around middleware.TrimStrings.
//
// This middleware automatically trims whitespace from string fields in JSON
// request bodies before they reach your handlers.
//
// Example:
//
//	e := echo.New()
//	e.POST("/api/register",
//	    echo.TrimStrings(handleRegister),
//	)
//
// Features:
// - Recursively trims all string values in JSON request bodies
// - Handles nested objects and arrays
// - Uses sonic for high-performance JSON parsing
// - Graceful fallback for invalid JSON (passes through unchanged)
// - Fast-path optimization for strings without whitespace
//
// Performance:
// - Uses sonic for fast JSON operations (~2x faster than encoding/json)
// - Minimal overhead for requests without whitespace
var TrimStrings = middleware.TrimStrings
