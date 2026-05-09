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
	TypeValidation     ErrorType = "validation_failure"
	TypeBadRequest     ErrorType = "bad_request"
	TypeInvalidRequest ErrorType = "invalid_request"
	TypeRequired       ErrorType = "required_field_missing"
	TypeInvalidFormat  ErrorType = "invalid_format"
	TypeOutOfRange     ErrorType = "value_out_of_range"
	TypeDuplicate      ErrorType = "duplicate_value"

	// Authentication/Authorization errors
	TypeUnauthorized ErrorType = "unauthorized_request"
	TypeForbidden    ErrorType = "forbidden_request"
	TypeInvalidToken ErrorType = "invalid_token"
	TypeExpiredToken ErrorType = "expired_token"

	// Resource errors
	TypeNotFound        ErrorType = "resource_not_found"
	TypeConflict        ErrorType = "resource_conflict"
	TypeAlreadyExists   ErrorType = "resource_already_exists"
	TypeGone            ErrorType = "resource_gone"
	TypeNotAcceptable   ErrorType = "not_acceptable"
	TypePayloadTooLarge ErrorType = "payload_too_large"

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
	TypeValidation:     {CategoryValidation, CodeValidationError, http.StatusBadRequest},
	TypeBadRequest:     {CategoryClient, CodeBadRequest, http.StatusBadRequest},
	TypeInvalidRequest: {CategoryClient, CodeInvalidRequest, http.StatusUnprocessableEntity},
	TypeRequired:       {CategoryValidation, CodeRequired, http.StatusBadRequest},
	TypeInvalidFormat:  {CategoryValidation, CodeInvalidFormat, http.StatusBadRequest},
	TypeOutOfRange:     {CategoryValidation, CodeOutOfRange, http.StatusBadRequest},
	TypeDuplicate:      {CategoryValidation, CodeDuplicate, http.StatusBadRequest},
	// Auth
	TypeUnauthorized: {CategoryAuth, CodeUnauthorized, http.StatusUnauthorized},
	TypeForbidden:    {CategoryAuth, CodeForbidden, http.StatusForbidden},
	TypeInvalidToken: {CategoryAuth, CodeInvalidToken, http.StatusUnauthorized},
	TypeExpiredToken: {CategoryAuth, CodeExpiredToken, http.StatusUnauthorized},
	// Resource
	TypeNotFound:        {CategoryClient, CodeResourceNotFound, http.StatusNotFound},
	TypeConflict:        {CategoryClient, CodeConflict, http.StatusConflict},
	TypeAlreadyExists:   {CategoryClient, CodeAlreadyExists, http.StatusConflict},
	TypeGone:            {CategoryClient, CodeGone, http.StatusGone},
	TypeNotAcceptable:   {CategoryClient, CodeInvalidRequest, http.StatusNotAcceptable},
	TypePayloadTooLarge: {CategoryClient, CodeRequestBodyTooLarge, http.StatusRequestEntityTooLarge},
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

var categoryByHTTPStatus = map[int]ErrorCategory{
	http.StatusUnauthorized:        CategoryAuth,
	http.StatusForbidden:           CategoryAuth,
	http.StatusTooManyRequests:     CategoryRateLimit,
	http.StatusRequestTimeout:      CategoryTimeout,
	http.StatusGatewayTimeout:      CategoryTimeout,
	http.StatusBadRequest:          CategoryClient,
	http.StatusNotFound:            CategoryClient,
	http.StatusConflict:            CategoryClient,
	http.StatusUnprocessableEntity: CategoryClient,
}

