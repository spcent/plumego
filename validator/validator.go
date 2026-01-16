package validator

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
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

// Int validates that value is an integer
func Int() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		switch v := value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return nil
		case string:
			if v == "" {
				return nil
			}
			if _, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err != nil {
				return &ValidationError{Code: "int", Message: "must be an integer"}
			}
			return nil
		default:
			return &ValidationError{Code: "int", Message: "must be an integer"}
		}
	})
}

// Float validates that value is a float
func Float() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		switch v := value.(type) {
		case float32, float64:
			return nil
		case string:
			if v == "" {
				return nil
			}
			if _, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err != nil {
				return &ValidationError{Code: "float", Message: "must be a float"}
			}
			return nil
		default:
			return &ValidationError{Code: "float", Message: "must be a float"}
		}
	})
}

// Bool validates that value is a boolean
func Bool() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		switch v := value.(type) {
		case bool:
			return nil
		case string:
			if v == "" {
				return nil
			}
			lower := strings.ToLower(v)
			if lower == "true" || lower == "false" || lower == "1" || lower == "0" {
				return nil
			}
			return &ValidationError{Code: "bool", Message: "must be a boolean"}
		default:
			return &ValidationError{Code: "bool", Message: "must be a boolean"}
		}
	})
}

// UUID validates that string is a valid UUID
func UUID() Rule {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "uuid", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !uuidRegex.MatchString(strings.ToLower(str)) {
			return &ValidationError{Code: "uuid", Message: "must be a valid UUID"}
		}

		return nil
	})
}

// IPv4 validates that string is a valid IPv4 address
func IPv4() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "ipv4", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		ip := net.ParseIP(str)
		if ip == nil || ip.To4() == nil {
			return &ValidationError{Code: "ipv4", Message: "must be a valid IPv4 address"}
		}

		return nil
	})
}

// IPv6 validates that string is a valid IPv6 address
func IPv6() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "ipv6", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		ip := net.ParseIP(str)
		// IPv6 addresses contain colons, IPv4 addresses don't
		if ip == nil || !strings.Contains(str, ":") {
			return &ValidationError{Code: "ipv6", Message: "must be a valid IPv6 address"}
		}

		return nil
	})
}

// IP validates that string is a valid IP address (IPv4 or IPv6)
func IP() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "ip", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		ip := net.ParseIP(str)
		if ip == nil {
			return &ValidationError{Code: "ip", Message: "must be a valid IP address"}
		}

		return nil
	})
}

// MAC validates that string is a valid MAC address
func MAC() Rule {
	macRegex := regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)

	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "mac", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !macRegex.MatchString(str) {
			return &ValidationError{Code: "mac", Message: "must be a valid MAC address"}
		}

		return nil
	})
}

// Hex validates that string contains only hexadecimal characters
func Hex() Rule {
	hexRegex := regexp.MustCompile(`^[0-9A-Fa-f]+$`)

	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "hex", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !hexRegex.MatchString(str) {
			return &ValidationError{Code: "hex", Message: "must be a valid hexadecimal string"}
		}

		return nil
	})
}

// Base64 validates that string is a valid Base64 encoded string
func Base64() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "base64", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		// Try standard encoding first
		if _, err := base64.StdEncoding.DecodeString(str); err == nil {
			return nil
		}

		// Try URL encoding (no padding)
		if _, err := base64.RawURLEncoding.DecodeString(str); err == nil {
			return nil
		}

		return &ValidationError{Code: "base64", Message: "must be a valid Base64 string"}
	})
}

// Date validates that string is a valid date (YYYY-MM-DD)
func Date() Rule {
	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "date", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !dateRegex.MatchString(str) {
			return &ValidationError{Code: "date", Message: "must be a valid date (YYYY-MM-DD)"}
		}

		if _, err := time.Parse("2006-01-02", str); err != nil {
			return &ValidationError{Code: "date", Message: "must be a valid date"}
		}

		return nil
	})
}

// Time validates that string is a valid time (HH:MM:SS or HH:MM)
func Time() Rule {
	timeRegex := regexp.MustCompile(`^\d{2}:\d{2}(:\d{2})?$`)

	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "time", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !timeRegex.MatchString(str) {
			return &ValidationError{Code: "time", Message: "must be a valid time (HH:MM:SS or HH:MM)"}
		}

		if _, err := time.Parse("15:04:05", str); err != nil {
			if _, err := time.Parse("15:04", str); err != nil {
				return &ValidationError{Code: "time", Message: "must be a valid time"}
			}
		}

		return nil
	})
}

