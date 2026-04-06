package contract

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// WarnFunc is invoked when WriteError receives an APIError with missing required
// fields (Status, Code, or Category). It defaults to a no-op; override in tests
// or at application startup to surface misconfigured callers.
var WarnFunc = func(msg string) {}

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
	// CategoryAuth covers authentication and authorization errors.
	CategoryAuth ErrorCategory = "auth_error"
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
	TypeValidation    ErrorType = "validation_error"
	TypeRequired      ErrorType = "required_field_missing"
	TypeInvalidFormat ErrorType = "invalid_format"
	TypeOutOfRange    ErrorType = "value_out_of_range"
	TypeDuplicate     ErrorType = "duplicate_value"

	// Authentication/Authorization errors
	TypeUnauthorized ErrorType = "unauthorized_request"
	TypeForbidden    ErrorType = "forbidden_request"
	TypeInvalidToken ErrorType = "invalid_token"
	TypeExpiredToken ErrorType = "expired_token"

	// Resource errors
	TypeNotFound      ErrorType = "resource_not_found"
	TypeConflict      ErrorType = "resource_conflict"
	TypeAlreadyExists ErrorType = "resource_already_exists"
	TypeGone          ErrorType = "resource_gone"

	// System errors
	TypeInternal    ErrorType = "internal_error"
	TypeUnavailable ErrorType = "service_unavailable"
	TypeTimeout     ErrorType = "timeout_error"
	TypeRateLimited ErrorType = "rate_limited"
	TypeMaintenance ErrorType = "maintenance_mode"

	// Business logic errors
	TypeInvalidState        ErrorType = "invalid_state"
	TypeInsufficientFunds   ErrorType = "insufficient_funds"
	TypeOperationNotAllowed ErrorType = "operation_not_allowed"
)

// errorTypeMeta holds the canonical Category, Code, and HTTP status for an ErrorType.
type errorTypeMeta struct {
	Category ErrorCategory
	Code     string
	Status   int
}

// errorTypeLookup maps every ErrorType to its canonical metadata.
// Use ErrorType.Meta() to look up a type's defaults rather than duplicating
// switch statements across the codebase.
var errorTypeLookup = map[ErrorType]errorTypeMeta{
	// Validation
	TypeValidation:    {CategoryValidation, CodeValidationError, http.StatusBadRequest},
	TypeRequired:      {CategoryValidation, CodeRequired, http.StatusBadRequest},
	TypeInvalidFormat: {CategoryValidation, CodeInvalidFormat, http.StatusBadRequest},
	TypeOutOfRange:    {CategoryValidation, CodeOutOfRange, http.StatusBadRequest},
	TypeDuplicate:     {CategoryValidation, CodeDuplicate, http.StatusBadRequest},
	// Auth
	TypeUnauthorized: {CategoryAuth, CodeUnauthorized, http.StatusUnauthorized},
	TypeForbidden:    {CategoryAuth, CodeForbidden, http.StatusForbidden},
	TypeInvalidToken: {CategoryAuth, CodeInvalidToken, http.StatusUnauthorized},
	TypeExpiredToken: {CategoryAuth, CodeExpiredToken, http.StatusUnauthorized},
	// Resource
	TypeNotFound:      {CategoryClient, CodeResourceNotFound, http.StatusNotFound},
	TypeConflict:      {CategoryClient, CodeConflict, http.StatusConflict},
	TypeAlreadyExists: {CategoryClient, CodeAlreadyExists, http.StatusConflict},
	TypeGone:          {CategoryClient, CodeGone, http.StatusGone},
	// System
	TypeInternal:    {CategoryServer, CodeInternalError, http.StatusInternalServerError},
	TypeUnavailable: {CategoryServer, CodeUnavailable, http.StatusServiceUnavailable},
	TypeTimeout:     {CategoryTimeout, CodeTimeout, http.StatusRequestTimeout},
	TypeRateLimited: {CategoryRateLimit, CodeRateLimited, http.StatusTooManyRequests},
	TypeMaintenance: {CategoryServer, CodeMaintenance, http.StatusServiceUnavailable},
	// Business
	TypeInvalidState:        {CategoryBusiness, CodeInvalidState, http.StatusUnprocessableEntity},
	TypeInsufficientFunds:   {CategoryBusiness, CodeInsufficientFunds, http.StatusUnprocessableEntity},
	TypeOperationNotAllowed: {CategoryBusiness, CodeOperationNotAllowed, http.StatusUnprocessableEntity},
}

