package playground

import (
	"errors"
	"strings"
	"testing"

	validator "github.com/go-playground/validator/v10"
	plumego "github.com/spcent/plumego/x/validate"
)

type signupRequest struct {
	Name  string `validate:"required,min=3,max=10"`
	Email string `validate:"required,email"`
	Code  string `validate:"startswith=PLU"`
}

func TestValidatorPassesForValidInput(t *testing.T) {
	var _ plumego.Validator = NewValidator()
	input := signupRequest{
		Name:  "plumego",
		Email: "dev@example.com",
		Code:  "PLU-123",
	}

	if err := NewValidator().Validate(input); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestValidatorRequiredFieldReportsFieldDetails(t *testing.T) {
	err := NewValidator().Validate(signupRequest{
		Email: "dev@example.com",
		Code:  "PLU-123",
	})

	got := requireValidationError(t, err)
	field := findField(t, got, "Name")
	if field.Tag != "required" {
		t.Fatalf("Name tag = %q, want required", field.Tag)
	}
	if field.Message != "Name is required" {
		t.Fatalf("Name message = %q, want required message", field.Message)
	}
	if field.Namespace == "" || field.StructNamespace == "" {
		t.Fatalf("expected namespace details, got %#v", field)
	}
}

func TestValidatorEmailFormatReportsDescriptiveMessage(t *testing.T) {
	err := NewValidator().Validate(signupRequest{
		Name:  "plumego",
		Email: "not-an-email",
		Code:  "PLU-123",
	})

	got := requireValidationError(t, err)
	field := findField(t, got, "Email")
	if field.Tag != "email" {
		t.Fatalf("Email tag = %q, want email", field.Tag)
	}
	if !strings.Contains(field.Message, "valid email") {
		t.Fatalf("Email message = %q, want descriptive email message", field.Message)
	}
}

func TestValidatorMinMaxViolationsReturnPerFieldErrors(t *testing.T) {
	err := NewValidator().Validate(signupRequest{
		Name:  "ab",
		Email: "dev@example.com",
		Code:  "PLU-123",
	})
	got := requireValidationError(t, err)
	minField := findField(t, got, "Name")
	if minField.Tag != "min" || minField.Param != "3" {
		t.Fatalf("Name min field = %#v, want tag min param 3", minField)
	}

	err = NewValidator().Validate(signupRequest{
		Name:  "this-is-too-long",
		Email: "dev@example.com",
		Code:  "PLU-123",
	})
	got = requireValidationError(t, err)
	maxField := findField(t, got, "Name")
	if maxField.Tag != "max" || maxField.Param != "10" {
		t.Fatalf("Name max field = %#v, want tag max param 10", maxField)
	}
}

func TestNewValidatorRegistersCustomValidationFunction(t *testing.T) {
	v := NewValidator(WithValidation("ticket", func(fl validator.FieldLevel) bool {
		value, ok := fl.Field().Interface().(string)
		return ok && strings.HasPrefix(value, "TICKET-")
	}))

	type request struct {
		ID string `validate:"ticket"`
	}
	if err := v.Validate(request{ID: "TICKET-123"}); err != nil {
		t.Fatalf("Validate with custom function returned error: %v", err)
	}

	err := v.Validate(request{ID: "BAD-123"})
	got := requireValidationError(t, err)
	field := findField(t, got, "ID")
	if field.Tag != "ticket" {
		t.Fatalf("ID tag = %q, want ticket", field.Tag)
	}
}

func requireValidationError(t *testing.T, err error) Error {
	t.Helper()
	if err == nil {
		t.Fatal("Validate returned nil error")
	}
	var validationErr Error
	if !errors.As(err, &validationErr) {
		t.Fatalf("Validate error = %T, want playground.Error", err)
	}
	if len(validationErr.Fields) == 0 {
		t.Fatal("playground.Error has no fields")
	}
	return validationErr
}

func findField(t *testing.T, err Error, name string) FieldError {
	t.Helper()
	for _, field := range err.Fields {
		if field.Field == name {
			return field
		}
	}
	t.Fatalf("field %q not found in %#v", name, err.Fields)
	return FieldError{}
}
