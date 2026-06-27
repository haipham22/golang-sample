package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/labstack/echo/v4"
)

// TrimStrings middleware automatically trims whitespace from string fields
// in JSON request bodies before they reach the handler.
//
// This ensures that all string inputs are cleaned of leading/trailing whitespace
// without requiring manual trimming in each handler or service.
func TrimStrings(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Skip if no body or empty body
		body := c.Request().Body
		if body == nil || body == http.NoBody {
			return next(c)
		}

		// Read the request body
		bodyBytes, err := io.ReadAll(body)
		if err != nil {
			return next(c)
		}

		// Skip if body is empty
		if len(bodyBytes) == 0 {
			return next(c)
		}

		// Parse JSON using sonic for faster performance
		var data map[string]interface{}
		if err := sonic.Unmarshal(bodyBytes, &data); err != nil {
			// Invalid JSON - let validation handler deal with it
			c.Request().Body = io.NopCloser(bytes.NewReader(bodyBytes))
			return next(c)
		}

		// Trim all string values recursively
		trimStringsRecursive(data)

		// Re-encode and set back to request using sonic
		encoded, err := sonic.Marshal(data)
		if err != nil {
			// Encoding failed - use original body
			c.Request().Body = io.NopCloser(bytes.NewReader(bodyBytes))
			return next(c)
		}

		// Set the cleaned body back to the request
		c.Request().Body = io.NopCloser(bytes.NewReader(encoded))

		return next(c)
	}
}

// trimStringsRecursive recursively trims all string values in the data structure
func trimStringsRecursive(v interface{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, v := range val {
			if str, ok := v.(string); ok {
				val[k] = trimString(str)
			} else {
				trimStringsRecursive(v)
			}
		}
	case []interface{}:
		for i, item := range val {
			if str, ok := item.(string); ok {
				val[i] = trimString(str)
			} else {
				trimStringsRecursive(item)
			}
		}
	}
}

// trimString trims leading and trailing whitespace from a string
func trimString(s string) string {
	// Fast path for empty strings
	if s == "" {
		return s
	}

	// Fast path for strings without leading/trailing spaces
	if s[0] != ' ' && s[len(s)-1] != ' ' {
		return s
	}

	start := 0
	end := len(s)

	// Trim leading spaces
	for start < end && s[start] == ' ' {
		start++
	}

	// Trim trailing spaces
	for end > start && s[end-1] == ' ' {
		end--
	}

	return s[start:end]
}
