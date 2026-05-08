package contract

import (
	"encoding/json"
	"net/http"
	"reflect"
)

// ErrorCategory describes the high-level class of an API error for observability.
type ErrorCategory string

const (
	// CategoryClient covers 4xx errors caused by client input.
	CategoryClient ErrorCategory = "client_error"
	// CategoryServer covers 5xx errors caused by infrastructure or server logic.
	CategoryServer ErrorCategory = "server_error"
	// CategoryTimeout covers timeout errors.
	CategoryTimeout ErrorCategory = "timeout_error"
	// CategoryValidation covers input validation errors.
	CategoryValidation ErrorCategory = "validation_error"
	// CategoryAuth covers authentication and authorization errors.
	CategoryAuth ErrorCategory = "auth_error"
	// CategoryRateLimit covers rate limiting errors.
	CategoryRateLimit ErrorCategory = "rate_limit_error"
)

// ErrorType represents specific error types for better categorization.
type ErrorType string

const (
	// Validation errors
	TypeValidation    ErrorType = "validation_failure"
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
	TypeInternal         ErrorType = "internal_error"
	TypeUnavailable      ErrorType = "service_unavailable"
	TypeTimeout          ErrorType = "timeout_failure"
	TypeRateLimited      ErrorType = "rate_limited"
	TypeMaintenance      ErrorType = "maintenance_mode"
	TypeMethodNotAllowed ErrorType = "method_not_allowed"
	TypeNotImplemented   ErrorType = "not_implemented"
	TypeBadGateway       ErrorType = "bad_gateway"
	TypeGatewayTimeout   ErrorType = "gateway_timeout"
)

// ErrorTypeMeta holds the canonical Category, Code, and HTTP status for an ErrorType.
type ErrorTypeMeta struct {
	Category ErrorCategory
	Code     string
	Status   int
}

// errorTypeLookup maps every ErrorType to its canonical metadata.
// Use ErrorType.Meta() to look up a type's defaults rather than duplicating
// switch statements across the codebase.
var errorTypeLookup = map[ErrorType]ErrorTypeMeta{
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
	TypeInternal:         {CategoryServer, CodeInternalError, http.StatusInternalServerError},
	TypeUnavailable:      {CategoryServer, CodeUnavailable, http.StatusServiceUnavailable},
	TypeTimeout:          {CategoryTimeout, CodeTimeout, http.StatusRequestTimeout},
	TypeRateLimited:      {CategoryRateLimit, CodeRateLimited, http.StatusTooManyRequests},
	TypeMaintenance:      {CategoryServer, CodeMaintenance, http.StatusServiceUnavailable},
	TypeMethodNotAllowed: {CategoryClient, CodeMethodNotAllowed, http.StatusMethodNotAllowed},
	TypeNotImplemented:   {CategoryServer, CodeNotImplemented, http.StatusNotImplemented},
	TypeBadGateway:       {CategoryServer, CodeBadGateway, http.StatusBadGateway},
	TypeGatewayTimeout:   {CategoryTimeout, CodeGatewayTimeout, http.StatusGatewayTimeout},
}

// Meta returns the canonical Category, Code, and HTTP status for the ErrorType.
// If the type is unrecognized, it returns server-error defaults.
func (t ErrorType) Meta() ErrorTypeMeta {
	if m, ok := errorTypeLookup[t]; ok {
		return m
	}
	return ErrorTypeMeta{CategoryServer, CodeInternalError, http.StatusInternalServerError}
}

