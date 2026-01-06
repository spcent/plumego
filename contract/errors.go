package contract

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/spcent/plumego/log"
)

// ErrorCategory describes the high-level class of an API error for observability.
type ErrorCategory string

const (
	// CategoryClient covers 4xx errors caused by client input.
	CategoryClient ErrorCategory = "client_error"
	// CategoryServer covers 5xx errors caused by infrastructure or server logic.
	CategoryServer ErrorCategory = "server_error"
	// CategoryBusiness covers domain-specific validation or invariants.
	CategoryBusiness ErrorCategory = "business_error"
	// CategoryTimeout covers timeout errors.
	CategoryTimeout ErrorCategory = "timeout_error"
	// CategoryValidation covers input validation errors.
	CategoryValidation ErrorCategory = "validation_error"
	// CategoryAuthentication covers authentication and authorization errors.
	CategoryAuthentication ErrorCategory = "auth_error"
	// CategoryRateLimit covers rate limiting errors.
	CategoryRateLimit ErrorCategory = "rate_limit_error"
)

// ErrorSeverity describes the severity level of an error.
type ErrorSeverity string

const (
	SeverityInfo     ErrorSeverity = "info"
	SeverityWarning  ErrorSeverity = "warning"
	SeverityError    ErrorSeverity = "error"
	SeverityCritical ErrorSeverity = "critical"
)

// ErrorType represents specific error types for better categorization.
type ErrorType string

const (
	// Validation errors
	ErrTypeValidation    ErrorType = "validation_error"
	ErrTypeRequired      ErrorType = "required_field_missing"
	ErrTypeInvalidFormat ErrorType = "invalid_format"
	ErrTypeOutOfRange    ErrorType = "value_out_of_range"
	ErrTypeDuplicate     ErrorType = "duplicate_value"

	// Authentication/Authorization errors
	ErrTypeUnauthorized ErrorType = "unauthorized"
	ErrTypeForbidden    ErrorType = "forbidden"
	ErrTypeInvalidToken ErrorType = "invalid_token"
	ErrTypeExpiredToken ErrorType = "expired_token"

	// Resource errors
	ErrTypeNotFound      ErrorType = "resource_not_found"
	ErrTypeConflict      ErrorType = "resource_conflict"
	ErrTypeAlreadyExists ErrorType = "resource_already_exists"
	ErrTypeGone          ErrorType = "resource_gone"

	// System errors
	ErrTypeInternal    ErrorType = "internal_error"
	ErrTypeUnavailable ErrorType = "service_unavailable"
	ErrTypeTimeout     ErrorType = "timeout"
	ErrTypeRateLimited ErrorType = "rate_limited"
	ErrTypeMaintenance ErrorType = "maintenance_mode"

	// Business logic errors
	ErrTypeInvalidState        ErrorType = "invalid_state"
	ErrTypeInsufficientFunds   ErrorType = "insufficient_funds"
	ErrTypeOperationNotAllowed ErrorType = "operation_not_allowed"
)

// APIError represents a normalized error payload for HTTP responses and logging.
type APIError struct {
	Status   int            `json:"-"`
	Code     string         `json:"code"`
	Message  string         `json:"message"`
	Category ErrorCategory  `json:"category"`
	TraceID  string         `json:"trace_id,omitempty"`
	Details  map[string]any `json:"details,omitempty"`
}

// Error implements the error interface for APIError
func (e APIError) Error() string {
	return e.Message
}

// ErrorResponse wraps the error payload for consistent JSON responses.
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// WriteError writes a structured error response with trace context when available.
func WriteError(w http.ResponseWriter, r *http.Request, err APIError) {
	if err.Status == 0 {
		err.Status = http.StatusInternalServerError
	}

	if err.Code == "" {
		err.Code = http.StatusText(err.Status)
	}

	if err.Category == "" {
		if err.Status >= 500 {
			err.Category = CategoryServer
		} else {
			err.Category = CategoryClient
		}
	}

	if err.TraceID == "" && r != nil {
		if traceID := TraceIDFromContext(r.Context()); traceID != "" {
			err.TraceID = traceID
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)

	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: err})
}

