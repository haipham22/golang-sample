package errors_test

import (
	"fmt"
	"testing"

	governerrors "github.com/haipham22/govern/errors"
)

type customErr struct{ msg string }

func (e *customErr) Error() string { return e.msg }

func TestNew(t *testing.T) {
	err := governerrors.New("test error")
	if err == nil {
		t.Fatal("New() returned nil")
	}
	if err.Error() != "test error" {
		t.Errorf("New() error = %v, want %v", err.Error(), "test error")
	}
}

func TestErrorf(t *testing.T) {
	err := governerrors.Errorf("test %d", 42)
	if err == nil {
		t.Fatal("Errorf() returned nil")
	}
	if err.Error() != "test 42" {
		t.Errorf("Errorf() error = %v, want %v", err.Error(), "test 42")
	}
}

func TestIs(t *testing.T) {
	baseErr := governerrors.New("base error")
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)

	if !governerrors.Is(wrappedErr, baseErr) {
		t.Error("Is() should return true for wrapped error")
	}

	otherErr := governerrors.New("other error")
	if governerrors.Is(wrappedErr, otherErr) {
		t.Error("Is() should return false for different error")
	}
}

func TestAs(t *testing.T) {
	ce := &customErr{msg: "custom"}
	err := fmt.Errorf("wrapped: %w", ce)

	var target *customErr
	if !governerrors.As(err, &target) {
		t.Error("As() should find the custom error")
	}
	if target.msg != "custom" {
		t.Errorf("As() target = %v, want %v", target.msg, "custom")
	}
}

func TestUnwrap(t *testing.T) {
	baseErr := governerrors.New("base")
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)

	if unwrapped := governerrors.Unwrap(wrappedErr); unwrapped != baseErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, baseErr)
	}
}

func TestJoin(t *testing.T) {
	err1 := governerrors.New("error 1")
	err2 := governerrors.New("error 2")
	err3 := governerrors.New("error 3")

	joined := governerrors.Join(err1, err2, err3)

	if !governerrors.Is(joined, err1) {
		t.Error("Join() should contain err1")
	}
	if !governerrors.Is(joined, err2) {
		t.Error("Join() should contain err2")
	}
	if !governerrors.Is(joined, err3) {
		t.Error("Join() should contain err3")
	}
}
