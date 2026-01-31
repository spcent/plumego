package contract

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/spcent/plumego/log"
)

// mockStructuredLogger implements StructuredLogger for testing
type mockStructuredLogger struct {
	onError func(msg string, fields log.Fields)
	fields  log.Fields
}

func (m *mockStructuredLogger) WithFields(fields log.Fields) log.StructuredLogger {
	// Create a new logger with the combined fields
	newLogger := &mockStructuredLogger{
		onError: m.onError,
		fields:  make(log.Fields),
	}
	// Copy existing fields
	for k, v := range m.fields {
		newLogger.fields[k] = v
	}
	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	return newLogger
}

func (m *mockStructuredLogger) Debug(msg string, fields log.Fields) {
	// Not used in this test
}

func (m *mockStructuredLogger) Info(msg string, fields log.Fields) {
	// Not used in this test
}

func (m *mockStructuredLogger) Warn(msg string, fields log.Fields) {
	// Not used in this test
}

func (m *mockStructuredLogger) Error(msg string, fields log.Fields) {
	if m.onError != nil {
		// Use the logger's fields (from WithFields) and merge with any additional fields
		combinedFields := make(log.Fields)
		for k, v := range m.fields {
			combinedFields[k] = v
		}
		for k, v := range fields {
			combinedFields[k] = v
		}
		m.onError(msg, combinedFields)
	}
}

func (m *mockStructuredLogger) DebugCtx(ctx context.Context, msg string, fields log.Fields) {
	// Not used in this test
}

func (m *mockStructuredLogger) InfoCtx(ctx context.Context, msg string, fields log.Fields) {
	// Not used in this test
}

func (m *mockStructuredLogger) WarnCtx(ctx context.Context, msg string, fields log.Fields) {
	// Not used in this test
}

func (m *mockStructuredLogger) ErrorCtx(ctx context.Context, msg string, fields log.Fields) {
	// Not used in this test
}

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

func TestErrorChain(t *testing.T) {
	rootErr := errors.New("root error")
	chain := NewErrorChain(rootErr)

	// Add errors to chain
	chain.Add(errors.New("validation error"), "validation failed", CategoryValidation, ErrTypeValidation)
	chain.Add(errors.New("database error"), "database operation failed", CategoryServer, ErrTypeInternal)

	if chain.Root() != rootErr {
		t.Fatalf("expected root error to be preserved")
	}

	if len(chain.Errors()) != 2 {
		t.Fatalf("expected 2 errors in chain, got %d", len(chain.Errors()))
	}

	latest := chain.Latest()
	if latest == nil || latest.Category != CategoryServer {
		t.Fatalf("expected latest error to be server error")
	}

	if !chain.HasCategory(CategoryValidation) {
		t.Fatalf("expected chain to have validation category")
	}

	if !chain.HasErrorType(ErrTypeInternal) {
		t.Fatalf("expected chain to have internal error type")
	}
}

func TestErrorChainContext(t *testing.T) {
	chain := NewErrorChain(errors.New("test error"))

	chain.Add(errors.New("validation error"), "validation failed", CategoryValidation, ErrTypeValidation)
	chain.AddContext("field", "email")
	chain.AddContext("value", "invalid")

	latest := chain.Latest()
	if latest == nil {
		t.Fatalf("expected latest error to exist")
	}

	if latest.Context["field"] != "email" {
		t.Fatalf("expected field context to be set")
	}

	if latest.Context["value"] != "invalid" {
		t.Fatalf("expected value context to be set")
	}
}

func TestErrorChainTimeoutDetection(t *testing.T) {
	chain := NewErrorChain(errors.New("timeout error"))

	// Add timeout error
	chain.Add(errors.New("operation timeout"), "operation timed out", CategoryTimeout, ErrTypeTimeout)

	if !chain.IsTimeoutError() {
		t.Fatalf("expected chain to detect timeout error")
	}
}