// ErrorLogger converts an error into structured logging fields while preserving correlation ids.
func ErrorLogger(logger log.StructuredLogger, r *http.Request, err APIError) {
	if logger == nil {
		return
	}

	fields := log.Fields{
		"code":     err.Code,
		"status":   err.Status,
		"category": err.Category,
	}

	if err.TraceID == "" && r != nil {
		err.TraceID = TraceIDFromContext(r.Context())
	}

	if err.TraceID != "" {
		fields["trace_id"] = err.TraceID
	}

	for k, v := range err.Details {
		fields[k] = v
	}

	logger.WithFields(fields).Error(err.Message, nil)
}

// ErrorBuilder provides a fluent builder for creating APIError instances.
type ErrorBuilder struct {
	err APIError
}

// NewErrorBuilder creates a new error builder with default values.
func NewErrorBuilder() *ErrorBuilder {
	return &ErrorBuilder{
		err: APIError{
			Status:   http.StatusInternalServerError,
			Category: CategoryServer,
			Details:  make(map[string]any),
		},
	}
}

// Status sets the HTTP status code for the error.
func (b *ErrorBuilder) Status(status int) *ErrorBuilder {
	b.err.Status = status
	return b
}

// Code sets the error code for the error.
func (b *ErrorBuilder) Code(code string) *ErrorBuilder {
	b.err.Code = code
	return b
}

// Message sets the error message for the error.
func (b *ErrorBuilder) Message(message string) *ErrorBuilder {
	b.err.Message = message
	return b
}

// Category sets the error category for the error.
func (b *ErrorBuilder) Category(category ErrorCategory) *ErrorBuilder {
	b.err.Category = category
	return b
}

// Severity sets the severity level for the error by adding it to details.
func (b *ErrorBuilder) Severity(severity ErrorSeverity) *ErrorBuilder {
	b.err.Details["severity"] = severity
	return b
}

// Type sets the specific error type for the error.
func (b *ErrorBuilder) Type(errorType ErrorType) *ErrorBuilder {
	b.err.Details["type"] = errorType
	return b
}

// TraceID sets the trace ID for the error.
func (b *ErrorBuilder) TraceID(traceID string) *ErrorBuilder {
	b.err.TraceID = traceID
	return b
}

// Detail adds a detail field to the error.
func (b *ErrorBuilder) Detail(key string, value any) *ErrorBuilder {
	b.err.Details[key] = value
	return b
}

// Details sets multiple detail fields for the error.
func (b *ErrorBuilder) Details(details map[string]any) *ErrorBuilder {
	for k, v := range details {
		b.err.Details[k] = v
	}
	return b
}

// Build creates the final APIError instance.
func (b *ErrorBuilder) Build() APIError {
	return b.err
}

// ErrorChain represents a chain of errors for tracking error propagation.
type ErrorChain struct {
	root      error
	errors    []WrappedError
	timestamp time.Time
}

// WrappedError represents an error in the chain with context.
type WrappedError struct {
	Error     error
	Message   string
	Category  ErrorCategory
	Type      ErrorType
	Timestamp time.Time
	Context   map[string]any
}

// NewErrorChain creates a new error chain starting with the given error.
func NewErrorChain(err error) *ErrorChain {
	return &ErrorChain{
		root:      err,
		errors:    make([]WrappedError, 0),
		timestamp: time.Now(),
	}
}

// Add adds a new error to the chain with context.
func (ec *ErrorChain) Add(err error, message string, category ErrorCategory, errorType ErrorType) *ErrorChain {
	wrapped := WrappedError{
		Error:     err,
		Message:   message,
		Category:  category,
		Type:      errorType,
		Timestamp: time.Now(),
		Context:   make(map[string]any),
	}
	ec.errors = append(ec.errors, wrapped)
	return ec
}

// AddContext adds context information to the last error in the chain.
func (ec *ErrorChain) AddContext(key string, value any) *ErrorChain {
	if len(ec.errors) > 0 {
		ec.errors[len(ec.errors)-1].Context[key] = value
	}
	return ec
}

// Error returns the root error of the chain.
func (ec *ErrorChain) Error() string {
	if ec.root != nil {
		return ec.root.Error()
	}
	if len(ec.errors) > 0 {
		return ec.errors[len(ec.errors)-1].Message
	}
	return "unknown error"
}

// Root returns the root error of the chain.
func (ec *ErrorChain) Root() error {
	return ec.root
}

// Errors returns all errors in the chain.
func (ec *ErrorChain) Errors() []WrappedError {
	return ec.errors
}

