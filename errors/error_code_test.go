package errors_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/haipham22/govern/errors"
)

// Test helpers

func assertErrorCode(t *testing.T, err error, wantCode errors.ErrorCode) {
	t.Helper()
	code, ok := errors.GetCode(err)
	assert.True(t, ok, "GetCode() should return true")
	assert.Equal(t, wantCode, code, "GetCode() should return correct code")
}

func TestErrorWithCode(t *testing.T) {
	tests := []struct {
		name string
		code errors.ErrorCode
		err  error
		want string
	}{
		{
			name: "with error",
			code: errors.CodeNotFound,
			err:  io.EOF,
			want: "[NOT_FOUND] EOF",
		},
		{
			name: "with nil error",
			code: errors.CodeInternal,
			err:  nil,
			want: "INTERNAL",
		},
		{
			name: "with custom message",
			code: errors.CodeInvalid,
			err:  errors.New("invalid input"),
			want: "[INVALID] invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &errors.ErrorWithCode{Code: tt.code, Err: tt.err}
			got := e.Error()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestErrorWithCode_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")
	e := &errors.ErrorWithCode{Code: errors.CodeInternal, Err: baseErr}

	unwrapped := e.Unwrap()
	assert.Equal(t, baseErr, unwrapped, "Unwrap() should return the underlying error")
}

func TestNewCode(t *testing.T) {
	err := errors.NewCode(errors.CodeNotFound, "user not found")
	require.NotNil(t, err, "NewCode() should not return nil")

	assertErrorCode(t, err, errors.CodeNotFound)
}

func TestWrapCode(t *testing.T) {
	t.Run("wraps error successfully", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := errors.WrapCode(errors.CodeInternal, baseErr)

		require.NotNil(t, err, "WrapCode() should not return nil")
		assertErrorCode(t, err, errors.CodeInternal)
	})

	t.Run("wrapping nil returns nil", func(t *testing.T) {
		err := errors.WrapCode(errors.CodeInternal, nil)
		assert.Nil(t, err, "WrapCode(nil) should return nil")
	})
}

func TestGetCode(t *testing.T) {
	baseErr := errors.New("base error")
	wrappedErr := errors.WrapCode(errors.CodeNotFound, baseErr)

	tests := []struct {
		name     string
		err      error
		wantCode errors.ErrorCode
		wantOk   bool
	}{
		{
			name:     "with code",
			err:      wrappedErr,
			wantCode: errors.CodeNotFound,
			wantOk:   true,
		},
		{
			name:     "without code",
			err:      baseErr,
			wantCode: "",
			wantOk:   false,
		},
		{
			name:     "nil error",
			err:      nil,
			wantCode: "",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, ok := errors.GetCode(tt.err)
			assert.Equal(t, tt.wantOk, ok, "GetCode() ok status")
			assert.Equal(t, tt.wantCode, code, "GetCode() code value")
		})
	}
}

func TestIsCode(t *testing.T) {
	err := errors.WrapCode(errors.CodeNotFound, errors.New("not found"))

	tests := []struct {
		name string
		err  error
		code errors.ErrorCode
		want bool
	}{
		{
			name: "matching code",
			err:  err,
			code: errors.CodeNotFound,
			want: true,
		},
		{
			name: "different code",
			err:  err,
			code: errors.CodeInternal,
			want: false,
		},
		{
			name: "no code",
			err:  errors.New("plain error"),
			code: errors.CodeNotFound,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errors.IsCode(tt.err, tt.code)
			assert.Equal(t, tt.want, got, "IsCode() result")
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  *errors.ErrorWithCode
		code errors.ErrorCode
	}{
		{"ErrInternal", errors.ErrInternal, errors.CodeInternal},
		{"ErrInvalid", errors.ErrInvalid, errors.CodeInvalid},
		{"ErrNotFound", errors.ErrNotFound, errors.CodeNotFound},
		{"ErrUnauthorized", errors.ErrUnauthorized, errors.CodeUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.code, tt.err.Code, "%s.Code", tt.name)
			assert.True(t, errors.IsCode(tt.err, tt.code), "%s should be identified with IsCode", tt.name)
		})
	}
}
