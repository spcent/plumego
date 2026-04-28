package validator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// RouteParamValidator is the interface for route parameter validators.
type RouteParamValidator interface {
	Validate(name, value string) error
}

// RouteParamRegexValidator validates parameters using a regular expression.
type RouteParamRegexValidator struct {
	pattern *regexp.Regexp
}

// RouteParamNewRegex creates a validator that accepts values matching the given pattern.
func RouteParamNewRegex(pattern string) (*RouteParamRegexValidator, error) {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &RouteParamRegexValidator{pattern: compiled}, nil
}

// RouteParamMustRegex is like RouteParamNewRegex but panics if the pattern fails to compile.
func RouteParamMustRegex(pattern string) *RouteParamRegexValidator {
	v, err := RouteParamNewRegex(pattern)
	if err != nil {
		panic(fmt.Sprintf("validator: invalid regex pattern %q: %v", pattern, err))
	}
	return v
}

// Validate implements RouteParamValidator.
func (rv *RouteParamRegexValidator) Validate(name, value string) error {
	if !rv.pattern.MatchString(value) {
		return fmt.Errorf("parameter %s value %s does not match pattern %s", name, value, rv.pattern.String())
	}
	return nil
}

// RouteParamRangeValidator validates that a numeric parameter falls within [min, max].
type RouteParamRangeValidator struct {
	min, max int64
}

// RouteParamNewRange creates a validator that accepts integers in [min, max].
func RouteParamNewRange(min, max int64) *RouteParamRangeValidator {
	return &RouteParamRangeValidator{min: min, max: max}
}

// Validate implements RouteParamValidator.
func (rv *RouteParamRangeValidator) Validate(name, value string) error {
	if rv.min > rv.max {
		return fmt.Errorf("parameter %s has invalid validator range [%d, %d]", name, rv.min, rv.max)
	}

	num, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("parameter %s value %s is not a valid integer", name, value)
	}
	if num < rv.min || num > rv.max {
		return fmt.Errorf("parameter %s value %d is not in range [%d, %d]", name, num, rv.min, rv.max)
	}
	return nil
}

// RouteParamLengthValidator validates string length falls within [min, max].
type RouteParamLengthValidator struct {
	min, max int
}

// RouteParamNewLength creates a validator that accepts strings with length in [min, max].
func RouteParamNewLength(min, max int) *RouteParamLengthValidator {
	return &RouteParamLengthValidator{min: min, max: max}
}

// Validate implements RouteParamValidator.
func (lv *RouteParamLengthValidator) Validate(name, value string) error {
	if lv.min > lv.max {
		return fmt.Errorf("parameter %s has invalid validator length range [%d, %d]", name, lv.min, lv.max)
	}

	length := utf8.RuneCountInString(value)
	if length < lv.min || length > lv.max {
		return fmt.Errorf("parameter %s length %d is not in range [%d, %d]", name, length, lv.min, lv.max)
	}
	return nil
}

// RouteParamCompositeValidator applies multiple validators in sequence, failing on the first error.
type RouteParamCompositeValidator struct {
	validators []RouteParamValidator
}

// RouteParamNewComposite creates a validator that runs all given validators in order.
func RouteParamNewComposite(validators ...RouteParamValidator) *RouteParamCompositeValidator {
	return &RouteParamCompositeValidator{validators: validators}
}

// Validate implements RouteParamValidator.
func (cv *RouteParamCompositeValidator) Validate(name, value string) error {
	for _, v := range cv.validators {
		if v == nil {
			return fmt.Errorf("parameter %s has nil validator", name)
		}
		if err := v.Validate(name, value); err != nil {
			return err
		}
	}
	return nil
}

// RouteParamEnumValidator validates that a value is one of the allowed values.
type RouteParamEnumValidator struct {
	allowed map[string]bool
	values  []string
}

// RouteParamNewEnum creates a validator that accepts only the listed values.
func RouteParamNewEnum(values ...string) *RouteParamEnumValidator {
	allowed := make(map[string]bool, len(values))
	for _, v := range values {
		allowed[v] = true
	}
	return &RouteParamEnumValidator{allowed: allowed, values: values}
}

// Validate implements RouteParamValidator.
func (ev *RouteParamEnumValidator) Validate(name, value string) error {
	if !ev.allowed[value] {
		return fmt.Errorf("parameter %s value %q is not one of: %s", name, value, strings.Join(ev.values, ", "))
	}
	return nil
}

// RouteParamNotEmptyValidator validates that a value is not blank.
type RouteParamNotEmptyValidator struct{}

// Validate implements RouteParamValidator.
func (nv *RouteParamNotEmptyValidator) Validate(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("parameter %s cannot be empty", name)
	}
	return nil
}

// RouteParamCustomValidator wraps an arbitrary validation function.
type RouteParamCustomValidator struct {
	fn func(name, value string) error
}

// RouteParamNewCustom creates a validator from a function.
func RouteParamNewCustom(fn func(name, value string) error) *RouteParamCustomValidator {
	return &RouteParamCustomValidator{fn: fn}
}

// RouteParamNewCustomBool creates a validator from a boolean predicate plus an error message.
func RouteParamNewCustomBool(fn func(value string) bool, message string) *RouteParamCustomValidator {
	return &RouteParamCustomValidator{
		fn: func(name, value string) error {
			if !fn(value) {
				return fmt.Errorf("parameter %s: %s", name, message)
			}
			return nil
		},
	}
}

// Validate implements RouteParamValidator.
func (cv *RouteParamCustomValidator) Validate(name, value string) error {
	if cv.fn == nil {
		return fmt.Errorf("parameter %s has nil custom validator", name)
	}
	return cv.fn(name, value)
}

// Reusable route parameter validators for common patterns.
var (
	// RouteParamNotEmpty rejects blank values.
	RouteParamNotEmpty = &RouteParamNotEmptyValidator{}

	// RouteParamUUID accepts UUID strings (case-insensitive).
	RouteParamUUID = RouteParamMustRegex(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	// RouteParamPositiveInt accepts positive integer strings (no leading zeros, no sign).
	RouteParamPositiveInt = RouteParamMustRegex(`^[1-9]\d*$`)

	// RouteParamSlug accepts lowercase URL slugs (e.g. "my-blog-post").
	RouteParamSlug = RouteParamMustRegex(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

	// RouteParamAlphanumeric accepts strings containing only letters and digits.
	RouteParamAlphanumeric = RouteParamMustRegex(`^[a-zA-Z0-9]+$`)

	// RouteParamAlpha accepts strings containing only letters.
	RouteParamAlpha = RouteParamMustRegex(`^[a-zA-Z]+$`)

	// RouteParamNumeric accepts integer strings, including negative values.
	RouteParamNumeric = RouteParamMustRegex(`^-?\d+$`)

	// RouteParamHexColor accepts hex color codes with or without a leading #.
	RouteParamHexColor = RouteParamMustRegex(`^#?([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

	// RouteParamIPv4 accepts dotted-decimal IPv4 addresses.
	RouteParamIPv4 = RouteParamMustRegex(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)

	// RouteParamDate accepts ISO 8601 date strings (YYYY-MM-DD).
	RouteParamDate = RouteParamMustRegex(`^\d{4}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12]\d|3[01])$`)
)