// Latest returns the most recent error in the chain.
func (ec *ErrorChain) Latest() *WrappedError {
	if len(ec.errors) == 0 {
		return nil
	}
	return &ec.errors[len(ec.errors)-1]
}

// HasCategory checks if any error in the chain has the given category.
func (ec *ErrorChain) HasCategory(category ErrorCategory) bool {
	for _, err := range ec.errors {
		if err.Category == category {
			return true
		}
	}
	return false
}

// IsTimeoutError checks if the error chain contains timeout errors.
func (ec *ErrorChain) IsTimeoutError() bool {
	return ec.HasCategory(CategoryTimeout) ||
		ec.HasErrorType(ErrTypeTimeout)
}

// HasErrorType checks if any error in the chain has the given error type.
func (ec *ErrorChain) HasErrorType(errorType ErrorType) bool {
	for _, err := range ec.errors {
		if err.Type == errorType {
			return true
		}
	}
	return false
}

// ValidateError validates an APIError and returns validation errors if any.
func ValidateError(err APIError) []string {
	var validationErrors []string

	if err.Status < 100 || err.Status > 599 {
		validationErrors = append(validationErrors, "invalid HTTP status code")
	}

	if err.Code == "" {
		validationErrors = append(validationErrors, "error code is required")
	}

	if err.Message == "" {
		validationErrors = append(validationErrors, "error message is required")
	}

	if err.Category == "" {
		validationErrors = append(validationErrors, "error category is required")
	}

	return validationErrors
}

// Common error builders for frequently used error patterns.

func NewValidationError(field, message string) APIError {
	return NewErrorBuilder().
		Status(http.StatusBadRequest).
		Category(CategoryValidation).
		Type(ErrTypeValidation).
		Code("VALIDATION_ERROR").
		Message(fmt.Sprintf("validation failed for field '%s': %s", field, message)).
		Detail("field", field).
		Detail("validation_message", message).
		Build()
}

func NewNotFoundError(resource string) APIError {
	return NewErrorBuilder().
		Status(http.StatusNotFound).
		Category(CategoryClient).
		Type(ErrTypeNotFound).
		Code("RESOURCE_NOT_FOUND").
		Message(fmt.Sprintf("resource '%s' not found", resource)).
		Detail("resource", resource).
		Build()
}

func NewUnauthorizedError(message string) APIError {
	if message == "" {
		message = "authentication required"
	}
	return NewErrorBuilder().
		Status(http.StatusUnauthorized).
		Category(CategoryAuthentication).
		Type(ErrTypeUnauthorized).
		Code("UNAUTHORIZED").
		Message(message).
		Build()
}

func NewForbiddenError(message string) APIError {
	if message == "" {
		message = "access forbidden"
	}
	return NewErrorBuilder().
		Status(http.StatusForbidden).
		Category(CategoryAuthentication).
		Type(ErrTypeForbidden).
		Code("FORBIDDEN").
		Message(message).
		Build()
}

func NewTimeoutError(message string) APIError {
	if message == "" {
		message = "operation timed out"
	}
	return NewErrorBuilder().
		Status(http.StatusRequestTimeout).
		Category(CategoryTimeout).
		Type(ErrTypeTimeout).
		Code("TIMEOUT").
		Message(message).
		Build()
}

func NewInternalError(message string) APIError {
	if message == "" {
		message = "internal server error"
	}
	return NewErrorBuilder().
		Status(http.StatusInternalServerError).
		Category(CategoryServer).
		Type(ErrTypeInternal).
		Code("INTERNAL_ERROR").
		Message(message).
		Build()
}

func NewRateLimitError(message string) APIError {
	if message == "" {
		message = "rate limit exceeded"
	}
	return NewErrorBuilder().
		Status(http.StatusTooManyRequests).
		Category(CategoryRateLimit).
		Type(ErrTypeRateLimited).
		Code("RATE_LIMITED").
		Message(message).
		Build()
}