var codeByHTTPStatus = map[int]string{
	http.StatusBadRequest:            CodeBadRequest,
	http.StatusUnauthorized:          CodeUnauthorized,
	http.StatusForbidden:             CodeForbidden,
	http.StatusNotFound:              CodeResourceNotFound,
	http.StatusMethodNotAllowed:      CodeMethodNotAllowed,
	http.StatusConflict:              CodeConflict,
	http.StatusGone:                  CodeGone,
	http.StatusRequestEntityTooLarge: CodeRequestBodyTooLarge,
	http.StatusUnprocessableEntity:   CodeInvalidRequest,
	http.StatusTooManyRequests:       CodeRateLimited,
	http.StatusRequestTimeout:        CodeTimeout,
	http.StatusNotImplemented:        CodeNotImplemented,
	http.StatusBadGateway:            CodeBadGateway,
	http.StatusGatewayTimeout:        CodeGatewayTimeout,
	http.StatusServiceUnavailable:    CodeUnavailable,
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
// required fields are populated consistently. APIError is intentionally opaque;
// use its read-only methods to inspect normalized values.
type APIError struct {
	status    int
	code      string
	message   string
	category  ErrorCategory
	errorType ErrorType
	requestID string
	details   map[string]any
}

// Error implements the error interface for APIError
func (e APIError) Error() string {
	return e.message
}

// Status returns the normalized HTTP status associated with the error.
func (e APIError) Status() int {
	return normalizeAPIError(e).status
}

// Code returns the normalized machine-readable error code.
func (e APIError) Code() string {
	return normalizeAPIError(e).code
}

// Message returns the normalized client-facing error message.
func (e APIError) Message() string {
	return normalizeAPIError(e).message
}

// Category returns the normalized high-level error category.
func (e APIError) Category() ErrorCategory {
	return normalizeAPIError(e).category
}

// Type returns the normalized error type, or an empty value for untyped errors.
func (e APIError) Type() ErrorType {
	return normalizeAPIError(e).errorType
}

// RequestID returns the normalized request id associated with the error.
func (e APIError) RequestID() string {
	return normalizeAPIError(e).requestID
}

// Details returns an isolated copy of the normalized error detail map.
func (e APIError) Details() map[string]any {
	return cloneAnyMap(normalizeAPIError(e).details)
}

type errorPayload struct {
	Code     string         `json:"code"`
	Message  string         `json:"message"`
	Category ErrorCategory  `json:"category"`
	Type     ErrorType      `json:"type,omitempty"`
	Details  map[string]any `json:"details,omitempty"`
}

type errorResponse struct {
	Error     errorPayload `json:"error"`
	RequestID string       `json:"request_id,omitempty"`
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

	if err.requestID == "" && r != nil {
		if requestID := RequestIDFromContext(r.Context()); requestID != "" {
			err.requestID = requestID
		}
	}

	resp := errorResponse{
		Error:     errorPayloadFromAPIError(err),
		RequestID: err.requestID,
	}

	buf := getJSONBuffer()
	defer putJSONBuffer(buf)

	if encErr := json.NewEncoder(buf).Encode(resp); encErr != nil {
		return encErr
	}

	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(err.status)
	_, writeErr := w.Write(buf.Bytes())
	return writeErr
}

func errorPayloadFromAPIError(err APIError) errorPayload {
	return errorPayload{
		Code:     err.code,
		Message:  err.message,
		Category: err.category,
		Type:     err.errorType,
		Details:  cloneAnyMap(err.details),
	}
}

func categoryForStatus(status int) ErrorCategory {
	if category, ok := categoryByHTTPStatus[status]; ok {
		return category
	}
	if status >= http.StatusInternalServerError {
		return CategoryServer
	}
	if status >= http.StatusBadRequest {
		return CategoryClient
	}
	return ""
}

// ErrorBuilder provides a fluent builder for creating APIError instances.
type ErrorBuilder struct {
	err APIError
}

// NewErrorBuilder creates a new error builder with default values.
func NewErrorBuilder() *ErrorBuilder {
	return &ErrorBuilder{
		err: APIError{
			status:  http.StatusInternalServerError,
			details: make(map[string]any),
		},
	}
}

// Message sets the error message for the error.
func (b *ErrorBuilder) Message(message string) *ErrorBuilder {
	b.err.message = message
	return b
}

// Code sets an extension-owned machine code while preserving the type's
// canonical status and category.
func (b *ErrorBuilder) Code(code string) *ErrorBuilder {
	b.err.code = code
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
	b.err.errorType = errorType
	b.err.category = meta.Category
	b.err.code = meta.Code
	b.err.status = meta.Status
	return b
}

// RequestID sets the request id for the error.
func (b *ErrorBuilder) RequestID(requestID string) *ErrorBuilder {
	if requestID, ok := normalizeRequestID(requestID); ok {
		b.err.requestID = requestID
	} else {
		b.err.requestID = ""
	}
	return b
}

// Detail adds a detail field to the error.
func (b *ErrorBuilder) Detail(key string, value any) *ErrorBuilder {
	if key == "" {
		return b
	}
	b.ensureDetails()
	b.err.details[key] = value
	return b
}

// Details sets multiple detail fields for the error.
func (b *ErrorBuilder) Details(details map[string]any) *ErrorBuilder {
	b.ensureDetails()
	for k, v := range details {
		if k == "" {
			continue
		}
		b.err.details[k] = v
	}
	return b
}

func (b *ErrorBuilder) ensureDetails() {
	if b.err.details == nil {
		b.err.details = make(map[string]any)
	}
}

// Build creates the final APIError instance.
// It fills any missing status, code, and category with safe defaults so that
// every value returned by a builder is fully populated.
func (b *ErrorBuilder) Build() APIError {
	return normalizeAPIError(b.err)
}

func normalizeAPIError(err APIError) APIError {
	err, invalidStatus := normalizeAPIErrorBase(err)
	err, typed := normalizeTypedAPIError(err)
	if !typed {
		normalizeUntypedAPIError(&err, invalidStatus)
	}
	applyDefaultAPIErrorMessage(&err)
	return err
}

func normalizeAPIErrorBase(err APIError) (APIError, bool) {
	status, invalid := normalizeErrorHTTPStatus(err.status)
	err.status = status
	err.details = cloneAnyMap(err.details)
	err.requestID = normalizeRequestIDOrEmpty(err.requestID)
	return err, invalid
}

func normalizeTypedAPIError(err APIError) (APIError, bool) {
	if err.errorType != "" {
		if meta, ok := errorTypeLookup[err.errorType]; ok {
			err.status = meta.Status
			err.category = meta.Category
			if err.code == "" {
				err.code = meta.Code
			}
			return err, true
		} else {
			err.errorType = ""
		}
	}
	return err, false
}

func normalizeUntypedAPIError(err *APIError, invalidStatus bool) {
	if invalidStatus {
		err.category = CategoryServer
		if err.code == "" {
			err.code = CodeInternalError
		}
	} else if err.code == "" {
		err.code = codeForStatus(err.status)
	}

	if err.category == "" {
		err.category = categoryForStatus(err.status)
		if err.category == "" {
			err.category = CategoryServer
		}
	}
}

func normalizeRequestIDOrEmpty(requestID string) string {
	if requestID, ok := normalizeRequestID(requestID); ok {
		return requestID
	}
	return ""
}

func applyDefaultAPIErrorMessage(err *APIError) {
	if err.message == "" {
		err.message = http.StatusText(err.status)
	}
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
	if code, ok := codeByHTTPStatus[status]; ok {
		return code
	}
	if status >= http.StatusInternalServerError {
		return CodeInternalError
	}
	if status >= http.StatusBadRequest {
		return CodeInvalidRequest
	}
	return CodeInternalError
}