func TestCommonErrorConstructors(t *testing.T) {
	// Test validation error
	valErr := NewValidationError("email", "invalid format")
	if valErr.Category != CategoryValidation {
		t.Fatalf("expected validation category")
	}
	if valErr.Details["field"] != "email" {
		t.Fatalf("expected field detail")
	}

	// Test not found error
	notFoundErr := NewNotFoundError("user")
	if notFoundErr.Category != CategoryClient {
		t.Fatalf("expected client category for not found")
	}
	if notFoundErr.Details["resource"] != "user" {
		t.Fatalf("expected resource detail")
	}

	// Test unauthorized error
	authErr := NewUnauthorizedError("invalid token")
	if authErr.Category != CategoryAuthentication {
		t.Fatalf("expected authentication category")
	}

	// Test timeout error
	timeoutErr := NewTimeoutError("database timeout")
	if timeoutErr.Category != CategoryTimeout {
		t.Fatalf("expected timeout category")
	}

	// Test rate limit error
	rateLimitErr := NewRateLimitError("too many requests")
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

	validationErrors := ValidateError(validErr)
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
		errs := ValidateError(invalidErr)
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
		{CategoryAuthentication, http.StatusUnauthorized},
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
		if !IsRetryableError(err) {
			t.Fatalf("expected status %d to be retryable", status)
		}
	}

	for _, status := range nonRetryableStatuses {
		err := APIError{Status: status}
		if IsRetryableError(err) {
			t.Fatalf("expected status %d to not be retryable", status)
		}
	}

	// Test timeout error
	timeoutErr := APIError{Category: CategoryTimeout, Status: http.StatusOK}
	if !IsRetryableError(timeoutErr) {
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
	ctx := ContextWithTraceContext(req.Context(), TraceContext{
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

	if response.Error.TraceID != "test-trace-id" {
		t.Fatalf("expected trace ID in response")
	}
}

func TestWriteErrorPreservesTraceID(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := ContextWithTraceContext(req.Context(), TraceContext{
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

	if response.Error.TraceID != "explicit-trace-id" {
		t.Fatalf("expected explicit trace ID to be preserved")
	}
}

func TestErrorLogging(t *testing.T) {
	// Create a mock logger that captures log entries
	var loggedFields map[string]any
	var loggedMessage string

	mockLogger := &mockStructuredLogger{
		onError: func(msg string, fields log.Fields) {
			loggedFields = make(map[string]any)
			for k, v := range fields {
				loggedFields[k] = v
			}
			loggedMessage = msg
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	err := APIError{
		Status:   http.StatusInternalServerError,
		Code:     "INTERNAL_ERROR",
		Message:  "internal server error",
		Category: CategoryServer,
		Details:  map[string]any{"service": "database"},
	}

	ErrorLogger(mockLogger, req, err)

	if loggedMessage != "internal server error" {
		t.Fatalf("expected logged message to match")
	}

	if loggedFields["code"] != "INTERNAL_ERROR" {
		t.Fatalf("expected code in logged fields")
	}

	if loggedFields["service"] != "database" {
		t.Fatalf("expected service detail in logged fields")
	}
}

func TestErrorMetrics(t *testing.T) {
	metrics := NewErrorMetrics()

	// Record various errors
	errors := []APIError{
		{Status: http.StatusBadRequest, Category: CategoryValidation, Details: map[string]any{"type": ErrTypeValidation}},
		{Status: http.StatusNotFound, Category: CategoryClient, Details: map[string]any{"type": ErrTypeNotFound}},
		{Status: http.StatusInternalServerError, Category: CategoryServer, Details: map[string]any{"type": ErrTypeInternal}},
		{Status: http.StatusBadRequest, Category: CategoryValidation, Details: map[string]any{"type": ErrTypeValidation}},
	}

	for _, err := range errors {
		metrics.RecordError(err)
	}

	if metrics.TotalErrors != 4 {
		t.Fatalf("expected 4 total errors, got %d", metrics.TotalErrors)
	}

	if metrics.ByCategory[CategoryValidation] != 2 {
		t.Fatalf("expected 2 validation errors, got %d", metrics.ByCategory[CategoryValidation])
	}

	if metrics.ByStatus[http.StatusBadRequest] != 2 {
		t.Fatalf("expected 2 bad request errors, got %d", metrics.ByStatus[http.StatusBadRequest])
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

	if err.Details["severity"] != SeverityWarning {
		t.Fatalf("expected severity to be set in details")
	}

	if err.Details["type"] != ErrTypeValidation {
		t.Fatalf("expected type to be set in details")
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

func TestErrorChainWithEmptyRoot(t *testing.T) {
	chain := NewErrorChain(nil)

	chain.Add(errors.New("first error"), "first error occurred", CategoryServer, ErrTypeInternal)

	if chain.Error() != "first error occurred" {
		t.Fatalf("expected chain error to be the message of the latest error")
	}

	if len(chain.Errors()) != 1 {
		t.Fatalf("expected 1 error in chain")
	}
}

func TestErrorChainNoErrors(t *testing.T) {
	chain := NewErrorChain(errors.New("root"))

	if chain.Error() != "root" {
		t.Fatalf("expected chain error to be root error")
	}

	if chain.Latest() != nil {
		t.Fatalf("expected latest to be nil when no errors added")
	}

	if chain.HasCategory(CategoryValidation) {
		t.Fatalf("expected chain to not have validation category")
	}
}
