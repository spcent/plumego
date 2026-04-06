package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorBuilder(t *testing.T) {
	builder := NewErrorBuilder()

	err := builder.
		Status(http.StatusBadRequest).
		Category(CategoryValidation).
		Type(ErrTypeValidation).
		Code("TEST_ERROR").
		Message("test error message").
		Detail("field", "email").
		Detail("value", "invalid").
		Build()

	if err.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, err.Status)
	}

	if err.Category != CategoryValidation {
		t.Fatalf("expected category %s, got %s", CategoryValidation, err.Category)
	}

	if err.Code != "TEST_ERROR" {
		t.Fatalf("expected code %s, got %s", "TEST_ERROR", err.Code)
	}

	if err.Message != "test error message" {
		t.Fatalf("expected message %s, got %s", "test error message", err.Message)
	}

	if err.Details["field"] != "email" {
		t.Fatalf("expected field detail, got %v", err.Details["field"])
	}
}

func TestBuilderTypeOverwritesPriorFields(t *testing.T) {
	got := NewErrorBuilder().
		Status(999).
		Code("CUSTOM").
		Category(CategoryServer).
		Type(ErrTypeNotFound).
		Build()

	if got.Status != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, got.Status)
	}
	if got.Code != CodeResourceNotFound {
		t.Fatalf("expected code %q, got %q", CodeResourceNotFound, got.Code)
	}
	if got.Category != CategoryClient {
		t.Fatalf("expected category %q, got %q", CategoryClient, got.Category)
	}
}

func TestBuilderStatusAfterTypeWins(t *testing.T) {
	got := NewErrorBuilder().
		Type(ErrTypeNotFound).
		Status(http.StatusUnprocessableEntity).
		Build()

	if got.Status != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d", http.StatusUnprocessableEntity, got.Status)
	}
}

func TestErrorBuilderChaining(t *testing.T) {
	err := NewErrorBuilder().
		Status(http.StatusNotFound).
		Category(CategoryClient).
		Type(ErrTypeNotFound).
		Code("NOT_FOUND").
		Message("resource not found").
		Detail("resource", "user").
		Detail("id", "123").
		Build()

	if err.Status != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, err.Status)
	}

	if err.Details["resource"] != "user" || err.Details["id"] != "123" {
		t.Fatalf("expected details to be set, got %v", err.Details)
	}
}

func TestCommonErrorBuilders(t *testing.T) {
	// Test validation error via builder
	valErr := NewErrorBuilder().
		Status(http.StatusBadRequest).
		Category(CategoryValidation).
		Type(ErrTypeValidation).
		Code(CodeValidationError).
		Message("validation failed for field 'email': invalid format").
		Detail("field", "email").
		Detail("validation_message", "invalid format").
		Build()
	if valErr.Category != CategoryValidation {
		t.Fatalf("expected validation category")
	}
	if valErr.Details["field"] != "email" {
		t.Fatalf("expected field detail")
	}

	// Test not found error via builder
	notFoundErr := NewErrorBuilder().
		Status(http.StatusNotFound).
		Category(CategoryClient).
		Type(ErrTypeNotFound).
		Code(CodeResourceNotFound).
		Message("resource 'user' not found").
		Detail("resource", "user").
		Build()
	if notFoundErr.Category != CategoryClient {
		t.Fatalf("expected client category for not found")
	}
	if notFoundErr.Details["resource"] != "user" {
		t.Fatalf("expected resource detail")
	}

	// Test unauthorized error via builder
	authErr := NewErrorBuilder().
		Status(http.StatusUnauthorized).
		Category(CategoryAuth).
		Type(ErrTypeUnauthorized).
		Code(CodeUnauthorized).
		Message("invalid token").
		Build()
	if authErr.Category != CategoryAuth {
		t.Fatalf("expected authentication category")
	}

	// Test timeout error via builder
	timeoutErr := NewErrorBuilder().
		Status(http.StatusRequestTimeout).
		Category(CategoryTimeout).
		Type(ErrTypeTimeout).
		Code(CodeTimeout).
		Message("database timeout").
		Build()
	if timeoutErr.Category != CategoryTimeout {
		t.Fatalf("expected timeout category")
	}

	// Test rate limit error via builder
	rateLimitErr := NewErrorBuilder().
		Status(http.StatusTooManyRequests).
		Category(CategoryRateLimit).
		Type(ErrTypeRateLimited).
		Code(CodeRateLimited).
		Message("too many requests").
		Build()
	if rateLimitErr.Category != CategoryRateLimit {
		t.Fatalf("expected rate limit category")
	}
}