// HTTPStatusFromCategory returns the appropriate HTTP status code for an error category.
func HTTPStatusFromCategory(category ErrorCategory) int {
	switch category {
	case CategoryClient, CategoryValidation:
		return http.StatusBadRequest
	case CategoryAuthentication:
		return http.StatusUnauthorized
	case CategoryRateLimit:
		return http.StatusTooManyRequests
	case CategoryServer:
		return http.StatusInternalServerError
	case CategoryTimeout:
		return http.StatusRequestTimeout
	case CategoryBusiness:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

// ParseErrorFromResponse attempts to parse an APIError from an HTTP response.
func ParseErrorFromResponse(resp *http.Response) (APIError, error) {
	defer resp.Body.Close()

	if resp.StatusCode < 400 {
		return APIError{}, fmt.Errorf("no error in successful response")
	}

	var errorResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
		return APIError{
			Status:   resp.StatusCode,
			Code:     http.StatusText(resp.StatusCode),
			Message:  fmt.Sprintf("failed to parse error response: %v", err),
			Category: CategoryServer,
		}, nil
	}

	return errorResp.Error, nil
}

// IsClientError checks if the error is a client error (4xx).
func IsClientError(err APIError) bool {
	return err.Status >= 400 && err.Status < 500
}

// IsServerError checks if the error is a server error (5xx).
func IsServerError(err APIError) bool {
	return err.Status >= 500
}

// IsRetryableError checks if the error is retryable based on its status code.
func IsRetryableError(err APIError) bool {
	// Retryable status codes: 408, 429, 500, 502, 503, 504
	retryableCodes := []int{408, 429, 500, 502, 503, 504}

	for _, code := range retryableCodes {
		if err.Status == code {
			return true
		}
	}

	// Also retry if it's a timeout error
	return err.Category == CategoryTimeout
}

// ErrorMetrics represents metrics for error tracking and monitoring.
type ErrorMetrics struct {
	TotalErrors   int64                   `json:"total_errors"`
	ByCategory    map[ErrorCategory]int64 `json:"by_category"`
	ByType        map[ErrorType]int64     `json:"by_type"`
	ByStatus      map[int]int64           `json:"by_status"`
	LastErrorTime time.Time               `json:"last_error_time"`
}

// ErrorContext represents contextual information for error wrapping
type ErrorContext struct {
	Operation string         `json:"operation"`
	Module    string         `json:"module"`
	Params    map[string]any `json:"params,omitempty"`
}

// WrappedErrorWithContext represents an error with full context
type WrappedErrorWithContext struct {
	Err     error        `json:"error"`
	Context ErrorContext `json:"context"`
	Message string       `json:"message"`
	When    time.Time    `json:"when"`
}

// Error implements the error interface
func (w *WrappedErrorWithContext) Error() string {
	if w.Message != "" {
		return w.Message
	}
	if w.Err != nil {
		return w.Err.Error()
	}
	return "unknown error"
}

// Unwrap returns the underlying error for errors.Is/As support
func (w *WrappedErrorWithContext) Unwrap() error {
	return w.Err
}

// NewWrappedError creates a new wrapped error with context
func NewWrappedError(err error, operation, module string, params map[string]any) *WrappedErrorWithContext {
	return &WrappedErrorWithContext{
		Err: err,
		Context: ErrorContext{
			Operation: operation,
			Module:    module,
			Params:    params,
		},
		When: time.Now(),
	}
}

// WrapError wraps an existing error with additional context
func WrapError(err error, operation, module string, params map[string]any) error {
	if err == nil {
		return nil
	}

	// If already wrapped, add to the chain
	if wrapped, ok := err.(*WrappedErrorWithContext); ok {
		// Add operation to existing context if not set
		if wrapped.Context.Operation == "" && operation != "" {
			wrapped.Context.Operation = operation
		}
		if wrapped.Context.Module == "" && module != "" {
			wrapped.Context.Module = module
		}
		// Merge params
		if params != nil {
			if wrapped.Context.Params == nil {
				wrapped.Context.Params = make(map[string]any)
			}
			for k, v := range params {
				wrapped.Context.Params[k] = v
			}
		}
		return wrapped
	}

	return NewWrappedError(err, operation, module, params)
}

// WrapErrorf creates a wrapped error with a formatted message
func WrapErrorf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	message := fmt.Sprintf(format, args...)
	return &WrappedErrorWithContext{
		Err:     err,
		Message: message,
		When:    time.Now(),
	}
}

