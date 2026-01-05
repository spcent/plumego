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

	// UUIDValidator validates UUID format
	UUIDValidator = &RegexValidator{
		pattern: regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
	}

	// PositiveIntValidator validates positive integers
	PositiveIntValidator = &RegexValidator{
		pattern: regexp.MustCompile(`^[1-9]\d*$`),
	}
)

// WithValidation creates a router option that enables route validation
func WithValidation(validations map[string]*RouteValidation) RouterOption {
	return func(r *Router) {
		r.routeValidations = validations
	}
}

// AddValidation adds a validation rule for a specific route
func (r *Router) AddValidation(method, path string, validation *RouteValidation) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.routeValidations == nil {
		r.routeValidations = make(map[string]*RouteValidation)
	}

	// Normalize the path for consistent lookup
	normalizedPath := strings.TrimRight(path, "/")
	if normalizedPath == "" {
		normalizedPath = "/"
	}

	key := method + " " + r.prefix + normalizedPath
	r.routeValidations[key] = validation
}

// validateRouteParams validates route parameters against registered validations
func (r *Router) validateRouteParams(method, path string, params map[string]string) error {
	if r.routeValidations == nil {
		return nil
	}

	// Normalize the path for consistent lookup
	normalizedPath := strings.TrimRight(path, "/")
	if normalizedPath == "" {
		normalizedPath = "/"
	}

	// Try exact match first
	key := method + " " + normalizedPath
	if validation, exists := r.routeValidations[key]; exists {
		return validation.Validate(params)
	}

	// Try to find validation for parameterized paths
	// We need to check all registered validations and see if any match the current path pattern
	for validationKey, validation := range r.routeValidations {
		if !strings.HasPrefix(validationKey, method+" ") {
			continue
		}

		validationPath := strings.TrimPrefix(validationKey, method+" ")

		// Check if this validation path matches the current path
		// This is a simple pattern matching - in production you'd want something more robust
		if r.pathMatchesPattern(normalizedPath, validationPath, params) {
			return validation.Validate(params)
		}
	}

	return nil
}

// pathMatchesPattern checks if a concrete path matches a pattern with parameters
func (r *Router) pathMatchesPattern(path, pattern string, params map[string]string) bool {
	// If pattern has no parameters, do exact match
	if !strings.Contains(pattern, ":") && !strings.Contains(pattern, "*") {
		return path == pattern
	}

	// Split both paths
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")

	if len(pathParts) != len(patternParts) {
		return false
	}

	// Check each part
	for i, patternPart := range patternParts {
		pathPart := pathParts[i]

		if strings.HasPrefix(patternPart, ":") {
			// Parameter - should match the value in params
			paramName := patternPart[1:]
			if params[paramName] != pathPart {
				return false
			}
		} else if strings.HasPrefix(patternPart, "*") {
			// Wildcard - matches everything
			return true
		} else {
			// Static part - must match exactly
			if patternPart != pathPart {
				return false
			}
		}
	}

	return true
}
