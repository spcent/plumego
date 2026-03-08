package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
)

func TestCreateUserHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBufferString("{"))
	rec := httptest.NewRecorder()

	createUserHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var resp contract.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error.Code != contract.CodeInvalidJSON {
		t.Fatalf("expected error code %s, got %s", contract.CodeInvalidJSON, resp.Error.Code)
	}
}

func TestCreateUserHandler_UnknownFields(t *testing.T) {
	body := `{"email":"user@example.com","password":"secret","extra":"x"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	createUserHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var resp contract.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error.Code != contract.CodeInvalidJSON {
		t.Fatalf("expected unknown fields to map to %s, got %s", contract.CodeInvalidJSON, resp.Error.Code)
	}
}

func TestCreateUserHandler_ValidationErrorCode(t *testing.T) {
	body := `{"email":"not-an-email","password":""}`
	req := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	createUserHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var resp contract.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error.Code != contract.CodeValidationError {
		t.Fatalf("expected validation code %s, got %s", contract.CodeValidationError, resp.Error.Code)
	}
}
