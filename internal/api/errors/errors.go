package errors

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"golang-sample/internal/api/schemas"
)

type APIError struct {
	Err         error               `json:"err"`
	ErrorDetail schemas.ErrorDetail `json:"error_detail"`
	HTTPCode    int                 `json:"http_code"`
}

func New(msg string, errorCode int, httpCode int) *APIError {
	if httpCode == 0 {
		httpCode = http.StatusBadRequest
	}
	return &APIError{
		Err:      errors.New(msg),
		HTTPCode: httpCode,
		ErrorDetail: schemas.ErrorDetail{
			Msg:       msg,
			ErrorCode: errorCode,
		},
	}
}

func NewBadRequest(msg string, errorCode int) *APIError {
	return &APIError{
		HTTPCode: http.StatusBadRequest,
		Err:      errors.New(msg),
		ErrorDetail: schemas.ErrorDetail{
			Msg:       msg,
			ErrorCode: errorCode,
		},
	}
}

func NewRequestBindingError(err error) error {
	errMsg := err.Error()
	if strings.Contains(errMsg, "Syntax error") {
		return Wrap(ErrValidationJSONFormatInvalid, ErrValidationJSONFormatInvalid, &schemas.ErrorDetail{})
	} else if strings.Contains(errMsg, "Unmarshal type error") {
		re := regexp.MustCompile(`expected=(.*?), got=(.*?), field=(.*?),`)
		matches := re.FindStringSubmatch(errMsg)

		if len(matches) < 3 {
			return Wrap(ErrValidationError, ErrValidationError, &schemas.ErrorDetail{})
		}

		expected := matches[1]
		got := matches[2]
		field := matches[3]
		return Wrap(ErrValidationJSONFieldTypeInvalid, ErrValidationJSONFieldTypeInvalid, &schemas.ErrorDetail{
			Property: field,
			MsgValues: map[string]interface{}{
				"got":      got,
				"expected": expected,
			},
		})
	}
	return Wrap(ErrValidationError, ErrValidationError, &schemas.ErrorDetail{})
}

func Wrap(err error, apiErr *APIError, info *schemas.ErrorDetail) error {
	errDetail := schemas.ErrorDetail{
		Msg:       apiErr.ErrorDetail.Msg,
		ErrorCode: apiErr.ErrorDetail.ErrorCode,
	}

	if info != nil {
		errDetail.Property = info.Property
		errDetail.MsgValues = info.MsgValues
	}

	return &APIError{
		HTTPCode:    apiErr.HTTPCode,
		Err:         err,
		ErrorDetail: errDetail,
	}
}

func (a *APIError) Error() string {
	return a.Err.Error()
}

func (a *APIError) HTTPError(path string) *echo.HTTPError {
	if a.HTTPCode == http.StatusInternalServerError || errors.Is(a.Err, ErrInternalServerError) {
		return &echo.HTTPError{
			Code: http.StatusInternalServerError,
			Message: schemas.ErrResponseBody{
				Timestamp: time.Now().UnixMilli(),
				Msg:       "INTERNAL_SERVER_ERROR",
				ErrorCode: http.StatusInternalServerError,
				Path:      path,
			},
			Internal: a.Err,
		}
	}

	errorDetail := schemas.ErrorDetail{
		Msg:       a.ErrorDetail.Msg,
		MsgValues: a.ErrorDetail.MsgValues,
		ErrorCode: a.ErrorDetail.ErrorCode,
		Property:  a.ErrorDetail.Property,
		Detail:    a.Err.Error(),
	}

	return &echo.HTTPError{
		Code: a.HTTPCode,
		Message: schemas.ErrResponseBody{
			Timestamp: time.Now().UnixMilli(),
			Msg:       a.Err.Error(),
			ErrorCode: a.HTTPCode,
			Errors: []*schemas.ErrorDetail{
				&errorDetail,
			},
			Path: path,
		},
		Internal: a.Err,
	}
}

var (
	ErrInternalServerError            = New("INTERNAL_SERVER_ERROR", 100000, http.StatusInternalServerError)
	ErrValidationError                = NewBadRequest("VALIDATION_ERROR", 100001)
	ErrValidationJSONFormatInvalid    = NewBadRequest("JSON_FORMAT_INVALID", 100002)
	ErrValidationJSONFieldTypeInvalid = NewBadRequest("FIELD_TYPE_INVALID", 100003)

	ErrUserAlreadyExist  = NewBadRequest("ErrUserAlreadyExist", 110000)
	ErrEmailAlreadyExist = NewBadRequest("ErrEmailAlreadyExist", 110000)
)
