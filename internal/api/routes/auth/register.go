package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"golang-sample/internal/api/errors"
	"golang-sample/internal/api/schemas"
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
	_, err := a.validateUserRegisterRequest(e)
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

func (a *Controller) validateUserRegisterRequest(e echo.Context) (req *schemas.UserRegisterRequest, err error) {
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
