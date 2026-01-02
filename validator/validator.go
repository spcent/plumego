package validator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

// ValidationError represents a structured validation error
type ValidationError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s (%s)", e.Field, e.Message, e.Code)
}

// FieldErrors represents a collection of field validation errors
type FieldErrors struct {
	errors []ValidationError
}

func (fe FieldErrors) Error() string {
	var msgs []string
	for _, err := range fe.errors {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

func (fe FieldErrors) Errors() []ValidationError {
	return append([]ValidationError(nil), fe.errors...)
}

// Rule interface for validation rules
type Rule interface {
	// Validate validates a value and returns ValidationError or nil
	Validate(value any) *ValidationError
}

// RuleFunc type alias for simple validation functions
type RuleFunc func(value any) *ValidationError

// Validate implements Rule interface for RuleFunc
func (f RuleFunc) Validate(value any) *ValidationError {
	return f(value)
}

// RuleRegistry manages validation rules
type RuleRegistry struct {
	rules map[string]Rule
	mu    sync.RWMutex
}

var (
	// DefaultRuleRegistry is the global rule registry
	DefaultRuleRegistry = NewRuleRegistry()
)

// NewRuleRegistry creates a new RuleRegistry with built-in rules
func NewRuleRegistry() *RuleRegistry {
	registry := &RuleRegistry{
		rules: make(map[string]Rule),
	}

	// Register built-in rules
	registry.Register("required", Required())
	registry.Register("email", Email())
	registry.Register("min", Min(0))
	registry.Register("max", Max(0))
	registry.Register("minLength", MinLength(0))
	registry.Register("maxLength", MaxLength(0))
	registry.Register("numeric", Numeric())
	registry.Register("alpha", Alpha())
	registry.Register("alphaNum", AlphaNum())
	registry.Register("url", URL())
	registry.Register("phone", Phone())
	registry.Register("regex", Regex(""))

	return registry
}

// Register registers a new validation rule
func (r *RuleRegistry) Register(name string, rule Rule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules[name] = rule
}

// Get retrieves a validation rule by name
func (r *RuleRegistry) Get(name string) (Rule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rule, exists := r.rules[name]
	return rule, exists
}

// Built-in validation rules

// Required ensures the value is non-nil and non-empty
func Required() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return &ValidationError{Code: "required", Message: "field is required"}
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return &ValidationError{Code: "required", Message: "field is required"}
			}
			rv = rv.Elem()
		}

		switch rv.Kind() {
		case reflect.String:
			if strings.TrimSpace(rv.String()) == "" {
				return &ValidationError{Code: "required", Message: "field is required"}
			}
		case reflect.Slice, reflect.Map, reflect.Array, reflect.Chan:
			if rv.Len() == 0 {
				return &ValidationError{Code: "required", Message: "field is required"}
			}
		default:
			if rv.IsZero() {
				return &ValidationError{Code: "required", Message: "field is required"}
			}
		}
		return nil
	})
}

// Email performs email format validation
func Email() Rule {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "email", Message: "must be a string"}
		}

		str = strings.TrimSpace(str)
		if str == "" {
			return nil
		}

		if !emailRegex.MatchString(str) {
			return &ValidationError{Code: "email", Message: "invalid email format"}
		}

		return nil
	})
}

