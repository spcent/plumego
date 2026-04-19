package validator

import (
	"encoding/json"
	"net/http"
)

func (v *Validator) BindJSON(r *http.Request, target any) error {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return err
	}
	return v.Validate(target)
}

func BindJSON(r *http.Request, target any) error {
	validator := NewValidator(nil)
	return validator.BindJSON(r, target)
}
