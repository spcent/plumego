package contract

import (
	"errors"
	"net/http"
)

// BindOptions configures JSON binding/validation for Ctx helpers.
type BindOptions struct {
	// MaxBodySize caps the body size checked after the initial read.
	// Zero means no additional cap beyond what is set in RequestConfig.
	MaxBodySize           int64
	DisallowUnknownFields bool
	DisableValidation     bool
	Validator             func(any) error
	Redact                func(any) any
}

// FieldError represents a field-level validation error.
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// fieldErrorProvider is implemented by validationErrors (and any custom validator
// that produces []FieldError directly).
type fieldErrorProvider interface {
	Errors() []FieldError
}

// FieldErrorsFrom extracts field-level validation errors from an error.
func FieldErrorsFrom(err error) []FieldError {
	for current := err; current != nil; current = errors.Unwrap(current) {
		if p, ok := current.(fieldErrorProvider); ok {
			return p.Errors()
		}
	}
	return nil
}

// BindErrorToAPIError converts a binding/validation error into a structured APIError.
func BindErrorToAPIError(err error) APIError {
	fields := FieldErrorsFrom(err)

	status := http.StatusBadRequest
	code := CodeRequestBindError
	message := "invalid request payload"
	category := CategoryValidation
	errorType := ErrTypeValidation

	switch {
	case errors.Is(err, ErrRequestBodyTooLarge):
		status = http.StatusRequestEntityTooLarge
		code = CodeRequestBodyTooLarge
		message = ErrRequestBodyTooLarge.Error()
	case errors.Is(err, ErrEmptyRequestBody):
		code = CodeEmptyBody
		message = ErrEmptyRequestBody.Error()
	case errors.Is(err, ErrInvalidJSON):
		code = CodeInvalidJSON
		message = ErrInvalidJSON.Error()
	case errors.Is(err, ErrUnexpectedExtraData):
		code = CodeUnexpectedExtraData
		message = ErrUnexpectedExtraData.Error()
	default:
		if len(fields) == 0 {
			var bindErr *BindError
			if errors.As(err, &bindErr) && bindErr != nil && bindErr.Message != "" {
				message = bindErr.Message
				if bindErr.Err != nil {
					code = CodeRequestBodyReadFailed
				}
			}
		}
	}

	if len(fields) > 0 {
		code = CodeValidationError
		message = "validation failed"
	}

	builder := NewErrorBuilder().
		Status(status).
		Category(category).
		Type(errorType).
		Code(code).
		Message(message)

	if len(fields) > 0 {
		builder.Detail("fields", fields)
	}

	return builder.Build()
}

// WriteBindError writes a binding/validation error using the standard error envelope.
func WriteBindError(w http.ResponseWriter, r *http.Request, err error) error {
	return WriteError(w, r, BindErrorToAPIError(err))
}
