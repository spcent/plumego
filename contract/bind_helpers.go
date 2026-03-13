package contract

import (
	"errors"
	"net/http"
	"reflect"

	logpkg "github.com/spcent/plumego/log"
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
	for current := err; current != nil; current = errors.Unwrap(current) {
		if fields, ok := extractFieldErrors(current); ok {
			return fields
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
func WriteBindError(w http.ResponseWriter, r *http.Request, err error) {
	WriteError(w, r, BindErrorToAPIError(err))
}

func extractFieldErrors(err error) ([]FieldError, bool) {
	if err == nil {
		return nil, false
	}

	method := reflect.ValueOf(err).MethodByName("Errors")
	if !method.IsValid() || method.Type().NumIn() != 0 || method.Type().NumOut() != 1 {
		return nil, false
	}

	results := method.Call(nil)
	if len(results) != 1 {
		return nil, false
	}

	value := results[0]
	if value.Kind() != reflect.Slice {
		return nil, false
	}

	fields := make([]FieldError, 0, value.Len())
	for i := 0; i < value.Len(); i++ {
		item := value.Index(i)
		if item.Kind() == reflect.Pointer {
			if item.IsNil() {
				continue
			}
			item = item.Elem()
		}
		if item.Kind() != reflect.Struct {
			return nil, false
		}

		field := FieldError{
			Field:   structStringField(item, "Field"),
			Code:    structStringField(item, "Code"),
			Message: structStringField(item, "Message"),
		}
		if field.Field == "" && field.Code == "" && field.Message == "" {
			return nil, false
		}
		fields = append(fields, field)
	}

	return fields, len(fields) > 0
}

func structStringField(item reflect.Value, name string) string {
	field := item.FieldByName(name)
	if !field.IsValid() || field.Kind() != reflect.String {
		return ""
	}
	return field.String()
}
