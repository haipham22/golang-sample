package validator

import (
	"testing"

	validatePkg "github.com/go-playground/validator/v10"

	apperrors "github.com/haipham22/golang-sample/internal/errors"
)

type validatorTestRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required"`
}

type indexedRequest struct {
	Items []indexedItem `validate:"dive"`
}

type indexedItem struct {
	Name string `validate:"required"`
}

func TestCustomValidator_Validate(t *testing.T) {
	v := NewCustomValidator()

	t.Run("valid request", func(t *testing.T) {
		if err := v.Validate(validatorTestRequest{Email: "user@example.com", Name: "Alice"}); err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("invalid request returns typed validation error", func(t *testing.T) {
		err := v.Validate(validatorTestRequest{Name: "Alice"})
		if !apperrors.IsCode(err, apperrors.CodeInvalid) {
			t.Fatalf("Validate() error = %v, want CodeInvalid", err)
		}

		_, body := apperrors.Resolve(err, "/api/test", "")
		if len(body.Errors) != 1 || body.Errors[0].Property != "email" {
			t.Fatalf("body.Errors = %+v, want email field error", body.Errors)
		}
	})
}

func TestFormatStructField_StripsArrayIndexes(t *testing.T) {
	validate := validatePkg.New()
	err := validate.Struct(indexedRequest{Items: []indexedItem{{}}})
	if err == nil {
		t.Fatal("Struct() error = nil, want validation error")
	}

	fieldErrors := err.(validatePkg.ValidationErrors)
	if got := FormatStructField(fieldErrors[0]); got != "Name" {
		t.Fatalf("FormatStructField() = %q, want %q", got, "Name")
	}
}
