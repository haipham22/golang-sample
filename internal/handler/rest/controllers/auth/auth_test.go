package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	governerrors "github.com/haipham22/govern/errors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	serviceMocks "golang-sample/internal/mocks/service"
	"golang-sample/internal/model"
	schemas "golang-sample/internal/schemas"
	authservice "golang-sample/internal/service/auth"
	apiValidator "golang-sample/internal/validator"
)

func newTestHandler(service *serviceMocks.MockService) *Controller {
	return &Controller{
		service: service,
	}
}

func newEchoContext(method, path string, body interface{}) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	e.Validator = apiValidator.NewCustomValidator()

	reqBody := []byte("")
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			panic(err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func assertJSONResponse(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int, containsSubstrings ...string) {
	t.Helper()
	require.Equal(t, expectedStatus, rec.Code)

	body := rec.Body.String()
	require.NotEmpty(t, body, "Response body should not be empty")

	var jsonBody map[string]interface{}
	err := json.Unmarshal([]byte(body), &jsonBody)
	require.NoError(t, err, "Response should be valid JSON")

	for _, substr := range containsSubstrings {
		require.True(t, strings.Contains(body, substr),
			"Response body should contain '%s', got: %s", substr, body)
	}
}

// TestHTTPHandler_PostRegister_Success tests successful registration via HTTP handler
func TestHTTPHandler_PostRegister_Success(t *testing.T) {
	t.Run("successfully registers user via HTTP", func(t *testing.T) {
		mockService := serviceMocks.NewMockService(t)
		mockService.EXPECT().Register(mock.Anything, mock.MatchedBy(func(req authservice.RegisterRequest) bool {
			return req.Username == "testuser" && req.Email == "test@example.com"
		})).Return(&model.User{
			ID:       1,
			Username: "testuser",
			Email:    "test@example.com",
		}, nil)

		handler := newTestHandler(mockService)

		req := &schemas.UserRegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "SecurePassword123!",
			FullName: "Test User",
		}
		c, rec := newEchoContext(http.MethodPost, "/api/register", req)

		err := handler.PostRegister(c)

		require.NoError(t, err)
		assertJSONResponse(t, rec, http.StatusCreated, "\"id\":1", "\"username\":\"testuser\"", "\"email\":\"test@example.com\"")
	})
}

// TestHTTPHandler_PostRegister_Conflict tests username conflict scenario
func TestHTTPHandler_PostRegister_Conflict(t *testing.T) {
	t.Run("returns conflict error when username exists", func(t *testing.T) {
		mockService := serviceMocks.NewMockService(t)
		mockService.EXPECT().Register(mock.Anything, mock.AnythingOfType("auth.RegisterRequest")).
			Return(nil, governerrors.NewCode(governerrors.CodeConflict, "username already exists"))

		handler := newTestHandler(mockService)

		req := &schemas.UserRegisterRequest{
			Username: "existinguser",
			Email:    "new@example.com",
			Password: "SecurePassword123!",
			FullName: "New User",
		}
		c, _ := newEchoContext(http.MethodPost, "/api/register", req)

		err := handler.PostRegister(c)

		require.Error(t, err)
		code, ok := governerrors.GetCode(err)
		assert.True(t, ok)
		assert.Equal(t, governerrors.CodeConflict, code)
	})
}

// TestHTTPHandler_PostLogin_Success tests successful login via HTTP handler
func TestHTTPHandler_PostLogin_Success(t *testing.T) {
	t.Run("successfully logs in user via HTTP", func(t *testing.T) {
		mockService := serviceMocks.NewMockService(t)
		mockService.EXPECT().Login(mock.Anything, mock.MatchedBy(func(req authservice.LoginRequest) bool {
			return req.Username == "testuser"
		})).Return(&authservice.LoginResponse{
			Token: "mock-jwt-token-12345",
			User: &model.User{
				ID:       1,
				Username: "testuser",
				Email:    "test@example.com",
			},
		}, nil)

		handler := newTestHandler(mockService)

		req := &schemas.LoginRequest{
			Username: "testuser",
			Password: "SecurePassword123!",
		}
		c, rec := newEchoContext(http.MethodPost, "/api/login", req)

		err := handler.PostLogin(c)

		require.NoError(t, err)
		assertJSONResponse(t, rec, http.StatusOK, "\"token\":")
	})
}

// TestHTTPHandler_PostLogin_Unauthorized tests invalid credentials
func TestHTTPHandler_PostLogin_Unauthorized(t *testing.T) {
	t.Run("returns unauthorized error for invalid credentials", func(t *testing.T) {
		mockService := serviceMocks.NewMockService(t)
		mockService.EXPECT().Login(mock.Anything, mock.AnythingOfType("auth.LoginRequest")).
			Return(nil, governerrors.ErrUnauthorized)

		handler := newTestHandler(mockService)

		req := &schemas.LoginRequest{
			Username: "testuser",
			Password: "wrongpassword",
		}
		c, _ := newEchoContext(http.MethodPost, "/api/login", req)

		err := handler.PostLogin(c)

		require.Error(t, err)
		code, ok := governerrors.GetCode(err)
		assert.True(t, ok)
		assert.Equal(t, governerrors.CodeUnauthorized, code)
	})
}
