package validate

import (
	"encoding/json"
	"net/http"

	"github.com/spcent/plumego/contract"
)

// Validator validates a decoded request value.
type Validator interface {
	Validate(v any) error
}

// ValidationError adapts validator failures to Plumego's contract error shape.
type ValidationError struct {
	contract.APIError
	Err error
}

// Error returns the underlying validator error message.
func (e ValidationError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.APIError.Error()
}

// Unwrap returns the underlying validator error.
func (e ValidationError) Unwrap() error {
	return e.Err
}

// BindJSON decodes the request JSON body into T without validation.
func BindJSON[T any](r *http.Request) (T, error) {
	var out T
	if err := json.NewDecoder(r.Body).Decode(&out); err != nil {
		return out, err
	}
	return out, nil
}

// Bind decodes the request JSON body into T and validates the decoded value.
func Bind[T any](r *http.Request, v Validator) (T, error) {
	out, err := BindJSON[T](r)
	if err != nil {
		return out, err
	}
	if v == nil {
		return out, nil
	}
	if err := v.Validate(out); err != nil {
		return out, newValidationError(err)
	}
	return out, nil
}

func newValidationError(err error) ValidationError {
	return ValidationError{
		APIError: contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Code(contract.CodeValidationError).
			Message("validation failed").
			Build(),
		Err: err,
	}
}
