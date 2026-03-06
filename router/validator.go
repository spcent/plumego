package router

import (
	"fmt"
	"regexp"
	"strings"
)

// ParamValidator defines the interface for route parameter validation
type ParamValidator interface {
	Validate(name, value string) error
}

// RegexValidator validates parameters using regular expressions
type RegexValidator struct {
	pattern *regexp.Regexp
}

// NewRegexValidator creates a new regex validator
func NewRegexValidator(pattern string) (*RegexValidator, error) {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &RegexValidator{pattern: compiled}, nil
}

// Validate checks if the value matches the regex pattern
func (rv *RegexValidator) Validate(name, value string) error {
	if !rv.pattern.MatchString(value) {
		return fmt.Errorf("parameter %s value %s does not match pattern %s", name, value, rv.pattern.String())
	}
	return nil
}

// RangeValidator validates numeric parameters within a range
type RangeValidator struct {
	min, max int64
}

// NewRangeValidator creates a new range validator
func NewRangeValidator(min, max int64) *RangeValidator {
	return &RangeValidator{min: min, max: max}
}

// Validate checks if the value is within the specified range
func (rv *RangeValidator) Validate(name, value string) error {
	var num int64
	_, err := fmt.Sscanf(value, "%d", &num)
	if err != nil {
		return fmt.Errorf("parameter %s value %s is not a valid integer", name, value)
	}
	if num < rv.min || num > rv.max {
		return fmt.Errorf("parameter %s value %d is not in range [%d, %d]", name, num, rv.min, rv.max)
	}
	return nil
}

// LengthValidator validates string length
type LengthValidator struct {
	min, max int
}

// NewLengthValidator creates a new length validator
func NewLengthValidator(min, max int) *LengthValidator {
	return &LengthValidator{min: min, max: max}
}

// Validate checks if the string length is within bounds
func (lv *LengthValidator) Validate(name, value string) error {
	length := len(value)
	if length < lv.min || length > lv.max {
		return fmt.Errorf("parameter %s length %d is not in range [%d, %d]", name, length, lv.min, lv.max)
	}
	return nil
}

// RouteValidation represents validation rules for a route
type RouteValidation struct {
	Params map[string]ParamValidator
}

// NewRouteValidation creates a new route validation
func NewRouteValidation() *RouteValidation {
	return &RouteValidation{
		Params: make(map[string]ParamValidator),
	}
}

// AddParam adds a parameter validator
func (rv *RouteValidation) AddParam(name string, validator ParamValidator) *RouteValidation {
	rv.Params[name] = validator
	return rv
}

// Validate validates all parameters
func (rv *RouteValidation) Validate(params map[string]string) error {
	for name, value := range params {
		if validator, exists := rv.Params[name]; exists {
			if err := validator.Validate(name, value); err != nil {
				return err
			}
		}
	}
	return nil
}

// Predefined validators
var (
	// UserIDValidator validates user IDs (alphanumeric, 1-20 chars)
	UserIDValidator = NewLengthValidator(1, 20)

	// EmailValidator validates email addresses
	EmailValidator = &RegexValidator{
		pattern: regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
	}

	// UUIDValidator validates UUID format (case insensitive)
	UUIDValidator = &RegexValidator{
		pattern: regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
	}

	// PositiveIntValidator validates positive integers
	PositiveIntValidator = &RegexValidator{
		pattern: regexp.MustCompile(`^[1-9]\d*$`),
	}

	// SlugValidator validates URL slugs (lowercase alphanumeric with hyphens)
	SlugValidator = &RegexValidator{
		pattern: regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`),
	}

	// AlphanumericValidator validates alphanumeric strings
	AlphanumericValidator = &RegexValidator{
		pattern: regexp.MustCompile(`^[a-zA-Z0-9]+$`),
	}

	// AlphaValidator validates alphabetic strings only
	AlphaValidator = &RegexValidator{
		pattern: regexp.MustCompile(`^[a-zA-Z]+$`),
	}

	// NumericValidator validates numeric strings (including negative and decimals)
	NumericValidator = &RegexValidator{
		pattern: regexp.MustCompile(`^-?\d+(\.\d+)?$`),
	}

	// IntegerValidator validates integer strings (including negative)
	IntegerValidator = &RegexValidator{
		pattern: regexp.MustCompile(`^-?\d+$`),
	}

	// HexColorValidator validates hex color codes (with or without #)
	HexColorValidator = &RegexValidator{
		pattern: regexp.MustCompile(`^#?([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`),
	}

	// IPv4Validator validates IPv4 addresses
	IPv4Validator = &RegexValidator{
		pattern: regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`),
	}

	// IPv6Validator validates IPv6 addresses (simplified)
	IPv6Validator = &RegexValidator{
		pattern: regexp.MustCompile(`^(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$|^::(?:[0-9a-fA-F]{1,4}:){0,6}[0-9a-fA-F]{1,4}$|^(?:[0-9a-fA-F]{1,4}:){1,7}:$`),
	}

	// DateValidator validates ISO 8601 date format (YYYY-MM-DD)
	DateValidator = &RegexValidator{
		pattern: regexp.MustCompile(`^\d{4}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12]\d|3[01])$`),
	}

	// TimeValidator validates time format (HH:MM or HH:MM:SS)
	TimeValidator = &RegexValidator{
		pattern: regexp.MustCompile(`^(?:[01]\d|2[0-3]):[0-5]\d(?::[0-5]\d)?$`),
	}

	// DateTimeValidator validates ISO 8601 datetime format
	DateTimeValidator = &RegexValidator{
		pattern: regexp.MustCompile(`^\d{4}-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12]\d|3[01])T(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d(?:\.\d+)?(?:Z|[+-](?:[01]\d|2[0-3]):[0-5]\d)?$`),
	}

	// PhoneValidator validates phone numbers (international format)
	PhoneValidator = &RegexValidator{
		pattern: regexp.MustCompile(`^\+?[1-9]\d{1,14}$`),
	}

	// UsernameValidator validates usernames (alphanumeric with underscores, 3-20 chars)
	UsernameValidator = &CompositeValidator{
		validators: []ParamValidator{
			NewLengthValidator(3, 20),
			&RegexValidator{pattern: regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)},
		},
	}

	// Base64Validator validates base64 encoded strings
	Base64Validator = &RegexValidator{
		pattern: regexp.MustCompile(`^[A-Za-z0-9+/]*={0,2}$`),
	}

	// URLPathValidator validates URL path segments (no special characters)
	URLPathValidator = &RegexValidator{
		pattern: regexp.MustCompile(`^[a-zA-Z0-9._~:/?#\[\]@!$&'()*+,;=-]*$`),
	}
)

