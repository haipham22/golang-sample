package middlewares

import (
	"golang-sample/internal/errors"
	"golang-sample/internal/schemas"
	"reflect"
	"regexp"
	"strings"

	validatePkg "github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	validator *validatePkg.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		for _, fieldErr := range err.(validatePkg.ValidationErrors) {
			return errors.Wrap(errors.ErrValidationError, errors.ErrValidationError, &schemas.ErrorDetail{
				Property: FormatStructField(fieldErr),
			})
		}
	}
	return nil
}

func NewCustomValidator() *CustomValidator {
	validate := validatePkg.New(validatePkg.WithRequiredStructEnabled())
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

func FormatStructField(fieldError validatePkg.FieldError) string {
	field := fieldError.Field()
	re := regexp.MustCompile(`\[\d+]`)

	// Replace array index notation with an empty string
	propertyWithoutIndex := re.ReplaceAllString(field, "")
	return propertyWithoutIndex
}