func TestErrorValidation(t *testing.T) {
	// Valid error
	validErr := APIError{
		Status:   http.StatusBadRequest,
		Code:     "VALIDATION_ERROR",
		Message:  "validation failed",
		Category: CategoryValidation,
	}

	validationErrors := validateAPIError(validErr)
	if len(validationErrors) > 0 {
		t.Fatalf("expected no validation errors, got %v", validationErrors)
	}

	// Invalid errors
	invalidCases := []APIError{
		{
			Status:   999, // Invalid status
			Code:     "ERROR",
			Message:  "message",
			Category: CategoryClient,
		},
		{
			Status:   http.StatusBadRequest,
			Code:     "", // Empty code
			Message:  "message",
			Category: CategoryClient,
		},
		{
			Status:   http.StatusBadRequest,
			Code:     "ERROR",
			Message:  "", // Empty message
			Category: CategoryClient,
		},
		{
			Status:   http.StatusBadRequest,
			Code:     "ERROR",
			Message:  "message",
			Category: "", // Empty category
		},
	}

	for i, invalidErr := range invalidCases {
		errs := validateAPIError(invalidErr)
		if len(errs) == 0 {
			t.Fatalf("case %d: expected validation errors", i)
		}
	}
}

func TestHTTPStatusFromCategory(t *testing.T) {
	tests := []struct {
		category ErrorCategory
		expected int
	}{
		{CategoryClient, http.StatusBadRequest},
		{CategoryValidation, http.StatusBadRequest},
		{CategoryAuth, http.StatusUnauthorized},
		{CategoryRateLimit, http.StatusTooManyRequests},
		{CategoryServer, http.StatusInternalServerError},
		{CategoryTimeout, http.StatusRequestTimeout},
		{CategoryBusiness, http.StatusUnprocessableEntity},
		{"unknown", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		if got := HTTPStatusFromCategory(tt.category); got != tt.expected {
			t.Fatalf("category %s: expected status %d, got %d", tt.category, tt.expected, got)
		}
	}
}

func TestErrorClassification(t *testing.T) {
	clientErr := APIError{Status: http.StatusBadRequest}
	serverErr := APIError{Status: http.StatusInternalServerError}

	if !IsClientError(clientErr) {
		t.Fatalf("expected client error to be classified as client error")
	}

	if IsClientError(serverErr) {
		t.Fatalf("expected server error to not be classified as client error")
	}

	if !IsServerError(serverErr) {
		t.Fatalf("expected server error to be classified as server error")
	}

	if IsServerError(clientErr) {
		t.Fatalf("expected client error to not be classified as server error")
	}
}

func TestRetryableErrorDetection(t *testing.T) {
	retryableStatuses := []int{408, 429, 500, 502, 503, 504}
	nonRetryableStatuses := []int{400, 401, 403, 404, 422}

	for _, status := range retryableStatuses {
		err := APIError{Status: status}
		if !IsAPIErrorRetryable(err) {
			t.Fatalf("expected status %d to be retryable", status)
		}
	}

	for _, status := range nonRetryableStatuses {
		err := APIError{Status: status}
		if IsAPIErrorRetryable(err) {
			t.Fatalf("expected status %d to not be retryable", status)
		}
	}

	// Test timeout error
	timeoutErr := APIError{Category: CategoryTimeout, Status: http.StatusOK}
	if !IsAPIErrorRetryable(timeoutErr) {
		t.Fatalf("expected timeout error to be retryable")
	}
}

func TestErrorResponseWriting(t *testing.T) {
	// Test WriteError function
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	err := APIError{
		Status:   http.StatusBadRequest,
		Code:     "VALIDATION_ERROR",
		Message:  "validation failed",
		Category: CategoryValidation,
		Details:  map[string]any{"field": "email"},
	}

	WriteError(recorder, req, err)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	if ct := recorder.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected content type application/json, got %s", ct)
	}

	var response ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected code in response")
	}
}

