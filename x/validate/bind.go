package validate

import (
	"encoding/json"
	"net/http"
)

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
