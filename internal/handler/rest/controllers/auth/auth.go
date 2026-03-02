package auth

import (
	"net/http"

	governerrors "github.com/haipham22/govern/errors"
	"github.com/labstack/echo/v4"

	schemas "golang-sample/internal/schemas"
	authservice "golang-sample/internal/service/auth"
)

// Controller handles HTTP requests for authentication endpoints.
type Controller struct {
	service authservice.Service
}

// New creates a new auth HTTP handler.
func New(service authservice.Service) *Controller {
	return &Controller{
		service: service,
	}
}

// PostRegister godoc
//
//	@Summary	Register user
//	@Description	Create a new user account
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		req	body		schemas.UserRegisterRequest	true	"Register request"
//	@Success	201			{object}	schemas.Response[schemas.User]
//	@Router		/api/register [post]
func (h *Controller) PostRegister(c echo.Context) error {
	var req schemas.UserRegisterRequest

	if err := c.Bind(&req); err != nil {
		return governerrors.WrapCode(governerrors.CodeInvalid, err)
	}

	// Call service - returns domain model
	modelUser, err := h.service.Register(c.Request().Context(), authservice.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
	})
	if err != nil {
		return err
	}

	// Convert model → schema
	schemaUser := modelToSchemaUser(modelUser)

	return c.JSON(
		http.StatusCreated,
		schemas.NewResponse(*schemaUser),
	)
}

// PostLogin godoc
//
//	@Summary	Login user
//	@Description	Authenticate user with username and password
//	@Tags	auth
//	@Accept		json
//	@Produce	json
//	@Param		req	body		schemas.LoginRequest	true	"Login request"
//	@Success	200			{object}	schemas.Response[schemas.JwtResponse]
//	@Router		/api/login [post]
func (h *Controller) PostLogin(c echo.Context) error {
	var req schemas.LoginRequest

	if err := c.Bind(&req); err != nil {
		return governerrors.WrapCode(governerrors.CodeInvalid, err)
	}

	// Call service - returns domain model
	modelResp, err := h.service.Login(c.Request().Context(), authservice.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return err
	}

	// Convert model → schema
	schemaResp := &schemas.LoginResponse{
		Token:     modelResp.Token,
		User:      modelToSchemaUser(modelResp.User),
		ExpiresAt: modelResp.ExpiresAt,
	}

	return c.JSON(http.StatusOK, schemas.NewResponse(schemaResp))
}