// APIError represents a normalized error payload for HTTP responses and logging.
//
// Callers outside this package should build APIError values through
// NewErrorBuilder(), rather than struct literals, to guarantee that all
// required fields (Status, Code, Category) are populated consistently. Direct
// literals are retained for compatibility and are normalized by WriteError.
type APIError struct {
	Status    int            `json:"-"`
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Category  ErrorCategory  `json:"category"`
	Type      ErrorType      `json:"type,omitempty"`
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

// WriteError writes a structured error response with request id context when available.
// It returns the encoding error, if any; callers may ignore it when the response
// headers have already been sent.
//
// Prefer building APIError values through NewErrorBuilder() so that required
// fields are always populated. WriteError still normalizes incomplete APIError
// values, but it does so deterministically with no package-global side effects.
func WriteError(w http.ResponseWriter, r *http.Request, err APIError) error {
	if w == nil {
		return ErrResponseWriterNil
	}

	err = normalizeAPIError(err)

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

	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(err.Status)
	_, writeErr := w.Write(buf.Bytes())
	return writeErr
}

func categoryForStatus(status int) ErrorCategory {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return CategoryAuth
	case http.StatusTooManyRequests:
		return CategoryRateLimit
	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
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
// It preserves caller input; prefer Code* constants or uppercase stable codes.
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

// Type sets the error type and populates Category, Code, and Status with the
// canonical values for that type. Build normalizes typed errors back to the
// type's canonical Status and Category, so callers may only customize Code
// and other non-classification fields after Type.
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
	if requestID, ok := normalizeRequestID(requestID); ok {
		b.err.RequestID = requestID
	} else {
		b.err.RequestID = ""
	}
	return b
}

// Detail adds a detail field to the error.
func (b *ErrorBuilder) Detail(key string, value any) *ErrorBuilder {
	if key == "" {
		return b
	}
	b.ensureDetails()
	b.err.Details[key] = value
	return b
}

// Details sets multiple detail fields for the error.
func (b *ErrorBuilder) Details(details map[string]any) *ErrorBuilder {
	b.ensureDetails()
	for k, v := range details {
		if k == "" {
			continue
		}
		b.err.Details[k] = v
	}
	return b
}

func (b *ErrorBuilder) ensureDetails() {
	if b.err.Details == nil {
		b.err.Details = make(map[string]any)
	}
}

// Build creates the final APIError instance.
// It fills any missing Status, Code, and Category with safe defaults so that
// every value returned by a builder is fully populated.
func (b *ErrorBuilder) Build() APIError {
	return normalizeAPIError(b.err)
}

func normalizeAPIError(err APIError) APIError {
	status, invalid := normalizeErrorHTTPStatus(err.Status)
	err.Status = status
	err.Details = cloneAnyMap(err.Details)
	if requestID, ok := normalizeRequestID(err.RequestID); ok {
		err.RequestID = requestID
	} else {
		err.RequestID = ""
	}

	typed := false
	if err.Type != "" {
		if meta, ok := errorTypeLookup[err.Type]; ok {
			typed = true
			err.Status = meta.Status
			err.Category = meta.Category
			if err.Code == "" {
				err.Code = meta.Code
			}
		} else {
			err.Type = ""
		}
	}

	if !typed {
		if invalid {
			err.Category = CategoryServer
			if err.Code == "" {
				err.Code = CodeInternalError
			}
		} else if err.Code == "" {
			err.Code = codeForStatus(err.Status)
		}

		if err.Category == "" {
			err.Category = categoryForStatus(err.Status)
			if err.Category == "" {
				err.Category = CategoryServer
			}
		}
	}

	if err.Message == "" {
		err.Message = http.StatusText(err.Status)
	}

	return err
}

func cloneAnyMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		if k == "" {
			continue
		}
		out[k] = cloneDetailValue(v)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func cloneDetailValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return cloneAnyMap(v)
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = cloneDetailValue(item)
		}
		return out
	case []map[string]any:
		out := make([]map[string]any, len(v))
		for i, item := range v {
			out[i] = cloneAnyMap(item)
		}
		return out
	case []string:
		return append([]string(nil), v...)
	case []int:
		return append([]int(nil), v...)
	case []int64:
		return append([]int64(nil), v...)
	case []float64:
		return append([]float64(nil), v...)
	case []bool:
		return append([]bool(nil), v...)
	default:
		if cloned, ok := cloneReflectDetailValue(value); ok {
			return cloned
		}
		return value
	}
}

