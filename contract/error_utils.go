package contract

import (
	"net/http"
	"runtime/debug"

	log "github.com/spcent/plumego/log"
)

// ErrorHandler provides unified error handling utilities
type ErrorHandler struct {
	logger log.StructuredLogger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger log.StructuredLogger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// Handle handles an error with context and returns an appropriate response
func (eh *ErrorHandler) Handle(
	w http.ResponseWriter,
	r *http.Request,
	err error,
	operation string,
	module string,
	params map[string]any,
) {
	if err == nil {
		return
	}

	// Wrap error with context
	wrappedErr := WrapError(err, operation, module, params)

	// Log the error with full context
	eh.logError(r, wrappedErr)

	// Convert to APIError for response
	apiErr := eh.toAPIError(wrappedErr)

	// Write error response
	WriteError(w, r, apiErr)
}

// HandlePanic recovers from panic and converts to error
func (eh *ErrorHandler) HandlePanic(r any) error {
	if r == nil {
		return nil
	}

	err := PanicToError(r)

	// Log panic with stack trace
	if eh.logger != nil {
		eh.logger.WithFields(log.Fields{
			"panic":     r,
			"stack":     string(debug.Stack()),
			"operation": "panic_recovery",
		}).Error("Recovered from panic", nil)
	}

	return err
}

// Wrap adds context to an error
func (eh *ErrorHandler) Wrap(err error, operation, module string, params map[string]any) error {
	return WrapError(err, operation, module, params)
}

// Wrapf adds context to an error with formatted message
func (eh *ErrorHandler) Wrapf(err error, format string, args ...any) error {
	return WrapErrorf(err, format, args...)
}

// IsRetryable checks if an error is retryable
func (eh *ErrorHandler) IsRetryable(err error) bool {
	return IsRetryable(err)
}

// ToAPIError converts any error to APIError
func (eh *ErrorHandler) ToAPIError(err error) APIError {
	return eh.toAPIError(err)
}

// toAPIError converts an error to APIError
func (eh *ErrorHandler) toAPIError(err error) APIError {
	if err == nil {
		return NewInternalError("unknown error")
	}

	// Handle APIError directly
	if apiErr, ok := err.(APIError); ok {
		return apiErr
	}

	// Handle WrappedErrorWithContext
	if wrapped, ok := err.(*WrappedErrorWithContext); ok {
		// If wrapped error is APIError, use it
		if apiErr, ok := wrapped.Err.(APIError); ok {
			return apiErr
		}
		// If wrapped error is another wrapped error, recurse
		if wrapped.Err != nil {
			return eh.toAPIError(wrapped.Err)
		}
		// Otherwise create internal error
		return NewInternalError(wrapped.Error())
	}

	// Handle standard errors
	return NewInternalError(err.Error())
}

// logError logs an error with full context
func (eh *ErrorHandler) logError(r *http.Request, err error) {
	if eh.logger == nil || err == nil {
		return
	}

	details := GetErrorDetails(err)
	fields := log.Fields{
		"operation": details["operation"],
		"module":    details["module"],
	}

	if params, ok := details["params"].(map[string]any); ok {
		for k, v := range params {
			fields[k] = v
		}
	}

	if wrappedErr, ok := err.(*WrappedErrorWithContext); ok {
		for k, v := range wrappedErr.Context.Params {
			fields[k] = v
		}
	}

	// Add trace ID if available
	if r != nil {
		if traceID := TraceIDFromContext(r.Context()); traceID != "" {
			fields["trace_id"] = traceID
		}
	}

	eh.logger.WithFields(fields).Error(FormatError(err), nil)
}

// WithTrace adds trace context to an error
func (eh *ErrorHandler) WithTrace(err error, r *http.Request) error {
	if err == nil || r == nil {
		return err
	}

	traceID := TraceIDFromContext(r.Context())
	if traceID == "" {
		return err
	}

	return WrapError(err, "add_trace", "error_handler", map[string]any{
		"trace_id": traceID,
	})
}

// ValidationError creates a validation error with context
func (eh *ErrorHandler) ValidationError(field, message string, params map[string]any) APIError {
	err := NewValidationError(field, message)
	for k, v := range params {
		err.Details[k] = v
	}

	return err
}

// NotFoundError creates a not found error with context
func (eh *ErrorHandler) NotFoundError(resource string, params map[string]any) APIError {
	err := NewNotFoundError(resource)
	for k, v := range params {
		err.Details[k] = v
	}

	return err
}

// UnauthorizedError creates an unauthorized error with context
func (eh *ErrorHandler) UnauthorizedError(message string, params map[string]any) APIError {
	err := NewUnauthorizedError(message)
	for k, v := range params {
		err.Details[k] = v
	}

	return err
}

// ForbiddenError creates a forbidden error with context
func (eh *ErrorHandler) ForbiddenError(message string, params map[string]any) APIError {
	err := NewForbiddenError(message)
	for k, v := range params {
		err.Details[k] = v
	}

	return err
}

// TimeoutError creates a timeout error with context
func (eh *ErrorHandler) TimeoutError(message string, params map[string]any) APIError {
	err := NewTimeoutError(message)
	for k, v := range params {
		err.Details[k] = v
	}

	return err
}

// InternalError creates an internal error with context
func (eh *ErrorHandler) InternalError(message string, params map[string]any) APIError {
	err := NewInternalError(message)
	for k, v := range params {
		err.Details[k] = v
	}

	return err
}

// RateLimitError creates a rate limit error with context
func (eh *ErrorHandler) RateLimitError(message string, params map[string]any) APIError {
	err := NewRateLimitError(message)
	for k, v := range params {
		err.Details[k] = v
	}

	return err
}

// SafeExecute executes a function and handles any errors safely
func (eh *ErrorHandler) SafeExecute(fn func() error, operation, module string, params map[string]any) error {
	defer func() {
		if r := recover(); r != nil {
			eh.HandlePanic(r)
		}
	}()

	err := fn()
	if err != nil {
		return WrapError(err, operation, module, params)
	}
	return nil
}

// SafeExecuteWithResult executes a function that returns a result and error
func SafeExecuteWithResult[T any](fn func() (T, error), operation, module string, params map[string]any) (T, error) {
	var zero T
	defer func() {
		if r := recover(); r != nil {
			// Log panic but don't return it since we can't change the signature
			if eh := NewErrorHandler(nil); eh != nil {
				eh.HandlePanic(r)
			}
		}
	}()

	result, err := fn()
	if err != nil {
		return zero, WrapError(err, operation, module, params)
	}
	return result, nil
}

// Global error handler instance
var globalErrorHandler = NewErrorHandler(nil)

// SetGlobalErrorHandler sets the global error handler
func SetGlobalErrorHandler(logger log.StructuredLogger) {
	globalErrorHandler = NewErrorHandler(logger)
}

// Global helper functions

// WrapGlobal wraps an error with context using the global error handler
func WrapGlobal(err error, operation, module string, params map[string]any) error {
	return globalErrorHandler.Wrap(err, operation, module, params)
}

// HandleGlobal handles an error using the global error handler
func HandleGlobal(w http.ResponseWriter, r *http.Request, err error, operation, module string, params map[string]any) {
	globalErrorHandler.Handle(w, r, err, operation, module, params)
}

// ToAPIErrorGlobal converts an error to APIError using the global error handler
func ToAPIErrorGlobal(err error) APIError {
	return globalErrorHandler.ToAPIError(err)
}