// CompositeValidator combines multiple validators
type CompositeValidator struct {
	validators []ParamValidator
}

// NewCompositeValidator creates a new composite validator
func NewCompositeValidator(validators ...ParamValidator) *CompositeValidator {
	return &CompositeValidator{validators: validators}
}

// Validate checks all validators in sequence
func (cv *CompositeValidator) Validate(name, value string) error {
	for _, v := range cv.validators {
		if err := v.Validate(name, value); err != nil {
			return err
		}
	}
	return nil
}

// EnumValidator validates that a value is one of the allowed values
type EnumValidator struct {
	allowed map[string]bool
	values  []string // For error message
}

// NewEnumValidator creates a new enum validator
func NewEnumValidator(values ...string) *EnumValidator {
	allowed := make(map[string]bool, len(values))
	for _, v := range values {
		allowed[v] = true
	}
	return &EnumValidator{allowed: allowed, values: values}
}

// Validate checks if the value is one of the allowed values
func (ev *EnumValidator) Validate(name, value string) error {
	if !ev.allowed[value] {
		return fmt.Errorf("parameter %s value %q is not one of: %s", name, value, strings.Join(ev.values, ", "))
	}
	return nil
}

// NotEmptyValidator validates that a value is not empty
type NotEmptyValidator struct{}

// Validate checks if the value is not empty
func (nv *NotEmptyValidator) Validate(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("parameter %s cannot be empty", name)
	}
	return nil
}

// NotEmpty is a predefined not-empty validator
var NotEmpty = &NotEmptyValidator{}

// CustomValidator wraps a custom validation function
type CustomValidator struct {
	fn      func(name, value string) error
	message string
}

// NewCustomValidator creates a new custom validator
func NewCustomValidator(fn func(name, value string) error) *CustomValidator {
	return &CustomValidator{fn: fn}
}

// NewCustomValidatorWithMessage creates a new custom validator with error message
func NewCustomValidatorWithMessage(fn func(value string) bool, message string) *CustomValidator {
	return &CustomValidator{
		fn: func(name, value string) error {
			if !fn(value) {
				return fmt.Errorf("parameter %s: %s", name, message)
			}
			return nil
		},
		message: message,
	}
}

// Validate runs the custom validation function
func (cv *CustomValidator) Validate(name, value string) error {
	return cv.fn(name, value)
}

// WithValidation creates a router option that enables route validation
func WithValidation(validations map[string]*RouteValidation) RouterOption {
	return func(r *Router) {
		r.state.routeValidations = make(map[string]map[string]*RouteValidation)
		for key, validation := range validations {
			parts := strings.SplitN(key, " ", 2)
			if len(parts) != 2 {
				continue
			}
			r.setValidation(parts[0], parts[1], validation)
		}
	}
}

// WithValidationRule creates a router option that adds a single validation rule
// for a specific route. This is a convenience function for adding individual
// validation rules without needing to create a full validation map.
//
// Parameters:
//   - method: HTTP method (GET, POST, PUT, DELETE, PATCH)
//   - path: URL path
//   - validation: Validation rules for the route
//
// Example:
//
//	router := NewRouter(
//	    WithValidationRule("POST", "/users", &RouteValidation{
//	        Headers: map[string]Validator{
//	            "Content-Type": StringValidator("application/json"),
//	        },
//	    }),
//	)
func WithValidationRule(method, path string, validation *RouteValidation) RouterOption {
	return func(r *Router) {
		r.setValidation(method, path, validation)
	}
}

// AddValidation adds a validation rule for a specific route
func (r *Router) AddValidation(method, path string, validation *RouteValidation) {
	r.state.mu.Lock()
	defer r.state.mu.Unlock()

	// Normalize the path for consistent lookup
	normalizedPath := strings.TrimRight(path, "/")
	if normalizedPath == "" {
		normalizedPath = "/"
	}

	pattern := r.prefix + normalizedPath
	r.setValidation(method, pattern, validation)

	if routeNode := r.findRouteNodeLocked(method, pattern); routeNode != nil {
		routeNode.validation = validation
	}
}
