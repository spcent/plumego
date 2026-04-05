package contract

import (
	"fmt"
	"strings"
	"time"
)

// ErrorContext represents contextual information for error wrapping.
type ErrorContext struct {
	Operation string         `json:"operation"`
	Module    string         `json:"module"`
	Params    map[string]any `json:"params,omitempty"`
}

// WrappedErrorWithContext represents an error with attached operation context.
// Use WrapError to construct values; use errors.Is / errors.As to inspect them.
type WrappedErrorWithContext struct {
	Err     error        `json:"error"`
	Context ErrorContext `json:"context"`
	Message string       `json:"message"`
	When    time.Time    `json:"when"`
}

// Error implements the error interface.
func (w *WrappedErrorWithContext) Error() string {
	if w.Message != "" {
		return w.Message
	}
	if w.Err != nil {
		return w.Err.Error()
	}
	return "unknown error"
}

// Unwrap returns the underlying error for errors.Is/As support.
func (w *WrappedErrorWithContext) Unwrap() error {
	return w.Err
}

// NewWrappedError creates a new WrappedErrorWithContext with operation context.
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

// WrapError wraps an existing error with additional context.
// If err is already a *WrappedErrorWithContext, inner fields take precedence:
//   - Operation and Module: inner value is used if non-empty; outer is the fallback.
//   - Params: inner keys win; outer keys are added only if not already present.
//
// This preserves the original failure-site context. The intermediate wrapper is
// flattened — the returned error's Unwrap chain is: result -> inner.Err.
func WrapError(err error, operation, module string, params map[string]any) error {
	if err == nil {
		return nil
	}

	if wrapped, ok := err.(*WrappedErrorWithContext); ok {
		var mergedParams map[string]any
		if len(wrapped.Context.Params) > 0 || len(params) > 0 {
			mergedParams = make(map[string]any, len(wrapped.Context.Params)+len(params))
			for k, v := range wrapped.Context.Params {
				mergedParams[k] = v
			}
			for k, v := range params {
				if _, exists := mergedParams[k]; !exists {
					mergedParams[k] = v
				}
			}
		}

		op := wrapped.Context.Operation
		if op == "" {
			op = operation
		}
		mod := wrapped.Context.Module
		if mod == "" {
			mod = module
		}

		return &WrappedErrorWithContext{
			Err:     wrapped.Err,
			Message: wrapped.Message,
			When:    wrapped.When,
			Context: ErrorContext{
				Operation: op,
				Module:    mod,
				Params:    mergedParams,
			},
		}
	}

	return NewWrappedError(err, operation, module, params)
}

// WrapErrorf creates a message-only wrapper with no operation or module context.
// Use WrapError when structured context is required for logging or diagnostics.
func WrapErrorf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &WrappedErrorWithContext{
		Err:     err,
		Message: fmt.Sprintf(format, args...),
		When:    time.Now(),
	}
}

// IsRetryable reports whether err represents a transient condition that a caller
// may safely retry. It unwraps WrappedErrorWithContext chains and delegates to
// IsAPIErrorRetryable for APIError values.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	if apiErr, ok := err.(APIError); ok {
		return IsAPIErrorRetryable(apiErr)
	}

	if wrapped, ok := err.(*WrappedErrorWithContext); ok {
		if apiErr, ok := wrapped.Err.(APIError); ok {
			return IsAPIErrorRetryable(apiErr)
		}
		return IsRetryable(wrapped.Err)
	}

	if netErr, ok := err.(interface{ Timeout() bool }); ok {
		return netErr.Timeout()
	}

	return false
}

// GetErrorDetails extracts structured information from any error for logging or
// diagnostics. It handles APIError, WrappedErrorWithContext, and plain errors.
func GetErrorDetails(err error) map[string]any {
	if err == nil {
		return nil
	}

	details := make(map[string]any)

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
		if wrapped.Err != nil {
			details["wrapped_error"] = GetErrorDetails(wrapped.Err)
		}
		return details
	}

	if apiErr, ok := err.(APIError); ok {
		details["status"] = apiErr.Status
		details["code"] = apiErr.Code
		details["category"] = apiErr.Category
		details["message"] = apiErr.Message
		if apiErr.Type != "" {
			details["type"] = apiErr.Type
		}
		if apiErr.Severity != "" {
			details["severity"] = apiErr.Severity
		}
		if len(apiErr.Details) > 0 {
			details["details"] = apiErr.Details
		}
		if apiErr.TraceID != "" {
			details["trace_id"] = apiErr.TraceID
		}
		return details
	}

	details["message"] = err.Error()
	details["type"] = fmt.Sprintf("%T", err)
	return details
}

// FormatError returns a concise string representation of err for log messages.
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	details := GetErrorDetails(err)

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

// PanicToError converts a recovered panic value to a wrapped error.
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
