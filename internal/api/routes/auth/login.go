package auth

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"

	"golang-sample/internal/api/errors"
	"golang-sample/internal/api/schemas"
	"golang-sample/pkg/config"
	"golang-sample/pkg/models"
	"golang-sample/pkg/utils/password"
)

// PostLogin godoc
//
//	@Summary	Login user
//	@Description
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		req	body		schemas.LoginRequest	true	"Login request"
//	@Success	200			{object}	schemas.Response
//	@Security	BearerAuth
//	@Router		/api/login [post]

func (a *Controller) PostLogin(c echo.Context) (err error) {
	var req schemas.LoginRequest
	if err = c.Bind(&req); err != nil {
		return errors.NewRequestBindingError(err)
	}

	if err = c.Validate(req); err != nil {
		return err
	}

	var user *models.User
	if user, err = a.storage.FindUserByUsername(c.Request().Context(), req.Username); err != nil {
		return errors.Wrap(err, errors.ErrInternalServerError, nil)
	}

	if user == nil {
		return echo.ErrUnauthorized
	}

	// Throws unauthorized error
	if password.CheckPasswordHash(user.PasswordHash, req.Password) {
		return echo.ErrUnauthorized
	}

	//if err = a.storage.UpdateLastLogin(c.Request().Context(), user, time.Now()); err != nil {
	//	a.log.Errorf("a.PostLogin: failed to update last login, err: %#v", err)
	//	return apiErr.Wrap(err, apiErr.ErrInternalServerError, nil)
	//}

	// Set custom claims
	claims := schemas.JwtClaims{
		ID:       cast.ToString(int(user.ID)),
		Email:    user.Email,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)), // TODO: Make it configurable
		},
	}

	// Create a token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate an encoded token and send it as a response.
	t, err := token.SignedString([]byte(config.ENV.API.Secret))
	if err != nil {
		a.log.Errorf("a.PostLogin: Oops, error while signing token: %v", err)
		return err
	}

	return c.JSON(
		http.StatusOK,
		schemas.NewResponse(schemas.JwtResponse{Token: t}, http.StatusOK),
	)
}
