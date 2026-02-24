package auth

import (
	"golang-sample/internal/errors"
	schemas2 "golang-sample/internal/schemas"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"golang-sample/pkg/models"
	"golang-sample/pkg/utils/password"
)

// PostRegister godoc
//
//	@Summary	Login user
//	@Description
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		req	body		schemas.UserRegisterRequest	true	"Register request"
//	@Success	200			{object}	schemas.Response
//	@Security	BearerAuth
//	@Router		/api/register [post]

func (a *Controller) PostRegister(e echo.Context) error {
	req, err := a.validateUserRegisterRequest(e)
	if err != nil {
		return err
	}

	// Hash password
	hashedPassword, err := password.HashPassword(req.Password)
	if err != nil {
		a.log.Errorf("a.PostRegister: failed to hash password, err: %#v", err)
		return errors.Wrap(err, errors.ErrInternalServerError, nil)
	}

	// Create user
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	createdUser, err := a.storage.CreateUser(e.Request().Context(), user)
	if err != nil {
		a.log.Errorf("a.PostRegister: failed to create user, err: %#v", err)
		return errors.Wrap(err, errors.ErrInternalServerError, nil)
	}

	// Clear password hash from response
	createdUser.PasswordHash = ""

	return e.JSON(
		http.StatusCreated,
		schemas2.NewResponse(createdUser, http.StatusCreated),
	)
}

func (a *Controller) validateUserRegisterRequest(e echo.Context) (req *schemas2.UserRegisterRequest, err error) {
	if err = e.Bind(&req); err != nil {
		return nil, errors.NewRequestBindingError(err)
	}

	if err = e.Validate(req); err != nil {
		return req, err
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	req.Email = strings.TrimSpace(req.Email)
	req.FullName = strings.TrimSpace(req.FullName)

	if isUsernameOk, err := a.validateUniqueField("username", req.Username, errors.ErrUserAlreadyExist); !isUsernameOk && err != nil {
		return nil, err
	}
	if isEmailOk, err := a.validateUniqueField("email", req.Email, errors.ErrEmailAlreadyExist); !isEmailOk && err != nil {
		return nil, err
	}

	return
}

func (a *Controller) validateUniqueField(field, param string, errorCode error) (bool, error) {
	isExistBy, err := a.storage.IsExistBy(field, param)
	if err != nil {
		a.log.Errorf("a.validateUserRegisterRequest: Oops, error while checking if %s is taken: %v", field, err)
		return true, errors.ErrInternalServerError
	}

	if isExistBy {
		a.log.Warnf("a.validateUserRegisterRequest: Oops, %s=%s is already taken", field, param)
		return true, errorCode
	}
	return false, nil
}
