package contract

import (
	"net/http"
	"testing"

	"github.com/spcent/plumego/validator"
)

func TestBindErrorToAPIErrorFields(t *testing.T) {
	type payload struct {
		Email string `validate:"required,email"`
	}

	err := validator.Validate(&payload{})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	bindErr := &BindError{Status: http.StatusBadRequest, Message: err.Error(), Err: err}

	apiErr := BindErrorToAPIError(bindErr)
	if apiErr.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected validation error code, got %s", apiErr.Code)
	}
	raw, ok := apiErr.Details["fields"]
	if !ok {
		t.Fatalf("expected fields detail")
	}
	fields, ok := raw.([]FieldError)
	if !ok || len(fields) == 0 {
		t.Fatalf("expected field errors, got %T", raw)
	}
}