func cloneReflectDetailValue(value any) (any, bool) {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return value, true
	}
	cloned, ok := cloneReflectValue(v, 0)
	if !ok {
		return value, false
	}
	return cloned.Interface(), true
}

func cloneReflectValue(v reflect.Value, depth int) (reflect.Value, bool) {
	if depth > 16 {
		return v, false
	}
	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			return reflect.Zero(v.Type()), true
		}
		cloned, ok := cloneReflectValue(v.Elem(), depth+1)
		if !ok {
			return v, false
		}
		if cloned.Type().AssignableTo(v.Type()) {
			return cloned, true
		}
		if cloned.Type().AssignableTo(v.Type().Elem()) {
			out := reflect.New(v.Type()).Elem()
			out.Set(cloned)
			return out, true
		}
		return v, false
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return v, false
		}
		if v.IsNil() {
			return reflect.Zero(v.Type()), true
		}
		out := reflect.MakeMapWithSize(v.Type(), v.Len())
		iter := v.MapRange()
		for iter.Next() {
			cloned, ok := cloneReflectValue(iter.Value(), depth+1)
			if !ok {
				return v, false
			}
			cloned, ok = makeAssignable(cloned, v.Type().Elem())
			if !ok {
				return v, false
			}
			out.SetMapIndex(iter.Key(), cloned)
		}
		return out, true
	case reflect.Slice:
		if v.IsNil() {
			return reflect.Zero(v.Type()), true
		}
		out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			cloned, ok := cloneReflectValue(v.Index(i), depth+1)
			if !ok {
				return v, false
			}
			cloned, ok = makeAssignable(cloned, v.Type().Elem())
			if !ok {
				return v, false
			}
			out.Index(i).Set(cloned)
		}
		return out, true
	case reflect.Array:
		out := reflect.New(v.Type()).Elem()
		for i := 0; i < v.Len(); i++ {
			cloned, ok := cloneReflectValue(v.Index(i), depth+1)
			if !ok {
				return v, false
			}
			cloned, ok = makeAssignable(cloned, v.Type().Elem())
			if !ok {
				return v, false
			}
			out.Index(i).Set(cloned)
		}
		return out, true
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return v, true
	default:
		return v, false
	}
}

func makeAssignable(v reflect.Value, target reflect.Type) (reflect.Value, bool) {
	if v.Type().AssignableTo(target) {
		return v, true
	}
	if v.Type().ConvertibleTo(target) {
		return v.Convert(target), true
	}
	return v, false
}

func normalizeErrorHTTPStatus(status int) (int, bool) {
	if status == 0 {
		return http.StatusInternalServerError, false
	}
	if status < http.StatusBadRequest || status > 599 {
		return http.StatusInternalServerError, true
	}
	return status, false
}

func codeForStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return CodeBadRequest
	case http.StatusUnauthorized:
		return CodeUnauthorized
	case http.StatusForbidden:
		return CodeForbidden
	case http.StatusNotFound:
		return CodeResourceNotFound
	case http.StatusMethodNotAllowed:
		return CodeMethodNotAllowed
	case http.StatusConflict:
		return CodeConflict
	case http.StatusGone:
		return CodeGone
	case http.StatusRequestEntityTooLarge:
		return CodeRequestBodyTooLarge
	case http.StatusUnprocessableEntity:
		return CodeInvalidRequest
	case http.StatusTooManyRequests:
		return CodeRateLimited
	case http.StatusRequestTimeout:
		return CodeTimeout
	case http.StatusNotImplemented:
		return CodeNotImplemented
	case http.StatusBadGateway:
		return CodeBadGateway
	case http.StatusGatewayTimeout:
		return CodeGatewayTimeout
	case http.StatusServiceUnavailable:
		return CodeUnavailable
	default:
		if status >= http.StatusInternalServerError {
			return CodeInternalError
		}
		if status >= http.StatusBadRequest {
			return CodeInvalidRequest
		}
		return CodeInternalError
	}
}