// DateTime validates that string is a valid datetime (YYYY-MM-DD HH:MM:SS or ISO8601)
func DateTime() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "datetime", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		// Try common datetime formats
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02T15:04:05.999Z07:00",
			time.RFC3339,
			time.RFC3339Nano,
		}

		for _, format := range formats {
			if _, err := time.Parse(format, str); err == nil {
				return nil
			}
		}

		return &ValidationError{Code: "datetime", Message: "must be a valid datetime"}
	})
}

// Array validates that value is an array or slice
func Array() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil
			}
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return &ValidationError{Code: "array", Message: "must be an array or slice"}
		}

		return nil
	})
}

// Object validates that value is a struct or map
func Object() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil
			}
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Struct && rv.Kind() != reflect.Map {
			return &ValidationError{Code: "object", Message: "must be an object"}
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

// Advanced validation rules

// Range validates that value is within a range [min, max]
func Range(min, max int64) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		switch v := value.(type) {
		case int:
			if int64(v) < min || int64(v) > max {
				return &ValidationError{Code: "range", Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case int8:
			if int64(v) < min || int64(v) > max {
				return &ValidationError{Code: "range", Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case int16:
			if int64(v) < min || int64(v) > max {
				return &ValidationError{Code: "range", Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case int32:
			if int64(v) < min || int64(v) > max {
				return &ValidationError{Code: "range", Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case int64:
			if v < min || v > max {
				return &ValidationError{Code: "range", Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case uint:
			if min > 0 && (uint64(v) < uint64(min) || uint64(v) > uint64(max)) {
				return &ValidationError{Code: "range", Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case uint8:
			if min > 0 && (uint64(v) < uint64(min) || uint64(v) > uint64(max)) {
				return &ValidationError{Code: "range", Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case uint16:
			if min > 0 && (uint64(v) < uint64(min) || uint64(v) > uint64(max)) {
				return &ValidationError{Code: "range", Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case uint32:
			if min > 0 && (uint64(v) < uint64(min) || uint64(v) > uint64(max)) {
				return &ValidationError{Code: "range", Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case uint64:
			if min > 0 && (v < uint64(min) || v > uint64(max)) {
				return &ValidationError{Code: "range", Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case float32:
			if float64(v) < float64(min) || float64(v) > float64(max) {
				return &ValidationError{Code: "range", Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case float64:
			if v < float64(min) || v > float64(max) {
				return &ValidationError{Code: "range", Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case string:
			trimmed := strings.TrimSpace(v)
			if parsed, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
				if parsed < min || parsed > max {
					return &ValidationError{Code: "range", Message: fmt.Sprintf("must be between %d and %d", min, max)}
				}
				return nil
			} else if parsed, err := strconv.ParseUint(trimmed, 10, 64); err == nil {
				if min > 0 && (parsed < uint64(min) || parsed > uint64(max)) {
					return &ValidationError{Code: "range", Message: fmt.Sprintf("must be between %d and %d", min, max)}
				}
				return nil
			} else if parsed, err := strconv.ParseFloat(trimmed, 64); err == nil {
				if parsed < float64(min) || parsed > float64(max) {
					return &ValidationError{Code: "range", Message: fmt.Sprintf("must be between %d and %d", min, max)}
				}
				return nil
			}
		}

		return nil
	})
}

// MinFloat validates minimum float value
func MinFloat(min float64) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		switch v := value.(type) {
		case float32:
			if float64(v) < min {
				return &ValidationError{Code: "minFloat", Message: fmt.Sprintf("must be at least %f", min)}
			}
			return nil
		case float64:
			if v < min {
				return &ValidationError{Code: "minFloat", Message: fmt.Sprintf("must be at least %f", min)}
			}
			return nil
		case string:
			if v == "" {
				return nil
			}
			if parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
				if parsed < min {
					return &ValidationError{Code: "minFloat", Message: fmt.Sprintf("must be at least %f", min)}
				}
				return nil
			}
		}

		return nil
	})
}

// MaxFloat validates maximum float value
func MaxFloat(max float64) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		switch v := value.(type) {
		case float32:
			if float64(v) > max {
				return &ValidationError{Code: "maxFloat", Message: fmt.Sprintf("must be at most %f", max)}
			}
			return nil
		case float64:
			if v > max {
				return &ValidationError{Code: "maxFloat", Message: fmt.Sprintf("must be at most %f", max)}
			}
			return nil
		case string:
			if v == "" {
				return nil
			}
			if parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
				if parsed > max {
					return &ValidationError{Code: "maxFloat", Message: fmt.Sprintf("must be at most %f", max)}
				}
				return nil
			}
		}

		return nil
	})
}

// RangeFloat validates that float value is within a range [min, max]
func RangeFloat(min, max float64) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		switch v := value.(type) {
		case float32:
			if float64(v) < min || float64(v) > max {
				return &ValidationError{Code: "rangeFloat", Message: fmt.Sprintf("must be between %f and %f", min, max)}
			}
			return nil
		case float64:
			if v < min || v > max {
				return &ValidationError{Code: "rangeFloat", Message: fmt.Sprintf("must be between %f and %f", min, max)}
			}
			return nil
		case string:
			if v == "" {
				return nil
			}
			if parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
				if parsed < min || parsed > max {
					return &ValidationError{Code: "rangeFloat", Message: fmt.Sprintf("must be between %f and %f", min, max)}
				}
				return nil
			}
		}

		return nil
	})
}

// EmailList validates that string is a comma-separated list of valid emails
func EmailList() Rule {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "emailList", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		emails := strings.Split(str, ",")
		for _, email := range emails {
			email = strings.TrimSpace(email)
			if email == "" {
				continue
			}
			if !emailRegex.MatchString(email) {
				return &ValidationError{Code: "emailList", Message: "contains invalid email format"}
			}
		}

		return nil
	})
}

// URLList validates that string is a comma-separated list of valid URLs
func URLList() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "urlList", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		urls := strings.Split(str, ",")
		for _, urlStr := range urls {
			urlStr = strings.TrimSpace(urlStr)
			if urlStr == "" {
				continue
			}
			if _, err := url.ParseRequestURI(urlStr); err != nil {
				return &ValidationError{Code: "urlList", Message: "contains invalid URL"}
			}
		}

		return nil
	})
}

// OneOf validates that value is one of the allowed values
func OneOf(allowed ...string) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "oneOf", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		for _, allowedValue := range allowed {
			if str == allowedValue {
				return nil
			}
		}

		return &ValidationError{Code: "oneOf", Message: fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", "))}
	})
}

