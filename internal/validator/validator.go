package validator

import (
	"reflect"
	"regexp"
	"strings"

	validatePkg "github.com/go-playground/validator/v10"
	governerrors "github.com/haipham22/govern/errors"

	"golang-sample/internal/schemas"
)

// Precompiled regex for array index notation
var arrayIndexRe = regexp.MustCompile(`\[\d+]`)

type CustomValidator struct {
	validator *validatePkg.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		for _, fieldErr := range err.(validatePkg.ValidationErrors) {
			// Create detailed validation error with property information
			property := FormatStructField(fieldErr)
			detail := schemas.ErrorDetail{
				Property: property,
				Msg:      "Validation failed for field: " + property,
			}
			// Wrap the validation error with govern code and include detail in message
			return governerrors.WrapCode(governerrors.CodeInvalid,
				&ValidationError{Detail: detail})
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

	// Replace array index notation with an empty string
	return arrayIndexRe.ReplaceAllString(field, "")
}

// ValidationError represents a validation error with detailed information
type ValidationError struct {
	Detail schemas.ErrorDetail
}

func (e *ValidationError) Error() string {
	return e.Detail.Msg
}

// GetProperty returns the property that failed validation
func (e *ValidationError) GetProperty() string {
	return e.Detail.Property
}
