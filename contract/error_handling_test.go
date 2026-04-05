package contract

import (
	"errors"
	"net/http"
	"testing"
)

// TestWrappedErrorWithContext tests the new wrapped error functionality
func TestWrappedErrorWithContext(t *testing.T) {
	// Test basic wrapping
	originalErr := errors.New("database connection failed")
	wrappedErr := WrapError(originalErr, "query_user", "database", map[string]any{
		"user_id": 123,
		"query":   "SELECT * FROM users WHERE id = ?",
	})

	if wrappedErr == nil {
		t.Fatal("expected wrapped error, got nil")
	}

	// Check that it implements error interface
	var err error = wrappedErr
	if err.Error() == "" {
		t.Fatal("wrapped error should have a message")
	}

	// Check unwrapping
	unwrapped := errors.Unwrap(wrappedErr)
	if unwrapped != originalErr {
		t.Fatalf("expected unwrapped error to be original, got %v", unwrapped)
	}

	// Check details extraction
	details := GetErrorDetails(wrappedErr)
	if details["operation"] != "query_user" {
		t.Fatalf("expected operation 'query_user', got %v", details["operation"])
	}
	if details["module"] != "database" {
		t.Fatalf("expected module 'database', got %v", details["module"])
	}
	if params, ok := details["params"].(map[string]any); !ok || params["user_id"] != 123 {
		t.Fatalf("expected params with user_id 123, got %v", details["params"])
	}
}

// TestWrapErrorf tests formatted error wrapping
func TestWrapErrorf(t *testing.T) {
	originalErr := errors.New("timeout")
	wrappedErr := WrapErrorf(originalErr, "operation failed after %d retries", 3)

	details := GetErrorDetails(wrappedErr)
	if details["message"] != "operation failed after 3 retries" {
		t.Fatalf("expected formatted message, got %v", details["message"])
	}
}

// TestPanicToError tests panic recovery
func TestPanicToError(t *testing.T) {
	// Test string panic
	err1 := PanicToError("something went wrong")
	if err1 == nil {
		t.Fatal("expected error from string panic")
	}

	// Test error panic
	originalErr := errors.New("original error")
	err2 := PanicToError(originalErr)
	if err2 == nil {
		t.Fatal("expected error from error panic")
	}

	// Test wrapped panic
	wrapped := WrapError(err2, "panic_test", "test", nil)
	if wrapped == nil {
		t.Fatal("expected wrapped error")
	}
}

// TestIsRetryable tests retryable error detection
func TestIsRetryable(t *testing.T) {
	// Test timeout error
	timeoutErr := NewErrorBuilder().
		Status(http.StatusRequestTimeout).
		Category(CategoryTimeout).
		Type(ErrTypeTimeout).
		Code(CodeTimeout).
		Message("timeout").
		Build()
	t.Logf("Timeout Error: %+v", timeoutErr)
	t.Logf("Status: %d, Category: %s", timeoutErr.Status, timeoutErr.Category)

	retryable := IsRetryable(timeoutErr)
	t.Logf("IsRetryable(timeoutErr): %v", retryable)
	if !retryable {
		t.Fatal("timeout error should be retryable")
	}

	// Test wrapped timeout error
	wrappedTimeout := WrapError(timeoutErr, "operation", "service", nil)
	t.Logf("Wrapped Timeout Error: %+v", wrappedTimeout)
	retryableWrapped := IsRetryable(wrappedTimeout)
	t.Logf("IsRetryable(wrappedTimeout): %v", retryableWrapped)
	if !retryableWrapped {
		t.Fatal("wrapped timeout error should be retryable")
	}

	// Test non-retryable error
	notFoundErr := NewErrorBuilder().
		Status(http.StatusNotFound).
		Category(CategoryClient).
		Type(ErrTypeNotFound).
		Code(CodeResourceNotFound).
		Message("resource 'resource' not found").
		Detail("resource", "resource").
		Build()
	if IsRetryable(notFoundErr) {
		t.Fatal("not found error should not be retryable")
	}

	// Test timeout-style network error
	netErr := &timeoutError{timeout: true}
	if !IsRetryable(netErr) {
		t.Fatal("timeout network error should be retryable")
	}
}

// TestGetErrorDetails tests error details extraction
func TestGetErrorDetails(t *testing.T) {
	// Test APIError
	apiErr := NewErrorBuilder().
		Status(http.StatusBadRequest).
		Category(CategoryValidation).
		Type(ErrTypeValidation).
		Code(CodeValidationError).
		Message("validation failed for field 'email': invalid format").
		Detail("field", "email").
		Detail("validation_message", "invalid format").
		Build()
	details := GetErrorDetails(apiErr)
	if details["status"] != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %v", details["status"])
	}

	// Test WrappedErrorWithContext
	wrappedErr := WrapError(errors.New("base error"), "test_op", "test_module", map[string]any{
		"key": "value",
	})
	details = GetErrorDetails(wrappedErr)
	if details["operation"] != "test_op" {
		t.Fatalf("expected operation test_op, got %v", details["operation"])
	}

	// Test standard error
	stdErr := errors.New("standard error")
	details = GetErrorDetails(stdErr)
	if details["message"] != "standard error" {
		t.Fatalf("expected standard error message, got %v", details["message"])
	}
}

// TestFormatError tests error formatting
func TestFormatError(t *testing.T) {
	err := WrapError(errors.New("base"), "op", "mod", map[string]any{"a": 1})
	formatted := FormatError(err)
	if formatted == "" {
		t.Fatal("expected non-empty formatted error")
	}
	// Should contain operation and module
	if !contains(formatted, "op=op") || !contains(formatted, "module=mod") {
		t.Fatalf("formatted error missing context: %s", formatted)
	}
}

// Helper types and functions

type timeoutError struct {
	timeout bool
}

func (e *timeoutError) Error() string {
	return "timeout error"
}

func (e *timeoutError) Timeout() bool {
	return e.timeout
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
