package contract

import (
	"errors"
	"net/http"

	logpkg "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/validator"
)

// BindOptions configures JSON binding/validation for Ctx helpers.
type BindOptions struct {
	MaxBodyBytes          int64
	DisallowUnknownFields bool
	DisableValidation     bool
	Validator             func(any) error
	Logger                logpkg.StructuredLogger
	Redact                func(any) any
}

// FieldError represents a field-level validation error.
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// FieldErrorsFrom extracts field-level validation errors from an error.
func FieldErrorsFrom(err error) []FieldError {
	var fields []FieldError
	var fe validator.FieldErrors
	if errors.As(err, &fe) {
		for _, item := range fe.Errors() {
			fields = append(fields, FieldError{
				Field:   item.Field,
				Code:    item.Code,
				Message: item.Message,
			})
		}
	}
	return fields
}

// BindErrorToAPIError converts a binding/validation error into a structured APIError.
func BindErrorToAPIError(err error) APIError {
	fields := FieldErrorsFrom(err)

	status := http.StatusBadRequest
	code := "REQUEST_BIND_ERROR"
	message := "invalid request payload"
	category := CategoryValidation
	errorType := ErrTypeValidation

	var bindErr *BindError
	if errors.As(err, &bindErr) && bindErr != nil && bindErr.Message != "" {
		switch bindErr.Message {
		case ErrRequestBodyTooLarge.Error():
			status = http.StatusRequestEntityTooLarge
			code = "REQUEST_BODY_TOO_LARGE"
			message = ErrRequestBodyTooLarge.Error()
		case ErrEmptyRequestBody.Error():
			code = "EMPTY_BODY"
			message = ErrEmptyRequestBody.Error()
		case ErrInvalidJSON.Error():
			code = "INVALID_JSON"
			message = ErrInvalidJSON.Error()
		case ErrUnexpectedExtraData.Error(), "unexpected extra JSON data":
			code = "UNEXPECTED_EXTRA_DATA"
			message = ErrUnexpectedExtraData.Error()
		default:
			if len(fields) == 0 {
				message = bindErr.Message
			}
		}
	}

	if errors.Is(err, ErrRequestBodyTooLarge) {
		status = http.StatusRequestEntityTooLarge
		code = "REQUEST_BODY_TOO_LARGE"
		message = ErrRequestBodyTooLarge.Error()
	} else if errors.Is(err, ErrEmptyRequestBody) {
		code = "EMPTY_BODY"
		message = ErrEmptyRequestBody.Error()
	} else if errors.Is(err, ErrInvalidJSON) {
		code = "INVALID_JSON"
		message = ErrInvalidJSON.Error()
	} else if errors.Is(err, ErrUnexpectedExtraData) {
		code = "UNEXPECTED_EXTRA_DATA"
		message = ErrUnexpectedExtraData.Error()
	}

	if len(fields) > 0 {
		code = "VALIDATION_ERROR"
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
func WriteBindError(w http.ResponseWriter, r *http.Request, err error) {
	WriteError(w, r, BindErrorToAPIError(err))
}
