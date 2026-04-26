package contract

import (
	"errors"
	"net/http"
)

// BindOptions configures JSON binding for Ctx helpers.
type BindOptions struct {
	// MaxBodySize is a per-call size limit applied after the body has already
	// been read into memory. It does not prevent memory allocation for bodies
	// within the global RequestConfig.MaxBodySize limit. Use RequestConfig.
	// MaxBodySize for read-time protection; use this field only for stricter
	// endpoint-specific post-read caps.
	MaxBodySize           int64
	DisallowUnknownFields bool
}

// FieldError represents a field-level validation error.
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// FieldErrorsFrom extracts field-level validation errors from an error.
func FieldErrorsFrom(err error) []FieldError {
	var ve ValidationErrors
	if errors.As(err, &ve) {
		return ve.Errors()
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
	errorType := TypeInvalidFormat

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
	case errors.Is(err, ErrInvalidBindDst):
		code = CodeInvalidBindDst
		message = "invalid bind destination"
	case errors.Is(err, ErrInvalidQueryParam):
		code = CodeInvalidQuery
		message = "invalid query parameter"
	case errors.Is(err, ErrContextNil):
		code = CodeInvalidRequest
		message = ErrContextNil.Error()
	case errors.Is(err, ErrRequestNil):
		code = CodeInvalidRequest
		message = ErrRequestNil.Error()
	case errors.Is(err, ErrInvalidParam):
		code = CodeInvalidRequest
		message = ErrInvalidParam.Error()
	default:
		if len(fields) == 0 {
			var bindErr *bindError
			if errors.As(err, &bindErr) && bindErr != nil {
				if bindErr.Message != "" {
					message = bindErr.Message
				}
				if bindErr.Status != 0 {
					status = bindErr.Status
				}
				if bindErr.Err != nil {
					code = CodeRequestBodyReadFailed
				}
			}
		}
	}

	if len(fields) > 0 {
		errorType = TypeValidation
		code = CodeValidationError
		message = "validation failed"
	}

	builder := NewErrorBuilder().
		Status(status).
		Category(category).
		Code(code).
		Message(message).
		TypeOnly(errorType)

	if len(fields) > 0 {
		builder.Detail("fields", fields)
	}

	return builder.Build()
}

// WriteBindError writes a binding/validation error using the standard error envelope.
func WriteBindError(w http.ResponseWriter, r *http.Request, err error) error {
	return WriteError(w, r, BindErrorToAPIError(err))
}