// NotEmpty validates that string is not empty
func NotEmpty() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return &ValidationError{Code: "notEmpty", Message: "field is required"}
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "notEmpty", Message: "must be a string"}
		}

		if str == "" {
			return &ValidationError{Code: "notEmpty", Message: "field cannot be empty"}
		}

		return nil
	})
}

// NotZero validates that value is not the zero value for its type
func NotZero() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return &ValidationError{Code: "notZero", Message: "field is required"}
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return &ValidationError{Code: "notZero", Message: "field is required"}
			}
			rv = rv.Elem()
		}

		if rv.IsZero() {
			return &ValidationError{Code: "notZero", Message: "field cannot be zero"}
		}

		return nil
	})
}

// AfterDate validates that date is after a specified date
func AfterDate(after string) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "afterDate", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		parsedAfter, err := time.Parse("2006-01-02", after)
		if err != nil {
			return &ValidationError{Code: "afterDate", Message: "invalid reference date"}
		}

		parsedValue, err := time.Parse("2006-01-02", str)
		if err != nil {
			return &ValidationError{Code: "afterDate", Message: "must be a valid date"}
		}

		if !parsedValue.After(parsedAfter) {
			return &ValidationError{Code: "afterDate", Message: fmt.Sprintf("must be after %s", after)}
		}

		return nil
	})
}

// BeforeDate validates that date is before a specified date
func BeforeDate(before string) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "beforeDate", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		parsedBefore, err := time.Parse("2006-01-02", before)
		if err != nil {
			return &ValidationError{Code: "beforeDate", Message: "invalid reference date"}
		}

		parsedValue, err := time.Parse("2006-01-02", str)
		if err != nil {
			return &ValidationError{Code: "beforeDate", Message: "must be a valid date"}
		}

		if !parsedValue.Before(parsedBefore) {
			return &ValidationError{Code: "beforeDate", Message: fmt.Sprintf("must be before %s", before)}
		}

		return nil
	})
}

