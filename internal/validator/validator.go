package validator

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
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

	// New validation rules
	registry.Register("int", Int())
	registry.Register("float", Float())
	registry.Register("bool", Bool())
	registry.Register("uuid", UUID())
	registry.Register("ipv4", IPv4())
	registry.Register("ipv6", IPv6())
	registry.Register("ip", IP())
	registry.Register("mac", MAC())
	registry.Register("hex", Hex())
	registry.Register("base64", Base64())
	registry.Register("date", Date())
	registry.Register("time", Time())
	registry.Register("datetime", DateTime())
	registry.Register("array", Array())
	registry.Register("object", Object())

	// Security-focused validation rules
	registry.Register("secureEmail", SecureEmail())
	registry.Register("secureUrl", SecureURL())
	registry.Register("securePhone", SecurePhone())

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

func Validate(data any) error {
	validator := NewValidator(nil)
	return validator.Validate(data)
}