// Min validates minimum numeric value or string length
func Min(min int64) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil // Skip validation for nil values unless required
		}

		switch v := value.(type) {
		case int:
			if int64(v) < min {
				return &ValidationError{Code: "min", Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case int8:
			if int64(v) < min {
				return &ValidationError{Code: "min", Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case int16:
			if int64(v) < min {
				return &ValidationError{Code: "min", Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case int32:
			if int64(v) < min {
				return &ValidationError{Code: "min", Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case int64:
			if v < min {
				return &ValidationError{Code: "min", Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case uint:
			if min > 0 && uint64(v) < uint64(min) {
				return &ValidationError{Code: "min", Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case uint8:
			if min > 0 && uint64(v) < uint64(min) {
				return &ValidationError{Code: "min", Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case uint16:
			if min > 0 && uint64(v) < uint64(min) {
				return &ValidationError{Code: "min", Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case uint32:
			if min > 0 && uint64(v) < uint64(min) {
				return &ValidationError{Code: "min", Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case uint64:
			if min > 0 && v < uint64(min) {
				return &ValidationError{Code: "min", Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case float32:
			if float64(v) < float64(min) {
				return &ValidationError{Code: "min", Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case float64:
			if v < float64(min) {
				return &ValidationError{Code: "min", Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case string:
			trimmed := strings.TrimSpace(v)
			if parsed, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
				if parsed < min {
					return &ValidationError{Code: "min", Message: fmt.Sprintf("must be at least %d", min)}
				}
				return nil
			} else if parsed, err := strconv.ParseUint(trimmed, 10, 64); err == nil {
				if min > 0 && parsed < uint64(min) {
					return &ValidationError{Code: "min", Message: fmt.Sprintf("must be at least %d", min)}
				}
				return nil
			} else if parsed, err := strconv.ParseFloat(trimmed, 64); err == nil {
				if parsed < float64(min) {
					return &ValidationError{Code: "min", Message: fmt.Sprintf("must be at least %d", min)}
				}
				return nil
			} else {
				// For non-numeric strings, check length for backward compatibility
				if int64(utf8.RuneCountInString(v)) < min {
					return &ValidationError{Code: "min", Message: fmt.Sprintf("must be at least %d characters", min)}
				}
			}
		}

		return nil
	})
}

// Max validates maximum numeric value or string length
func Max(max int64) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		switch v := value.(type) {
		case int:
			if int64(v) > max {
				return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case int8:
			if int64(v) > max {
				return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case int16:
			if int64(v) > max {
				return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case int32:
			if int64(v) > max {
				return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case int64:
			if v > max {
				return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case uint:
			if max < 0 {
				return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
			}
			if uint64(v) > uint64(max) {
				return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case uint8:
			if max < 0 {
				return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
			}
			if uint64(v) > uint64(max) {
				return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case uint16:
			if max < 0 {
				return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
			}
			if uint64(v) > uint64(max) {
				return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case uint32:
			if max < 0 {
				return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
			}
			if uint64(v) > uint64(max) {
				return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case uint64:
			if max < 0 {
				return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
			}
			if v > uint64(max) {
				return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case float32:
			if float64(v) > float64(max) {
				return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case float64:
			if v > float64(max) {
				return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case string:
			trimmed := strings.TrimSpace(v)
			if parsed, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
				if parsed > max {
					return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
				}
				return nil
			} else if parsed, err := strconv.ParseUint(trimmed, 10, 64); err == nil {
				if max < 0 || parsed > uint64(max) {
					return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
				}
				return nil
			} else if parsed, err := strconv.ParseFloat(trimmed, 64); err == nil {
				if parsed > float64(max) {
					return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d", max)}
				}
				return nil
			} else {
				// For non-numeric strings, check length for backward compatibility
				if int64(utf8.RuneCountInString(v)) > max {
					return &ValidationError{Code: "max", Message: fmt.Sprintf("must be at most %d characters", max)}
				}
			}
		}

		return nil
	})
}

// MinLength validates minimum string length
func MinLength(min int) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return nil
		}

		if utf8.RuneCountInString(str) < min {
			return &ValidationError{Code: "minLength", Message: fmt.Sprintf("must be at least %d characters", min)}
		}

		return nil
	})
}

// MaxLength validates maximum string length
func MaxLength(max int) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return nil
		}

		if utf8.RuneCountInString(str) > max {
			return &ValidationError{Code: "maxLength", Message: fmt.Sprintf("must be at most %d characters", max)}
		}

		return nil
	})
}

// Numeric validates that value contains only digits
func Numeric() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "numeric", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if _, err := strconv.ParseFloat(str, 64); err != nil {
			return &ValidationError{Code: "numeric", Message: "must be a valid number"}
		}

		return nil
	})
}

// Alpha validates that string contains only alphabetic characters
func Alpha() Rule {
	alphaRegex := regexp.MustCompile(`^[a-zA-Z]+$`)

	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "alpha", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !alphaRegex.MatchString(str) {
			return &ValidationError{Code: "alpha", Message: "must contain only alphabetic characters"}
		}

		return nil
	})
}

// AlphaNum validates that string contains only alphanumeric characters
func AlphaNum() Rule {
	alphaNumRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)

	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "alphaNum", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !alphaNumRegex.MatchString(str) {
			return &ValidationError{Code: "alphaNum", Message: "must contain only alphanumeric characters"}
		}

		return nil
	})
}

// URL validates that string is a valid URL
func URL() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "url", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if _, err := url.ParseRequestURI(str); err != nil {
			return &ValidationError{Code: "url", Message: "must be a valid URL"}
		}

		return nil
	})
}

// Phone validates that string is a valid phone number (basic validation)
func Phone() Rule {
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "phone", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		// Remove common phone number characters
		cleaned := regexp.MustCompile(`[\s\-\(\)\.]`).ReplaceAllString(str, "")

		if !phoneRegex.MatchString(cleaned) {
			return &ValidationError{Code: "phone", Message: "must be a valid phone number"}
		}

		return nil
	})
}

// Regex validates that string matches the given regular expression
func Regex(pattern string) Rule {
	if pattern == "" {
		// Return a rule that always passes
		return RuleFunc(func(value any) *ValidationError {
			return nil
		})
	}

	regex := regexp.MustCompile(pattern)

	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "regex", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !regex.MatchString(str) {
			return &ValidationError{Code: "regex", Message: "does not match required pattern"}
		}

		return nil
	})
}

// Validator holds validation configuration
type Validator struct {
	registry *RuleRegistry
	cache    map[string][][]Rule // Cache for parsed validation rules per field
	mu       sync.RWMutex
}

// NewValidator creates a new validator with custom registry
func NewValidator(registry *RuleRegistry) *Validator {
	if registry == nil {
		registry = DefaultRuleRegistry
	}

	return &Validator{
		registry: registry,
		cache:    make(map[string][][]Rule),
	}
}

// Validate applies validation rules defined via struct tags
func (v *Validator) Validate(data any) error {
	if data == nil {
		return &ValidationError{Code: "data", Message: "validator: nil data"}
	}

	typ := reflect.TypeOf(data)
	val := reflect.ValueOf(data)

	// Handle pointers
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return &ValidationError{Code: "data", Message: "validator: nil data"}
		}
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		return &ValidationError{Code: "data", Message: "validator: expected struct"}
	}

	var validationErrors FieldErrors

	// Cache key based on struct type and tag content
	cacheKey := v.getCacheKey(typ)

	v.mu.RLock()
	rules, exists := v.cache[cacheKey]
	v.mu.RUnlock()

	if !exists {
		v.mu.Lock()
		// Double check pattern for thread safety
		if rules, exists = v.cache[cacheKey]; !exists {
			rules = v.parseValidationRules(typ)
			v.cache[cacheKey] = rules
		}
		v.mu.Unlock()
	}

	// Apply validation rules
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !fieldType.IsExported() {
			continue
		}

		// Get field value (dereference pointers)
		value := field.Interface()
		if field.Kind() == reflect.Pointer && !field.IsNil() {
			value = field.Elem().Interface()
		}

		// Apply all rules for this field
		for _, rule := range rules[i] {
			if err := rule.Validate(value); err != nil {
				err.Field = fieldType.Name
				err.Value = value
				validationErrors.errors = append(validationErrors.errors, *err)
			}
		}
	}

	if len(validationErrors.errors) > 0 {
		return validationErrors
	}

	return nil
}

// getCacheKey generates a cache key for struct type
func (v *Validator) getCacheKey(typ reflect.Type) string {
	return typ.String()
}

// parseValidationRules parses validation rules from struct tags
func (v *Validator) parseValidationRules(typ reflect.Type) [][]Rule {
	rules := make([][]Rule, typ.NumField())

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("validate")

		if tag == "" {
			continue
		}

		ruleSpecs := strings.Split(tag, ",")
		fieldRules := make([]Rule, 0, len(ruleSpecs))

		for _, spec := range ruleSpecs {
			spec = strings.TrimSpace(spec)
			if spec == "" {
				continue
			}

			if rule := v.parseRule(spec); rule != nil {
				fieldRules = append(fieldRules, rule)
			}
		}

		rules[i] = fieldRules
	}

	return rules
}

// parseRule parses a single validation rule specification
func (v *Validator) parseRule(spec string) Rule {
	// Handle parameterized rules like min=10, max=100, regex=^[A-Z]+$
	parts := strings.SplitN(spec, "=", 2)
	ruleName := strings.TrimSpace(parts[0])

	var rule Rule
	var param string

	if len(parts) == 2 {
		param = strings.TrimSpace(parts[1])
	}

	switch ruleName {
	case "min":
		if val, err := strconv.Atoi(param); err == nil {
			rule = Min(int64(val))
		}
	case "max":
		if val, err := strconv.ParseInt(param, 10, 64); err == nil {
			rule = Max(val)
		}
	case "minLength":
		if val, err := strconv.Atoi(param); err == nil {
			rule = MinLength(val)
		}
	case "maxLength":
		if val, err := strconv.Atoi(param); err == nil {
			rule = MaxLength(val)
		}
	case "regex":
		rule = Regex(param)
	default:
		// Look up rule from registry
		if registeredRule, exists := v.registry.Get(ruleName); exists {
			rule = registeredRule
		}
	}

	return rule
}

// BindJSON decodes a JSON request body into v and applies validation
func (v *Validator) BindJSON(r *http.Request, target any) error {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return err
	}
	return v.Validate(target)
}

// Convenience functions for backward compatibility

// Validate applies validation rules using default validator
func Validate(data any) error {
	validator := NewValidator(nil)
	return validator.Validate(data)
}

// BindJSON decodes a JSON request body and applies validation using default validator
func BindJSON(r *http.Request, target any) error {
	validator := NewValidator(nil)
	return validator.BindJSON(r, target)
}
