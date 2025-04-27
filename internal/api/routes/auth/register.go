package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"golang-sample/internal/api/errors"
	"golang-sample/internal/api/schemas"
)

type UserRegisterRequest struct {
	Username string `form:"username" json:"username" validate:"required"`
	Password string `form:"password" json:"password" validate:"required"`
	Email    string `form:"email" json:"email" validate:"required,email"`
	FullName string `form:"full_name" json:"full_name" validate:"required"`
}

func (a *authHandler) PostRegister(e echo.Context) error {
	req, err := a.validateUserRegisterRequest(e)
	if err != nil {
		return err
	}

	return e.JSON(
		http.StatusOK, schemas.Response{
			Data:       nil,
			Timestamp:  0,
			Pagination: nil,
			StatusCode: 0,
		},
	)
}

func (a *authHandler) validateUserRegisterRequest(e echo.Context) (req *UserRegisterRequest, err error) {
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

func (a *authHandler) validateUniqueField(field, param string, errorCode error) (bool, error) {
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