// MinLengthBytes validates minimum byte length
func MinLengthBytes(min int) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return nil
		}

		if len(str) < min {
			return &ValidationError{Code: "minLengthBytes", Message: fmt.Sprintf("must be at least %d bytes", min)}
		}

		return nil
	})
}

// MaxLengthBytes validates maximum byte length
func MaxLengthBytes(max int) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return nil
		}

		if len(str) > max {
			return &ValidationError{Code: "maxLengthBytes", Message: fmt.Sprintf("must be at most %d bytes", max)}
		}

		return nil
	})
}

// Contains validates that string contains a substring
func Contains(substring string) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "contains", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !strings.Contains(str, substring) {
			return &ValidationError{Code: "contains", Message: fmt.Sprintf("must contain '%s'", substring)}
		}

		return nil
	})
}

// HasPrefix validates that string starts with a prefix
func HasPrefix(prefix string) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "hasPrefix", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !strings.HasPrefix(str, prefix) {
			return &ValidationError{Code: "hasPrefix", Message: fmt.Sprintf("must start with '%s'", prefix)}
		}

		return nil
	})
}

// HasSuffix validates that string ends with a suffix
func HasSuffix(suffix string) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "hasSuffix", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !strings.HasSuffix(str, suffix) {
			return &ValidationError{Code: "hasSuffix", Message: fmt.Sprintf("must end with '%s'", suffix)}
		}

		return nil
	})
}

// CaseInsensitive validates that string matches pattern case-insensitively
func CaseInsensitive(pattern string) Rule {
	if pattern == "" {
		return RuleFunc(func(value any) *ValidationError {
			return nil
		})
	}

	regex := regexp.MustCompile("(?i)" + pattern)

	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: "caseInsensitive", Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !regex.MatchString(str) {
			return &ValidationError{Code: "caseInsensitive", Message: "does not match required pattern"}
		}

		return nil
	})
}

// MinItems validates minimum array/slice length
func MinItems(min int) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil
			}
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return nil
		}

		if rv.Len() < min {
			return &ValidationError{Code: "minItems", Message: fmt.Sprintf("must have at least %d items", min)}
		}

		return nil
	})
}

// MaxItems validates maximum array/slice length
func MaxItems(max int) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil
			}
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return nil
		}

		if rv.Len() > max {
			return &ValidationError{Code: "maxItems", Message: fmt.Sprintf("must have at most %d items", max)}
		}

		return nil
	})
}

// Unique validates that array/slice contains unique elements
func Unique() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil
			}
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return nil
		}

		seen := make(map[interface{}]bool)
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i).Interface()
			if seen[elem] {
				return &ValidationError{Code: "unique", Message: "array must contain unique elements"}
			}
			seen[elem] = true
		}

		return nil
	})
}

// MinMapKeys validates minimum map keys count
func MinMapKeys(min int) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil
			}
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Map {
			return nil
		}

		if rv.Len() < min {
			return &ValidationError{Code: "minMapKeys", Message: fmt.Sprintf("must have at least %d keys", min)}
		}

		return nil
	})
}

// MaxMapKeys validates maximum map keys count
func MaxMapKeys(max int) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil
			}
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Map {
			return nil
		}

		if rv.Len() > max {
			return &ValidationError{Code: "maxMapKeys", Message: fmt.Sprintf("must have at most %d keys", max)}
		}

		return nil
	})
}

// CustomRule creates a custom validation rule from a function
func CustomRule(name string, fn func(value any) bool) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		if !fn(value) {
			return &ValidationError{Code: name, Message: fmt.Sprintf("failed custom validation: %s", name)}
		}

		return nil
	})
}

// Optional marks a rule as optional (allows nil or empty values)
func Optional(rule Rule) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if ok && str == "" {
			return nil
		}

		return rule.Validate(value)
	})
}

// WithMessage creates a rule with a custom error message
func WithMessage(rule Rule, message string) Rule {
	return RuleFunc(func(value any) *ValidationError {
		err := rule.Validate(value)
		if err != nil {
			err.Message = message
		}
		return err
	})
}

// WithCode creates a rule with a custom error code
func WithCode(rule Rule, code string) Rule {
	return RuleFunc(func(value any) *ValidationError {
		err := rule.Validate(value)
		if err != nil {
			err.Code = code
		}
		return err
	})
}
