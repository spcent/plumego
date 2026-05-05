package contract_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/spcent/plumego/contract"
)

func TestValidationErrorsExternalConsumption(t *testing.T) {
	type signup struct {
		Email string `validate:"required,email"`
		Name  string `validate:"required"`
	}

	err := contract.ValidateStruct(&signup{Email: "not-an-email"})
	if err == nil {
		t.Fatal("expected validation error")
	}

	var validationErr contract.ValidationErrors
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	if msg := validationErr.Error(); !strings.Contains(msg, "Email") || !strings.Contains(msg, "Name") {
		t.Fatalf("expected Error() to include field names, got %q", msg)
	}

	fields := validationErr.Errors()
	if len(fields) != 2 {
		t.Fatalf("expected two field errors, got %+v", fields)
	}
	fields[0].Field = "mutated"
	fields[0].Code = "MUTATED"

	again := validationErr.Errors()
	if again[0].Field == "mutated" || again[0].Code == "MUTATED" {
		t.Fatalf("expected Errors() to return a defensive copy, got %+v", again[0])
	}

	extracted := contract.FieldErrorsFrom(err)
	if len(extracted) != len(again) {
		t.Fatalf("expected FieldErrorsFrom to expose field errors, got %+v", extracted)
	}
	extracted[0].Field = "changed"
	if validationErr.Errors()[0].Field == "changed" {
		t.Fatal("expected FieldErrorsFrom result not to mutate ValidationErrors")
	}
}
