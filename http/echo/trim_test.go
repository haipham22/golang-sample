package echo

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bytedance/sonic"
)

func TestTrimStrings_Integration(t *testing.T) {
	t.Run("trims whitespace in JSON request body", func(t *testing.T) {
		e := echo.New()

		// Request with whitespace in fields
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
		var receivedData map[string]interface{}
		handler := func(c echo.Context) error {
			err := c.Bind(&receivedData)
			require.NoError(t, err)
			return c.JSON(http.StatusOK, receivedData)
		}

		// Execute middleware
		err := TrimStrings(handler)(c)

		assert.NoError(t, err)
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
		handler := func(c echo.Context) error {
			err := c.Bind(&receivedData)
			require.NoError(t, err)
			return c.JSON(http.StatusOK, receivedData)
		}

		err := TrimStrings(handler)(c)

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
		handler := func(c echo.Context) error {
			err := c.Bind(&receivedData)
			require.NoError(t, err)
			return c.JSON(http.StatusOK, receivedData)
		}

		err := TrimStrings(handler)(c)

		assert.NoError(t, err)
		tags := receivedData["tags"].([]interface{})
		assert.Equal(t, "tag1", tags[0])
		assert.Equal(t, "tag2", tags[1])
		assert.Equal(t, "tag3", tags[2])
	})

	t.Run("passes through non-JSON requests", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte("plain text")))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerCalled := false
		handler := func(c echo.Context) error {
			handlerCalled = true
			return c.String(http.StatusOK, "ok")
		}

		err := TrimStrings(handler)(c)

		assert.NoError(t, err)
		assert.True(t, handlerCalled)
	})

	t.Run("skips empty body", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader([]byte{}))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handlerCalled := false
		handler := func(c echo.Context) error {
			handlerCalled = true
			return c.String(http.StatusOK, "ok")
		}

		err := TrimStrings(handler)(c)

		assert.NoError(t, err)
		assert.True(t, handlerCalled)
	})
}

func TestTrimStrings_EchoIntegration(t *testing.T) {
	t.Run("works with Echo routes", func(t *testing.T) {
		e := echo.New()

		// Setup route with TrimStrings middleware
		e.POST("/api/register", func(c echo.Context) error {
			var req struct {
				Username string `json:"username"`
				Email    string `json:"email"`
			}
			if err := c.Bind(&req); err != nil {
				return err
			}

			// Verify trimming happened
			assert.Equal(t, "testuser", req.Username)
			assert.Equal(t, "test@example.com", req.Email)

			return c.JSON(http.StatusOK, req)
		}, TrimStrings)

		// Create request with whitespace
		reqBody := map[string]string{
			"username": "  testuser  ",
			"email":    "  test@example.com  ",
		}
		bodyBytes, _ := sonic.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("works with Echo group", func(t *testing.T) {
		e := echo.New()
		g := e.Group("/api/v1")
		g.Use(TrimStrings)

		g.POST("/register", func(c echo.Context) error {
			var req struct {
				Username string `json:"username"`
			}
			if err := c.Bind(&req); err != nil {
				return err
			}

			assert.Equal(t, "testuser", req.Username)
			return c.JSON(http.StatusOK, req)
		})

		reqBody := map[string]string{"username": "  testuser  "}
		bodyBytes, _ := sonic.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/register", bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
