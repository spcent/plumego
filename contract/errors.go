package contract

import (
	"encoding/json"
	"fmt"
	"net/http"
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
