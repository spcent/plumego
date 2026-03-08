// Package routeparam provides concrete ParamValidator implementations and
// predefined validators for common route parameter patterns.
//
// Use these with router.RouteValidation to validate path parameters before
// the handler is called:
//
//	r.AddValidation("GET", "/users/:id", router.NewRouteValidation().
//	    AddParam("id", routeparam.UUID))
package routeparam

import (
	"fmt"
	"regexp"
	"strings"
)

// Validator is the interface satisfied by all routeparam validators.
// It is structurally identical to router.ParamValidator, so any value
// that satisfies Validator also satisfies router.ParamValidator.
type Validator interface {
	Validate(name, value string) error
}

// RegexValidator validates parameters using a regular expression.
type RegexValidator struct {
	pattern *regexp.Regexp
}

// NewRegex creates a validator that accepts values matching the given pattern.
func NewRegex(pattern string) (*RegexValidator, error) {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &RegexValidator{pattern: compiled}, nil
}

// MustRegex is like NewRegex but panics if the pattern fails to compile.
func MustRegex(pattern string) *RegexValidator {
	v, err := NewRegex(pattern)
	if err != nil {
		panic(fmt.Sprintf("routeparam: invalid regex pattern %q: %v", pattern, err))
	}
	return v
}

// Validate implements Validator.
func (rv *RegexValidator) Validate(name, value string) error {
	if !rv.pattern.MatchString(value) {
		return fmt.Errorf("parameter %s value %s does not match pattern %s", name, value, rv.pattern.String())
	}
	return nil
}

// RangeValidator validates that a numeric parameter falls within [min, max].
type RangeValidator struct {
	min, max int64
}

// NewRange creates a validator that accepts integers in [min, max].
func NewRange(min, max int64) *RangeValidator {
	return &RangeValidator{min: min, max: max}
}

// Validate implements Validator.
func (rv *RangeValidator) Validate(name, value string) error {
	var num int64
	if _, err := fmt.Sscanf(value, "%d", &num); err != nil {
		return fmt.Errorf("parameter %s value %s is not a valid integer", name, value)
	}
	if num < rv.min || num > rv.max {
		return fmt.Errorf("parameter %s value %d is not in range [%d, %d]", name, num, rv.min, rv.max)
	}
	return nil
}

// LengthValidator validates string length falls within [min, max].
type LengthValidator struct {
	min, max int
}

// NewLength creates a validator that accepts strings with length in [min, max].
func NewLength(min, max int) *LengthValidator {
	return &LengthValidator{min: min, max: max}
}

// Validate implements Validator.
func (lv *LengthValidator) Validate(name, value string) error {
	length := len(value)
	if length < lv.min || length > lv.max {
		return fmt.Errorf("parameter %s length %d is not in range [%d, %d]", name, length, lv.min, lv.max)
	}
	return nil
}

// CompositeValidator applies multiple validators in sequence, failing on the first error.
type CompositeValidator struct {
	validators []Validator
}

// NewComposite creates a validator that runs all given validators in order.
func NewComposite(validators ...Validator) *CompositeValidator {
	return &CompositeValidator{validators: validators}
}

// Validate implements Validator.
func (cv *CompositeValidator) Validate(name, value string) error {
	for _, v := range cv.validators {
		if err := v.Validate(name, value); err != nil {
			return err
		}
	}
	return nil
}

// EnumValidator validates that a value is one of the allowed values.
type EnumValidator struct {
	allowed map[string]bool
	values  []string
}

// NewEnum creates a validator that accepts only the listed values.
func NewEnum(values ...string) *EnumValidator {
	allowed := make(map[string]bool, len(values))
	for _, v := range values {
		allowed[v] = true
	}
	return &EnumValidator{allowed: allowed, values: values}
}

// Validate implements Validator.
func (ev *EnumValidator) Validate(name, value string) error {
	if !ev.allowed[value] {
		return fmt.Errorf("parameter %s value %q is not one of: %s", name, value, strings.Join(ev.values, ", "))
	}
	return nil
}

// NotEmptyValidator validates that a value is not blank.
type NotEmptyValidator struct{}

// Validate implements Validator.
func (nv *NotEmptyValidator) Validate(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("parameter %s cannot be empty", name)
	}
	return nil
}

// CustomValidator wraps an arbitrary validation function.
type CustomValidator struct {
	fn func(name, value string) error
}

// NewCustom creates a validator from a function.
func NewCustom(fn func(name, value string) error) *CustomValidator {
	return &CustomValidator{fn: fn}
}

// NewCustomBool creates a validator from a boolean predicate plus an error message.
func NewCustomBool(fn func(value string) bool, message string) *CustomValidator {
	return &CustomValidator{
		fn: func(name, value string) error {
			if !fn(value) {
				return fmt.Errorf("parameter %s: %s", name, message)
			}
			return nil
		},
	}
}

// Validate implements Validator.
func (cv *CustomValidator) Validate(name, value string) error {
	return cv.fn(name, value)
}

// Predefined validators for common route parameter patterns.
var (
	// NotEmpty rejects blank values.
	NotEmpty = &NotEmptyValidator{}

	// UUID accepts UUID v4 strings (case-insensitive).
	UUID = MustRegex(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	// PositiveInt accepts positive integer strings (no leading zeros, no sign).
	PositiveInt = MustRegex(`^[1-9]\d*$`)

	// Slug accepts lowercase URL slugs (e.g. "my-blog-post").
	Slug = MustRegex(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

	// Alphanumeric accepts strings containing only letters and digits.
	Alphanumeric = MustRegex(`^[a-zA-Z0-9]+$`)

	// Alpha accepts strings containing only letters.
	Alpha = MustRegex(`^[a-zA-Z]+$`)

	// Numeric accepts integer strings, including negative values.
	Numeric = MustRegex(`^-?\d+$`)

	// HexColor accepts hex color codes with or without a leading #.
	HexColor = MustRegex(`^#?([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

	// IPv4 accepts dotted-decimal IPv4 addresses.
	IPv4 = MustRegex(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)

	// Date accepts ISO 8601 date strings (YYYY-MM-DD).
	Date = MustRegex(`^\d{4}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12]\d|3[01])$`)
)
