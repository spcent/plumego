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
		{name: "invalid parameter", err: ErrInvalidParam, wantType: TypeInternal},
		{name: "context nil", err: ErrContextNil, wantType: TypeInternal},
		{name: "request nil", err: ErrRequestNil, wantType: TypeInternal},
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
		{name: "invalid bind destination", err: &bindError{Status: http.StatusBadRequest, Message: ErrInvalidBindDst.Error(), Err: ErrInvalidBindDst}, message: "invalid bind destination"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := BindErrorToAPIError(tt.err)
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
			if apiErr.Message != tt.message {
				t.Fatalf("expected message %q, got %q", tt.message, apiErr.Message)
			}
		})
	}
}

func TestBindErrorToAPIErrorInvalidParam(t *testing.T) {
	apiErr := BindErrorToAPIError(invalidBodySizeError())
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
	if apiErr.Message != ErrInvalidParam.Error() {
		t.Fatalf("expected invalid parameter message, got %q", apiErr.Message)
	}
}

func TestBindErrorToAPIErrorClientInputClassification(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{name: "body too large", err: ErrRequestBodyTooLarge, status: http.StatusRequestEntityTooLarge, code: CodeRequestBodyTooLarge},
		{name: "empty body", err: ErrEmptyRequestBody, status: http.StatusBadRequest, code: CodeEmptyBody},
		{name: "invalid json", err: ErrInvalidJSON, status: http.StatusBadRequest, code: CodeInvalidJSON},
		{name: "extra data", err: ErrUnexpectedExtraData, status: http.StatusBadRequest, code: CodeUnexpectedExtraData},
		{name: "invalid query", err: fmt.Errorf("%w: parse int", ErrInvalidQueryParam), status: http.StatusBadRequest, code: CodeInvalidQuery},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiErr := BindErrorToAPIError(tt.err)
			if apiErr.Status != tt.status {
				t.Fatalf("status=%d, want %d", apiErr.Status, tt.status)
			}
			if apiErr.Category != CategoryValidation {
				t.Fatalf("category=%s, want %s", apiErr.Category, CategoryValidation)
			}
			if apiErr.Code != tt.code {
				t.Fatalf("code=%s, want %s", apiErr.Code, tt.code)
			}
		})
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

func TestBindErrorToAPIErrorValidationConfigPrecedenceOverFields(t *testing.T) {
	err := errors.Join(
		fmt.Errorf("%w: unknown validation rule", ErrValidationConfig),
		ValidationErrors{errors: []FieldError{{
			Field:   "email",
			Code:    CodeRequired,
			Message: "email is required",
		}}},
	)

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
	if _, ok := apiErr.Details["fields"]; ok {
		t.Fatalf("expected validation config errors not to expose field details, got %+v", apiErr.Details)
	}
}
