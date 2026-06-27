package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrimStrings_Middleware(t *testing.T) {
	t.Run("trims whitespace from string fields in JSON body", func(t *testing.T) {
		e := echo.New()

		// Create request with whitespace in fields
		reqBody := map[string]interface{}{
			"username": "  testuser  ",
			"email":    "  test@example.com  ",
			"password": "  password123  ",
		}
		bodyBytes, _ := sonic.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Handler to verify trimmed data
		handlerCalled := false
		var receivedData map[string]interface{}

		next := func(c echo.Context) error {
			handlerCalled = true
			// Parse the body that middleware processed
			bodyBytes, _ := io.ReadAll(c.Request().Body)
			err := sonic.Unmarshal(bodyBytes, &receivedData)
			require.NoError(t, err)
			return nil
		}

		// Execute middleware
		middleware := TrimStrings(next)
		err := middleware(c)

		assert.NoError(t, err)
		assert.True(t, handlerCalled)
		assert.Equal(t, "testuser", receivedData["username"])
		assert.Equal(t, "test@example.com", receivedData["email"])
		assert.Equal(t, "password123", receivedData["password"])
	})

	t.Run("handles nested objects", func(t *testing.T) {
		e := echo.New()

		reqBody := map[string]interface{}{
			"user": map[string]interface{}{
				"name":  "  John Doe  ",
				"email": "  john@example.com  ",
			},
		}
		bodyBytes, _ := sonic.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/user", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		var receivedData map[string]interface{}
		next := func(c echo.Context) error {
			bodyBytes, _ := io.ReadAll(c.Request().Body)
			_ = sonic.Unmarshal(bodyBytes, &receivedData)
			return nil
		}

		middleware := TrimStrings(next)
		err := middleware(c)

		assert.NoError(t, err)
		user := receivedData["user"].(map[string]interface{})
		assert.Equal(t, "John Doe", user["name"])
		assert.Equal(t, "john@example.com", user["email"])
	})

	t.Run("handles arrays", func(t *testing.T) {
		e := echo.New()

		reqBody := map[string]interface{}{
			"tags": []interface{}{"  tag1  ", "  tag2  ", "  tag3  "},
		}
		bodyBytes, _ := sonic.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/tags", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		var receivedData map[string]interface{}
		next := func(c echo.Context) error {
			bodyBytes, _ := io.ReadAll(c.Request().Body)
			_ = sonic.Unmarshal(bodyBytes, &receivedData)
			return nil
		}

		middleware := TrimStrings(next)
		err := middleware(c)

		assert.NoError(t, err)
		tags := receivedData["tags"].([]interface{})
		assert.Equal(t, "tag1", tags[0])
		assert.Equal(t, "tag2", tags[1])
		assert.Equal(t, "tag3", tags[2])
	})

	t.Run("skips empty body", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte{}))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerCalled := false
		next := func(c echo.Context) error {
			handlerCalled = true
			return nil
		}

		middleware := TrimStrings(next)
		err := middleware(c)

		assert.NoError(t, err)
		assert.True(t, handlerCalled)
	})

	t.Run("passes through invalid JSON", func(t *testing.T) {
		e := echo.New()

		invalidJSON := []byte("{invalid json}")
		req := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader(invalidJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerCalled := false
		next := func(c echo.Context) error {
			handlerCalled = true
			// Body should still be readable (original invalid JSON)
			bodyBytes, _ := io.ReadAll(c.Request().Body)
			assert.Equal(t, invalidJSON, bodyBytes)
			return nil
		}

		middleware := TrimStrings(next)
		err := middleware(c)

		assert.NoError(t, err)
		assert.True(t, handlerCalled)
	})
}

func TestTrimString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "trims leading and trailing spaces",
			input:    "  hello world  ",
			expected: "hello world",
		},
		{
			name:     "trims only leading spaces",
			input:    "  hello world",
			expected: "hello world",
		},
		{
			name:     "trims only trailing spaces",
			input:    "hello world  ",
			expected: "hello world",
		},
		{
			name:     "returns empty string for empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "returns same string if no spaces",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "handles string with only spaces",
			input:    "     ",
			expected: "",
		},
		{
			name:     "preserves internal spaces",
			input:    "  hello   world  ",
			expected: "hello   world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := trimString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmarks to compare sonic performance
func BenchmarkTrimStringsMiddleware(b *testing.B) {
	e := echo.New()

	reqBody := map[string]interface{}{
		"username": "  testuser  ",
		"email":    "  test@example.com  ",
		"password": "  password123  ",
		"nested": map[string]interface{}{
			"field1": "  value1  ",
			"field2": "  value2  ",
		},
	}
	bodyBytes, _ := sonic.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		next := func(c echo.Context) error {
			return nil
		}

		middleware := TrimStrings(next)
		_ = middleware(c)
	}
}

func BenchmarkTrimString(b *testing.B) {
	testString := "  hello world  "
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = trimString(testString)
	}
}
