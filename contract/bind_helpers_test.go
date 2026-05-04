package contract

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestBindErrorToAPIErrorFields(t *testing.T) {
	type payload struct {
		Email string `validate:"required,email"`
	}

	err := ValidateStruct(&payload{})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	bindErr := &bindError{Status: http.StatusBadRequest, Message: err.Error(), Err: err}

	apiErr := BindErrorToAPIError(bindErr)
	if apiErr.Code != CodeValidationError {
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
	if apiErr.Type != TypeValidation {
		t.Fatalf("expected validation error type, got %s", apiErr.Type)
	}
}

func TestBindErrorToAPIErrorType(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		wantType ErrorType
	}{
		{name: "body too large", err: ErrRequestBodyTooLarge},
		{name: "empty body", err: ErrEmptyRequestBody},
		{name: "invalid json", err: ErrInvalidJSON},
		{name: "unexpected extra data", err: ErrUnexpectedExtraData},
		{name: "invalid parameter", err: ErrInvalidParam},
		{name: "context nil", err: ErrContextNil},
		{name: "request nil", err: ErrRequestNil},
		{
			name: "generic bind error fallback",
			err:  &bindError{Status: http.StatusBadRequest, Message: "failed to read request body", Err: errors.New("read failed")},
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

func TestBindErrorToAPIErrorInfrastructureErrors(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		message string
	}{
		{name: "context nil", err: &bindError{Status: http.StatusBadRequest, Message: ErrContextNil.Error(), Err: ErrContextNil}, message: ErrContextNil.Error()},
		{name: "request nil", err: &bindError{Status: http.StatusBadRequest, Message: ErrRequestNil.Error(), Err: ErrRequestNil}, message: ErrRequestNil.Error()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := BindErrorToAPIError(tt.err)
			if apiErr.Code != CodeInvalidRequest {
				t.Fatalf("expected invalid request code, got %s", apiErr.Code)
			}
			if apiErr.Message != tt.message {
				t.Fatalf("expected message %q, got %q", tt.message, apiErr.Message)
			}
		})
	}
}

func TestBindErrorToAPIErrorInvalidParam(t *testing.T) {
	apiErr := BindErrorToAPIError(invalidBodySizeError())
	if apiErr.Code != CodeInvalidRequest {
		t.Fatalf("expected invalid request code, got %s", apiErr.Code)
	}
	if apiErr.Message != ErrInvalidParam.Error() {
		t.Fatalf("expected invalid parameter message, got %q", apiErr.Message)
	}
}

func TestBindErrorToAPIErrorValidationConfig(t *testing.T) {
	err := fmt.Errorf("%w: unknown validation rule", ErrValidationConfig)
	apiErr := BindErrorToAPIError(err)
	if apiErr.Status != http.StatusInternalServerError {
		t.Fatalf("status=%d, want %d", apiErr.Status, http.StatusInternalServerError)
	}
	if apiErr.Code != CodeInternalError {
		t.Fatalf("code=%s, want %s", apiErr.Code, CodeInternalError)
	}
	if apiErr.Category != CategoryServer {
		t.Fatalf("category=%s, want %s", apiErr.Category, CategoryServer)
	}
	if apiErr.Type != TypeInternal {
		t.Fatalf("type=%s, want %s", apiErr.Type, TypeInternal)
	}
	if apiErr.Message != ErrValidationConfig.Error() {
		t.Fatalf("message=%q, want %q", apiErr.Message, ErrValidationConfig.Error())
	}
}
