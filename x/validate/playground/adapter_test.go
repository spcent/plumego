package playground

import (
	"errors"
	"testing"

	validator "github.com/go-playground/validator/v10"
)

type payload struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
	Score int    `validate:"min=1"`
}

func TestValidatorAcceptsValidStruct(t *testing.T) {
	v := NewValidator()
	err := v.Validate(payload{Name: "Ada", Email: "ada@example.com", Score: 1})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidatorReturnsFieldErrors(t *testing.T) {
	v := NewValidator()
	err := v.Validate(payload{Email: "not-email"})
	if err == nil {
		t.Fatal("expected validation error")
	}

	var validationErr Error
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected playground Error, got %T", err)
	}
	if len(validationErr.Fields) != 3 {
		t.Fatalf("expected 3 field errors, got %d", len(validationErr.Fields))
	}

	tags := make(map[string]bool)
	for _, field := range validationErr.Fields {
		tags[field.Tag] = true
		if field.Message == "" {
			t.Fatalf("field error missing message: %#v", field)
		}
	}
	for _, tag := range []string{"required", "email", "min"} {
		if !tags[tag] {
			t.Fatalf("expected tag %q in %#v", tag, validationErr.Fields)
		}
	}
}

func TestValidatorUsesCustomTagName(t *testing.T) {
	type taggedPayload struct {
		Name string `binding:"required"`
	}

	v := NewValidator(WithTagName("binding"))
	err := v.Validate(taggedPayload{})
	if err == nil {
		t.Fatal("expected validation error")
	}

	var validationErr Error
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected playground Error, got %T", err)
	}
	if len(validationErr.Fields) != 1 || validationErr.Fields[0].Tag != "required" {
		t.Fatalf("unexpected field errors: %#v", validationErr.Fields)
	}
}

func TestValidatorUsesCustomValidation(t *testing.T) {
	type customPayload struct {
		Code string `validate:"plumecode"`
	}

	v := NewValidator(WithValidation("plumecode", func(fl validator.FieldLevel) bool {
		return fl.Field().String() == "plumego"
	}))
	if err := v.Validate(customPayload{Code: "plumego"}); err != nil {
		t.Fatalf("Validate() valid custom payload error = %v", err)
	}
	if err := v.Validate(customPayload{Code: "other"}); err == nil {
		t.Fatal("expected custom validation error")
	}
}