// Meta returns the canonical Category, Code, and HTTP status for the ErrorType.
// If the type is unrecognized, it returns server-error defaults.
func (t ErrorType) Meta() errorTypeMeta {
	if m, ok := errorTypeLookup[t]; ok {
		return m
	}
	return errorTypeMeta{CategoryServer, CodeInternalError, http.StatusInternalServerError}
}

// APIError represents a normalized error payload for HTTP responses and logging.
//
// Callers outside this package should build APIError values through
// NewErrorBuilder(), rather than struct literals, to guarantee that all
// required fields (Status, Code, Category) are populated consistently.
type APIError struct {
	Status    int            `json:"-"`
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Category  ErrorCategory  `json:"category"`
	Type      ErrorType      `json:"type,omitempty"`
	Severity  ErrorSeverity  `json:"severity,omitempty"`
	RequestID string         `json:"-"`
	Details   map[string]any `json:"details,omitempty"`
}

// Error implements the error interface for APIError
func (e APIError) Error() string {
	return e.Message
}

// ErrorResponse wraps the error payload for consistent JSON responses.
type ErrorResponse struct {
	Error     APIError `json:"error"`
	RequestID string   `json:"request_id,omitempty"`
}

// WriteError writes a structured error response with trace context when available.
// It returns the encoding error, if any; callers may ignore it when the response
// headers have already been sent.
//
// Prefer building APIError values through NewErrorBuilder() so that required
// fields are always populated. WriteError keeps fallback defaults for
// backward compatibility and calls WarnFunc when required fields are missing.
func WriteError(w http.ResponseWriter, r *http.Request, err APIError) error {
	if w == nil {
		return ErrResponseWriterNil
	}
	if issues := validateAPIError(err); len(issues) > 0 {
		WarnFunc("WriteError received partially-populated APIError: " + strings.Join(issues, "; "))
	}

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

	if err.RequestID == "" && r != nil {
		if requestID := RequestIDFromContext(r.Context()); requestID != "" {
			err.RequestID = requestID
		}
	}

	resp := ErrorResponse{
		Error:     err,
		RequestID: err.RequestID,
	}

	buf := getJSONBuffer()
	defer putJSONBuffer(buf)

	if encErr := json.NewEncoder(buf).Encode(resp); encErr != nil {
		return encErr
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	_, writeErr := w.Write(buf.Bytes())
	return writeErr
}

// CategoryForStatus maps an HTTP status to a default error category.
func CategoryForStatus(status int) ErrorCategory {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return CategoryAuth
	case http.StatusTooManyRequests:
		return CategoryRateLimit
	case http.StatusRequestTimeout:
		return CategoryTimeout
	case http.StatusBadRequest, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity:
		return CategoryClient
	default:
		if status >= http.StatusInternalServerError {
			return CategoryServer
		}
		if status >= http.StatusBadRequest {
			return CategoryClient
		}
		return ""
	}
}

// ErrorBuilder provides a fluent builder for creating APIError instances.
type ErrorBuilder struct {
	err APIError
}