func TestErrorResponseWithTraceID(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Simulate trace ID in context
	ctx := WithTraceContext(req.Context(), TraceContext{
		TraceID: "test-trace-id",
		SpanID:  "test-span-id",
	})
	req = req.WithContext(ctx)

	err := APIError{
		Status:   http.StatusInternalServerError,
		Code:     "INTERNAL_ERROR",
		Message:  "internal server error",
		Category: CategoryServer,
	}

	WriteError(recorder, req, err)

	var response ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.TraceID != "test-trace-id" {
		t.Fatalf("expected trace ID in response")
	}
}

func TestWriteErrorPreservesTraceID(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := WithTraceContext(req.Context(), TraceContext{
		TraceID: "context-trace-id",
		SpanID:  "context-span-id",
	})
	req = req.WithContext(ctx)

	err := APIError{
		Status:   http.StatusBadRequest,
		Code:     "VALIDATION_ERROR",
		Message:  "validation failed",
		Category: CategoryValidation,
		TraceID:  "explicit-trace-id",
	}

	WriteError(recorder, req, err)

	var response ErrorResponse
	if decodeErr := json.NewDecoder(recorder.Body).Decode(&response); decodeErr != nil {
		t.Fatalf("failed to decode response: %v", decodeErr)
	}

	if response.TraceID != "explicit-trace-id" {
		t.Fatalf("expected explicit trace ID to be preserved")
	}
}

func TestParseErrorFromResponse(t *testing.T) {
	// Create a mock error response
	errorResp := ErrorResponse{
		Error: APIError{
			Status:   http.StatusNotFound,
			Code:     "RESOURCE_NOT_FOUND",
			Message:  "resource not found",
			Category: CategoryClient,
		},
	}

	body, _ := json.Marshal(errorResp)
	resp := httptest.NewRecorder()
	resp.WriteHeader(http.StatusNotFound)
	resp.Write(body)

	parsedErr, err := ParseErrorFromResponse(resp.Result())

	if err != nil {
		t.Fatalf("unexpected error parsing response: %v", err)
	}

	if parsedErr.Code != "RESOURCE_NOT_FOUND" {
		t.Fatalf("expected parsed error code to match")
	}

	if parsedErr.Category != CategoryClient {
		t.Fatalf("expected parsed error category to match")
	}
}

func TestParseErrorFromSuccessfulResponse(t *testing.T) {
	resp := httptest.NewRecorder()
	resp.WriteHeader(http.StatusOK)
	resp.WriteString("{}")

	_, err := ParseErrorFromResponse(resp.Result())

	if err == nil {
		t.Fatalf("expected error when parsing successful response")
	}
}

func TestParseErrorFromMalformedResponse(t *testing.T) {
	resp := httptest.NewRecorder()
	resp.WriteHeader(http.StatusInternalServerError)
	resp.WriteString("{invalid json}")

	parsedErr, err := ParseErrorFromResponse(resp.Result())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsedErr.Status != http.StatusInternalServerError {
		t.Fatalf("expected status to be preserved")
	}

	if parsedErr.Code != http.StatusText(http.StatusInternalServerError) {
		t.Fatalf("expected fallback code")
	}
}

