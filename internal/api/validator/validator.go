package middlewares

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"

	"golang-sample/internal/api/errors"
	"golang-sample/internal/api/schemas"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		for _, fieldErr := range err.(validator.ValidationErrors) {
			return errors.Wrap(errors.ErrValidationError, errors.ErrValidationError, &schemas.ErrorDetail{
				Property: FormatStructField(fieldErr),
			})
		}
	}
	return nil
}

func NewCustomValidator() *CustomValidator {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})
	return &CustomValidator{
		validator: validate,
	}
}

func FormatStructField(fieldError validator.FieldError) string {
	field := fieldError.Field()
	re := regexp.MustCompile(`\[\d+]`)

	// Replace array index notation with an empty string
	propertyWithoutIndex := re.ReplaceAllString(field, "")
	return propertyWithoutIndex
}