// NewErrorBuilder creates a new error builder with default values.
func NewErrorBuilder() *ErrorBuilder {
	return &ErrorBuilder{
		err: APIError{
			Status:  http.StatusInternalServerError,
			Details: make(map[string]any),
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

// Severity sets the severity level for the error.
func (b *ErrorBuilder) Severity(severity ErrorSeverity) *ErrorBuilder {
	if severity == "" {
		return b
	}
	b.err.Severity = severity
	return b
}

// Type sets the error type and populates Category, Code, and Status with the
// canonical values for that type. Any Category, Code, or Status set before
// calling Type will be overwritten. To customize those fields beyond the type
// defaults, call Category, Code, or Status after Type.
func (b *ErrorBuilder) Type(errorType ErrorType) *ErrorBuilder {
	if errorType == "" {
		return b
	}
	meta := errorType.Meta()
	b.err.Type = errorType
	b.err.Category = meta.Category
	b.err.Code = meta.Code
	b.err.Status = meta.Status
	return b
}

// RequestID sets the request id for the error.
func (b *ErrorBuilder) RequestID(requestID string) *ErrorBuilder {
	b.err.RequestID = requestID
	return b
}

// TypeOnly sets the Type field without changing Status, Category, or Code.
// Use this when Status, Category, and Code are already set explicitly and
// only need to tag the error's type for observability.
// Contrast with Type(), which also overwrites Status, Category, and Code from type metadata.
func (b *ErrorBuilder) TypeOnly(errorType ErrorType) *ErrorBuilder {
	if errorType != "" {
		b.err.Type = errorType
	}
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
// It fills any missing Status, Code, and Category with safe defaults so that
// every value returned by a builder is fully populated.
func (b *ErrorBuilder) Build() APIError {
	if b.err.Status == 0 {
		b.err.Status = http.StatusInternalServerError
	}
	if b.err.Code == "" {
		b.err.Code = http.StatusText(b.err.Status)
	}
	if b.err.Category == "" {
		b.err.Category = CategoryForStatus(b.err.Status)
		if b.err.Category == "" {
			b.err.Category = CategoryServer
		}
	}
	return b.err
}

// validateAPIError validates an APIError and returns validation errors if any.
func validateAPIError(err APIError) []string {
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
	if err.Type != "" {
		if _, ok := errorTypeLookup[err.Type]; !ok {
			validationErrors = append(validationErrors, "invalid error type")
		}
	}
	if err.Severity != "" && !isValidErrorSeverity(err.Severity) {
		validationErrors = append(validationErrors, "invalid error severity")
	}

	return validationErrors
}

func isValidErrorSeverity(severity ErrorSeverity) bool {
	switch severity {
	case SeverityInfo, SeverityWarning, SeverityError, SeverityCritical:
		return true
	default:
		return false
	}
}

// HTTPStatusFromCategory returns the representative HTTP status for a category.
// This is an intentionally coarse mapping; when a more specific ErrorType is
// known, use ErrorType.Meta().Status instead.
func HTTPStatusFromCategory(category ErrorCategory) int {
	switch category {
	case CategoryClient, CategoryValidation:
		return http.StatusBadRequest
	case CategoryAuth:
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
// Note: this closes resp.Body.
func ParseErrorFromResponse(resp *http.Response) (APIError, error) {
	if resp == nil || resp.Body == nil {
		return APIError{}, fmt.Errorf("response is nil")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 400 {
		return APIError{}, fmt.Errorf("no error in successful response")
	}

	var errorResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
		return NewErrorBuilder().
			Type(TypeInternal).
			Status(resp.StatusCode).
			Code(http.StatusText(resp.StatusCode)).
			Category(CategoryServer).
			Message(fmt.Sprintf("failed to parse error response: %v", err)).
			Build(), nil
	}

	if errorResp.Error.RequestID == "" {
		errorResp.Error.RequestID = errorResp.RequestID
	}
	return errorResp.Error, nil
}

// IsClientError reports whether err is a client error (4xx).
// It accepts any error and performs a type assertion to APIError internally.
func IsClientError(err error) bool {
	var apiErr APIError
	if !asAPIError(err, &apiErr) {
		return false
	}
	return apiErr.Status >= 400 && apiErr.Status < 500
}

// IsServerError reports whether err is a server error (5xx).
// It accepts any error and performs a type assertion to APIError internally.
func IsServerError(err error) bool {
	var apiErr APIError
	if !asAPIError(err, &apiErr) {
		return false
	}
	return apiErr.Status >= 500
}

// asAPIError attempts to extract an APIError from err via direct assertion or errors.As.
func asAPIError(err error, out *APIError) bool {
	if v, ok := err.(APIError); ok {
		*out = v
		return true
	}
	return errors.As(err, out)
}

// IsAPIErrorRetryable checks if the API error is retryable based on its status code.
func IsAPIErrorRetryable(err APIError) bool {
	switch err.Status {
	case http.StatusRequestTimeout,
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}

	return err.Category == CategoryTimeout
}
