package contract

import (
	"errors"
	"fmt"
	"net/mail"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ValidationErrors is the error type returned by ValidateStruct when one or more
// field-level validation rules fail. Callers can inspect it with errors.As:
//
//	var verr contract.ValidationErrors
//	if errors.As(err, &verr) {
//	    for _, fe := range verr.Errors() { ... }
//	}
type ValidationErrors struct {
	errors []FieldError
}

func (ve ValidationErrors) Error() string {
	if len(ve.errors) == 0 {
		return ""
	}

	parts := make([]string, 0, len(ve.errors))
	for _, item := range ve.errors {
		parts = append(parts, fmt.Sprintf("%s: %s (%s)", item.Field, item.Message, item.Code))
	}
	return strings.Join(parts, "; ")
}

func (ve ValidationErrors) Errors() []FieldError {
	return append([]FieldError(nil), ve.errors...)
}

// ValidateStruct validates a struct using the package's `validate` tag rules.
// It is the explicit validation step after BindJSON or BindQuery.
func ValidateStruct(dst any) error {
	return validateStructAtDepth(dst, "", 0)
}

func validateStructAtDepth(dst any, prefix string, depth int) error {
	if dst == nil {
		return nil
	}
	if depth > 10 {
		fieldName := prefix
		if fieldName == "" {
			fieldName = "(root)"
		}
		return ValidationErrors{errors: []FieldError{{
			Field:   fieldName,
			Code:    CodeOutOfRange,
			Message: "struct nesting exceeds maximum validation depth (10); deeper fields were not validated",
		}}}
	}

	rv := reflect.ValueOf(dst)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}

	rt := rv.Type()
	var issues []FieldError
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.PkgPath != "" {
			continue
		}

		fieldValue := rv.Field(i)
		fieldName := field.Name
		if prefix != "" {
			fieldName = prefix + "." + fieldName
		}

		tag := strings.TrimSpace(field.Tag.Get("validate"))
		if tag == "-" {
			continue
		}
		if tag != "" {
			for _, rule := range strings.Split(tag, ",") {
				rule = strings.TrimSpace(rule)
				if rule == "" {
					continue
				}

				issue, tagErr := applyValidationRule(fieldName, fieldValue, rule)
				if tagErr != nil {
					// Programmer error: unknown or mis-configured rule.
					// Propagate directly so callers can distinguish it from
					// data validation failures.
					return tagErr
				}
				if issue != nil {
					issues = append(issues, *issue)
				}
			}
		}

		nested, tagErr := validateNestedStructField(fieldValue, fieldName, depth+1)
		if tagErr != nil {
			return tagErr
		}
		if len(nested.errors) > 0 {
			issues = append(issues, nested.Errors()...)
		}
	}

	if len(issues) == 0 {
		return nil
	}
	return ValidationErrors{errors: issues}
}

func validateNestedStructField(value reflect.Value, fieldName string, depth int) (ValidationErrors, error) {
	if !value.IsValid() {
		return ValidationErrors{}, nil
	}

	switch value.Kind() {
	case reflect.Ptr:
		if value.IsNil() || value.Elem().Kind() != reflect.Struct {
			return ValidationErrors{}, nil
		}
		if err := validateStructAtDepth(value.Interface(), fieldName, depth); err != nil {
			var issues ValidationErrors
			if errors.As(err, &issues) {
				return issues, nil
			}
			return ValidationErrors{}, err
		}
	case reflect.Struct:
		if value.CanAddr() {
			if err := validateStructAtDepth(value.Addr().Interface(), fieldName, depth); err != nil {
				var issues ValidationErrors
				if errors.As(err, &issues) {
					return issues, nil
				}
				return ValidationErrors{}, err
			}
		}
	case reflect.Slice, reflect.Array:
		var all []FieldError
		for i := 0; i < value.Len(); i++ {
			elem := value.Index(i)
			nested, err := validateNestedStructField(elem, fmt.Sprintf("%s[%d]", fieldName, i), depth)
			if err != nil {
				return ValidationErrors{}, err
			}
			if len(nested.errors) > 0 {
				all = append(all, nested.Errors()...)
			}
		}
		if len(all) > 0 {
			return ValidationErrors{errors: all}, nil
		}
	}

	return ValidationErrors{}, nil
}