func TestWriteErrorDefaults(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Test with minimal error (no status, code, or category)
	err := APIError{Message: "test message"}

	WriteError(recorder, req, err)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected default status to be internal server error")
	}

	var response ErrorResponse
	json.NewDecoder(recorder.Body).Decode(&response)

	if response.Error.Code != http.StatusText(http.StatusInternalServerError) {
		t.Fatalf("expected default code to be set")
	}

	if response.Error.Category != CategoryServer {
		t.Fatalf("expected default category to be server")
	}
}

func TestErrorBuilderWithSeverityAndType(t *testing.T) {
	err := NewErrorBuilder().
		Status(http.StatusBadRequest).
		Category(CategoryValidation).
		Type(ErrTypeValidation).
		Severity(SeverityWarning).
		Code("VALIDATION_WARNING").
		Message("validation warning").
		Build()

	if err.Severity != SeverityWarning {
		t.Fatalf("expected severity to be set on APIError")
	}

	if err.Type != ErrTypeValidation {
		t.Fatalf("expected type to be set on APIError")
	}
}

func TestErrorBuilderDetails(t *testing.T) {
	builder := NewErrorBuilder()
	details := map[string]any{
		"field":  "email",
		"value":  "invalid",
		"reason": "format",
	}

	err := builder.
		Status(http.StatusBadRequest).
		Category(CategoryValidation).
		Code("VALIDATION_ERROR").
		Message("validation failed").
		Details(details).
		Build()

	if len(err.Details) != 3 {
		t.Fatalf("expected 3 details, got %d", len(err.Details))
	}

	if err.Details["field"] != "email" || err.Details["value"] != "invalid" {
		t.Fatalf("expected details to be copied correctly")
	}
}

func TestWriteErrorZeroValueEmitsWarning(t *testing.T) {
	var warnings []string
	prev := WarnFunc
	WarnFunc = func(msg string) { warnings = append(warnings, msg) }
	defer func() { WarnFunc = prev }()

	w := httptest.NewRecorder()
	_ = WriteError(w, nil, APIError{})

	if len(warnings) == 0 {
		t.Fatal("expected WarnFunc to be called for zero-value APIError")
	}
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestWriteErrorEncodingFailureDoesNotCommitHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()

	err := WriteError(recorder, nil, NewErrorBuilder().
		Status(http.StatusBadRequest).
		Category(CategoryClient).
		Code("BAD_DETAIL").
		Message("bad detail").
		Detail("bad", make(chan int)).
		Build())
	if err == nil {
		t.Fatal("expected encoding error, got nil")
	}

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected headers to remain uncommitted, got status %d", recorder.Code)
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("expected no body on encoding failure, got %q", recorder.Body.String())
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "" {
		t.Fatalf("expected no content type on encoding failure, got %q", contentType)
	}
}

func TestErrorBuilderBuildFillsDefaults(t *testing.T) {
	// A builder with no explicit Status/Code/Category should produce a fully-
	// populated APIError after Build().
	got := NewErrorBuilder().Message("something went wrong").Build()

	if got.Status == 0 {
		t.Error("Build() must set a non-zero Status")
	}
	if got.Code == "" {
		t.Error("Build() must set a non-empty Code")
	}
	if got.Category == "" {
		t.Error("Build() must set a non-empty Category")
	}
	// A fully-populated APIError from the builder should not trigger WarnFunc.
	var warnCalled bool
	prev := WarnFunc
	WarnFunc = func(string) { warnCalled = true }
	defer func() { WarnFunc = prev }()

	w := httptest.NewRecorder()
	_ = WriteError(w, nil, got)
	if warnCalled {
		t.Error("WarnFunc must not be called for a fully-populated APIError from the builder")
	}
}

func TestErrorBuilderStatusOnlyDerivesCategory(t *testing.T) {
	got := NewErrorBuilder().
		Status(http.StatusBadRequest).
		Code("bad_request").
		Message("bad request").
		Build()

	if got.Category != CategoryClient {
		t.Fatalf("expected category %q, got %q", CategoryClient, got.Category)
	}
}
