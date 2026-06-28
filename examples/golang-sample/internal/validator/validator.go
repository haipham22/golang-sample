package validator

import (
	"reflect"
	"regexp"
	"strings"

	validatePkg "github.com/go-playground/validator/v10"

	apperrors "github.com/haipham22/golang-sample/internal/errors"
)

// Precompiled regex for array index notation
var arrayIndexRe = regexp.MustCompile(`\[\d+]`)

type CustomValidator struct {
	validator *validatePkg.Validate
}

func (cv *CustomValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		for _, fieldErr := range err.(validatePkg.ValidationErrors) {
			property := FormatStructField(fieldErr)
			return apperrors.Validation(property, "Validation failed for field: "+property)
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