// applyValidationRule applies a single validation rule to a field value.
// It returns a *FieldError for data validation failures, nil when the value
// passes, or a non-nil error (not *FieldError) when the rule name is unknown —
// signalling a programmer configuration mistake, not a user input problem.
func applyValidationRule(fieldName string, value reflect.Value, rule string) (*FieldError, error) {
	name := rule
	arg := ""
	if idx := strings.Index(rule, "="); idx >= 0 {
		name = strings.TrimSpace(rule[:idx])
		arg = strings.TrimSpace(rule[idx+1:])
	}

	switch name {
	case "required":
		return validateRequired(fieldName, value), nil
	case "email":
		return validateEmail(fieldName, value), nil
	case "min":
		limit, err := strconv.ParseInt(arg, 10, 64)
		if err != nil || limit < 0 {
			return nil, fmt.Errorf("invalid min validation rule %q on field %s", rule, fieldName)
		}
		return validateMin(fieldName, value, limit), nil
	case "max":
		limit, err := strconv.ParseInt(arg, 10, 64)
		if err != nil || limit < 0 {
			return nil, fmt.Errorf("invalid max validation rule %q on field %s", rule, fieldName)
		}
		return validateMax(fieldName, value, limit), nil
	default:
		return nil, fmt.Errorf("unknown validation rule %q on field %s", name, fieldName)
	}
}

func validateRequired(fieldName string, value reflect.Value) *FieldError {
	if !value.IsValid() {
		return &FieldError{Field: fieldName, Code: CodeRequired, Message: "field is required"}
	}

	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return &FieldError{Field: fieldName, Code: CodeRequired, Message: "field is required"}
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.String:
		if strings.TrimSpace(value.String()) == "" {
			return &FieldError{Field: fieldName, Code: CodeRequired, Message: "field is required"}
		}
	case reflect.Slice, reflect.Map, reflect.Array, reflect.Chan:
		if value.Len() == 0 {
			return &FieldError{Field: fieldName, Code: CodeRequired, Message: "field is required"}
		}
	default:
		if value.IsZero() {
			return &FieldError{Field: fieldName, Code: CodeRequired, Message: "field is required"}
		}
	}

	return nil
}

func validateEmail(fieldName string, value reflect.Value) *FieldError {
	s, ok := stringValue(value)
	if !ok {
		return &FieldError{Field: fieldName, Code: CodeInvalidFormat, Message: "must be a string"}
	}

	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	addr, err := mail.ParseAddress(s)
	if err != nil || addr.Address != s {
		return &FieldError{Field: fieldName, Code: CodeInvalidFormat, Message: "invalid email format"}
	}
	return nil
}

func validateMin(fieldName string, value reflect.Value, limit int64) *FieldError {
	if !value.IsValid() {
		return nil
	}

	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.String:
		text := value.String()
		if int64(utf8.RuneCountInString(text)) < limit {
			return &FieldError{Field: fieldName, Code: CodeOutOfRange, Message: fmt.Sprintf("must be at least %d characters", limit)}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Int() < limit {
			return &FieldError{Field: fieldName, Code: CodeOutOfRange, Message: fmt.Sprintf("must be at least %d", limit)}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if limit > 0 && value.Uint() < uint64(limit) {
			return &FieldError{Field: fieldName, Code: CodeOutOfRange, Message: fmt.Sprintf("must be at least %d", limit)}
		}
	case reflect.Float32, reflect.Float64:
		if value.Float() < float64(limit) {
			return &FieldError{Field: fieldName, Code: CodeOutOfRange, Message: fmt.Sprintf("must be at least %d", limit)}
		}
	}

	return nil
}

func validateMax(fieldName string, value reflect.Value, limit int64) *FieldError {
	if !value.IsValid() {
		return nil
	}

	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.String:
		text := value.String()
		if int64(utf8.RuneCountInString(text)) > limit {
			return &FieldError{Field: fieldName, Code: CodeOutOfRange, Message: fmt.Sprintf("must be at most %d characters", limit)}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Int() > limit {
			return &FieldError{Field: fieldName, Code: CodeOutOfRange, Message: fmt.Sprintf("must be at most %d", limit)}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if value.Uint() > uint64(limit) {
			return &FieldError{Field: fieldName, Code: CodeOutOfRange, Message: fmt.Sprintf("must be at most %d", limit)}
		}
	case reflect.Float32, reflect.Float64:
		if value.Float() > float64(limit) {
			return &FieldError{Field: fieldName, Code: CodeOutOfRange, Message: fmt.Sprintf("must be at most %d", limit)}
		}
	}

	return nil
}

func stringValue(value reflect.Value) (string, bool) {
	if !value.IsValid() {
		return "", false
	}
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return "", true
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.String {
		return "", false
	}
	return value.String(), true
}
