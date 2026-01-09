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

type validationEntry struct {
	pattern    string
	validation *RouteValidation
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
		r.validationIndex = buildValidationIndex(validations)
	}
}

// AddValidation adds a validation rule for a specific route
func (r *Router) AddValidation(method, path string, validation *RouteValidation) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.routeValidations == nil {
		r.routeValidations = make(map[string]*RouteValidation)
	}
	if r.validationIndex == nil {
		r.validationIndex = make(map[string][]validationEntry)
	}

	// Normalize the path for consistent lookup
	normalizedPath := strings.TrimRight(path, "/")
	if normalizedPath == "" {
		normalizedPath = "/"
	}

	key := method + " " + r.prefix + normalizedPath
	r.routeValidations[key] = validation

	fullPath := r.prefix + normalizedPath
	if strings.Contains(fullPath, ":") || strings.Contains(fullPath, "*") {
		entries := r.validationIndex[method]
		for i := range entries {
			if entries[i].pattern == fullPath {
				entries[i].validation = validation
				r.validationIndex[method] = entries
				return
			}
		}
		r.validationIndex[method] = append(entries, validationEntry{
			pattern:    fullPath,
			validation: validation,
		})
	}
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
	if entries, exists := r.validationIndex[method]; exists {
		for _, entry := range entries {
			// This is a simple pattern matching - in production you'd want something more robust
			if r.pathMatchesPattern(normalizedPath, entry.pattern, params) {
				return entry.validation.Validate(params)
			}
		}
	}

	return nil
}

func buildValidationIndex(validations map[string]*RouteValidation) map[string][]validationEntry {
	index := make(map[string][]validationEntry)
	for key, validation := range validations {
		parts := strings.SplitN(key, " ", 2)
		if len(parts) != 2 {
			continue
		}
		method := parts[0]
		path := parts[1]
		if strings.Contains(path, ":") || strings.Contains(path, "*") {
			index[method] = append(index[method], validationEntry{
				pattern:    path,
				validation: validation,
			})
		}
	}
	return index
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

	wildIndex := -1
	for i, part := range patternParts {
		if strings.HasPrefix(part, "*") {
			wildIndex = i
			break
		}
	}

	if wildIndex == -1 {
		if len(pathParts) != len(patternParts) {
			return false
		}
	} else if len(pathParts) < wildIndex+1 {
		return false
	}

	// Check each part
	for i, patternPart := range patternParts {
		if strings.HasPrefix(patternPart, "*") {
			return true
		}
		if i >= len(pathParts) {
			return false
		}

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
