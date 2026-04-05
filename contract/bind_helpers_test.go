package contract

import (
	"errors"
	"net/http"
	"testing"
)

func TestBindErrorToAPIErrorFields(t *testing.T) {
	type payload struct {
		Email string `validate:"required,email"`
	}

	err := validateStruct(&payload{})
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
	if apiErr.Type != ErrTypeValidation {
		t.Fatalf("expected validation error type, got %s", apiErr.Type)
	}
}

func TestBindErrorToAPIErrorType(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		wantType ErrorType
	}{
		{name: "body too large", err: ErrRequestBodyTooLarge, wantType: ErrTypeInvalidFormat},
		{name: "empty body", err: ErrEmptyRequestBody, wantType: ErrTypeInvalidFormat},
		{name: "invalid json", err: ErrInvalidJSON, wantType: ErrTypeInvalidFormat},
		{name: "unexpected extra data", err: ErrUnexpectedExtraData, wantType: ErrTypeInvalidFormat},
		{
			name:     "generic bind error fallback",
			err:      &BindError{Status: http.StatusBadRequest, Message: "failed to read request body", Err: errors.New("read failed")},
			wantType: ErrTypeInvalidFormat,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := BindErrorToAPIError(tc.err)
			if got.Type != tc.wantType {
				t.Fatalf("BindErrorToAPIError(%v).Type = %v, want %v", tc.err, got.Type, tc.wantType)
			}
		})
	}
}