// IsRetryable checks if an error is retryable based on its type and context
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's an APIError directly
	if apiErr, ok := err.(APIError); ok {
		return IsRetryableError(apiErr)
	}

	// Check if it's a wrapped error with APIError
	if wrapped, ok := err.(*WrappedErrorWithContext); ok {
		// Check if wrapped error is an APIError by checking if it has Status field
		if apiErr, ok := wrapped.Err.(APIError); ok {
			return IsRetryableError(apiErr)
		}
		// Recursively check wrapped errors
		return IsRetryable(wrapped.Err)
	}

	// Network errors are typically retryable
	if netErr, ok := err.(interface{ Temporary() bool }); ok {
		return netErr.Temporary()
	}

	if netErr, ok := err.(interface{ Timeout() bool }); ok {
		return netErr.Timeout()
	}

	return false
}

// GetErrorDetails extracts detailed information from an error
func GetErrorDetails(err error) map[string]any {
	if err == nil {
		return nil
	}

	details := make(map[string]any)

	// Handle WrappedErrorWithContext
	if wrapped, ok := err.(*WrappedErrorWithContext); ok {
		details["message"] = wrapped.Error()
		details["when"] = wrapped.When
		if wrapped.Context.Operation != "" {
			details["operation"] = wrapped.Context.Operation
		}
		if wrapped.Context.Module != "" {
			details["module"] = wrapped.Context.Module
		}
		if len(wrapped.Context.Params) > 0 {
			details["params"] = wrapped.Context.Params
		}
		// Recursively get details from wrapped error
		if wrapped.Err != nil {
			details["wrapped_error"] = GetErrorDetails(wrapped.Err)
		}
		return details
	}

	// Handle APIError (which is a struct, not an error interface)
	// We check for the type by looking at the struct fields
	if apiErr, ok := err.(APIError); ok {
		details["status"] = apiErr.Status
		details["code"] = apiErr.Code
		details["category"] = apiErr.Category
		details["message"] = apiErr.Message
		if len(apiErr.Details) > 0 {
			details["details"] = apiErr.Details
		}
		if apiErr.TraceID != "" {
			details["trace_id"] = apiErr.TraceID
		}
		return details
	}

	// Basic error information
	details["message"] = err.Error()
	details["type"] = fmt.Sprintf("%T", err)

	return details
}

// FormatError formats an error for logging with full context
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	details := GetErrorDetails(err)

	// Simple string representation for logging
	var parts []string
	if msg, ok := details["message"].(string); ok {
		parts = append(parts, msg)
	}
	if op, ok := details["operation"].(string); ok {
		parts = append(parts, fmt.Sprintf("op=%s", op))
	}
	if mod, ok := details["module"].(string); ok {
		parts = append(parts, fmt.Sprintf("module=%s", mod))
	}
	if status, ok := details["status"].(int); ok {
		parts = append(parts, fmt.Sprintf("status=%d", status))
	}
	if code, ok := details["code"].(string); ok {
		parts = append(parts, fmt.Sprintf("code=%s", code))
	}

	return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}

// PanicToError converts a panic to an error (for recovery)
func PanicToError(r any) error {
	if r == nil {
		return nil
	}

	switch v := r.(type) {
	case error:
		return WrapError(v, "panic_recovery", "recovery", nil)
	case string:
		return fmt.Errorf("panic: %s", v)
	default:
		return fmt.Errorf("panic: %v", v)
	}
}

// Must wraps a function that returns an error and panics if it fails
// This is useful for initialization code where errors should be fatal
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// Must1 wraps a function that returns (T, error) and panics if it fails
func Must1[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// Must2 wraps a function that returns (T1, T2, error) and panics if it fails
func Must2[T1, T2 any](v1 T1, v2 T2, err error) (T1, T2) {
	if err != nil {
		panic(err)
	}
	return v1, v2
}

// NewErrorMetrics creates a new ErrorMetrics instance.
func NewErrorMetrics() *ErrorMetrics {
	return &ErrorMetrics{
		ByCategory: make(map[ErrorCategory]int64),
		ByType:     make(map[ErrorType]int64),
		ByStatus:   make(map[int]int64),
	}
}

// RecordError records an error in the metrics.
func (em *ErrorMetrics) RecordError(err APIError) {
	em.TotalErrors++
	em.ByCategory[err.Category]++
	em.ByStatus[err.Status]++

	if errorType, ok := err.Details["type"].(ErrorType); ok {
		em.ByType[errorType]++
	}

	em.LastErrorTime = time.Now()
}
