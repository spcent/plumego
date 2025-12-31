package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Validator interface for configuration validation
type Validator interface {
	Validate(value any, key string) error
	Name() string
}

// Required validator ensures a value is not empty
type Required struct{}

func (r *Required) Validate(value any, key string) error {
	if value == nil {
		return fmt.Errorf("config %s is required", key)
	}

	str, ok := value.(string)
	if ok && strings.TrimSpace(str) == "" {
		return fmt.Errorf("config %s is required and cannot be empty", key)
	}

	if str, ok := toString(value); ok && strings.TrimSpace(str) == "" {
		return fmt.Errorf("config %s is required and cannot be empty", key)
	}

	return nil
}

func (r *Required) Name() string {
	return "required"
}

// MinLength validator ensures string length is at least minimum
type MinLength struct {
	Min int
}

func (m *MinLength) Validate(value any, key string) error {
	if str, ok := toString(value); ok {
		if utf8.RuneCountInString(str) < m.Min {
			return fmt.Errorf("config %s must be at least %d characters, got %d", key, m.Min, utf8.RuneCountInString(str))
		}
	}
	return nil
}

func (m *MinLength) Name() string {
	return "min_length"
}

// MaxLength validator ensures string length is at most maximum
type MaxLength struct {
	Max int
}

func (m *MaxLength) Validate(value any, key string) error {
	if str, ok := toString(value); ok {
		if utf8.RuneCountInString(str) > m.Max {
			return fmt.Errorf("config %s must be at most %d characters, got %d", key, m.Max, utf8.RuneCountInString(str))
		}
	}
	return nil
}

func (m *MaxLength) Name() string {
	return "max_length"
}

// Pattern validator ensures string matches regex pattern
type Pattern struct {
	Pattern string
	regex   *regexp.Regexp
}

func (p *Pattern) compile() error {
	if p.regex == nil {
		compiled, err := regexp.Compile(p.Pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern %s: %w", p.Pattern, err)
		}
		p.regex = compiled
	}
	return nil
}

func (p *Pattern) Validate(value any, key string) error {
	if err := p.compile(); err != nil {
		return err
	}

	if str, ok := toString(value); ok {
		if !p.regex.MatchString(str) {
			return fmt.Errorf("config %s must match pattern %s, got %s", key, p.Pattern, str)
		}
	}
	return nil
}

func (p *Pattern) Name() string {
	return "pattern"
}

// URL validator ensures string is a valid URL
type URL struct{}

func (u *URL) Validate(value any, key string) error {
	if str, ok := toString(value); ok {
		if str == "" {
			return nil // Empty string is allowed (can use default)
		}
		if parsed, err := url.Parse(str); err != nil {
			return fmt.Errorf("config %s must be a valid URL, got %s: %w", key, str, err)
		} else {
			// Check that URL has both scheme and host
			if parsed.Scheme == "" {
				return fmt.Errorf("config %s must be a valid URL with scheme (e.g., http://, https://), got %s", key, str)
			}
			if parsed.Host == "" {
				return fmt.Errorf("config %s must be a valid URL with host, got %s", key, str)
			}
		}
	}
	return nil
}

func (u *URL) Name() string {
	return "url"
}

// Range validator ensures numeric values are within range
type Range struct {
	Min float64
	Max float64
}

func (r *Range) Validate(value any, key string) error {
	if num, err := toFloat64(value); err == nil {
		if num < r.Min || num > r.Max {
			return fmt.Errorf("config %s must be between %g and %g, got %g", key, r.Min, r.Max, num)
		}
	}
	return nil
}

func (r *Range) Name() string {
	return "range"
}

// OneOf validator ensures value is one of the allowed values
type OneOf struct {
	Values []string
}

func (o *OneOf) Validate(value any, key string) error {
	if str, ok := toString(value); ok {
		for _, allowed := range o.Values {
			if str == allowed {
				return nil
			}
		}
		return fmt.Errorf("config %s must be one of %v, got %s", key, o.Values, str)
	}
	return nil
}

func (o *OneOf) Name() string {
	return "one_of"
}

// ConfigSchema defines validation rules for configuration
type ConfigSchema struct {
	fields map[string][]Validator
}

// NewConfigSchema creates a new configuration schema
func NewConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		fields: make(map[string][]Validator),
	}
}

// AddField adds validation rules for a configuration field
func (s *ConfigSchema) AddField(key string, validators ...Validator) *ConfigSchema {
	if s.fields[key] == nil {
		s.fields[key] = make([]Validator, 0)
	}
	s.fields[key] = append(s.fields[key], validators...)
	return s
}

// Validate validates configuration data against the schema
func (s *ConfigSchema) Validate(data map[string]any) error {
	var errs []string

	for key, validators := range s.fields {
		value, exists := data[key]

		for _, validator := range validators {
			if err := validator.Validate(value, key); err != nil {
				if !exists && strings.Contains(err.Error(), "required") {
					// Only add required field errors if the field doesn't exist
					errs = append(errs, err.Error())
				} else if exists {
					// Add validation errors for existing fields
					errs = append(errs, err.Error())
				}
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration validation failed:\n%s", strings.Join(errs, "\n"))
	}

	return nil
}

// Type-safe configuration access

// String returns a type-safe string configuration
func (c *Config) String(key string, defaultValue string, validators ...Validator) (string, error) {
	value := c.GetString(key, defaultValue)

	for _, validator := range validators {
		if err := validator.Validate(value, key); err != nil {
			return defaultValue, err
		}
	}

	return value, nil
}

// Int returns a type-safe int configuration
func (c *Config) Int(key string, defaultValue int, validators ...Validator) (int, error) {
	value := c.GetInt(key, defaultValue)

	// Add range validator for numeric values if none specified
	hasRangeValidator := false
	for _, validator := range validators {
		if _, ok := validator.(*Range); ok {
			hasRangeValidator = true
			break
		}
	}

	if !hasRangeValidator {
		// Add default range validator to prevent overflow
		validators = append(validators, &Range{Min: -2147483648, Max: 2147483647})
	}

	for _, validator := range validators {
		if err := validator.Validate(value, key); err != nil {
			return defaultValue, err
		}
	}

	return value, nil
}

// Bool returns a type-safe bool configuration
func (c *Config) Bool(key string, defaultValue bool, validators ...Validator) (bool, error) {
	value := c.GetBool(key, defaultValue)

	for _, validator := range validators {
		if err := validator.Validate(value, key); err != nil {
			return defaultValue, err
		}
	}

	return value, nil
}

// Float returns a type-safe float64 configuration
func (c *Config) Float(key string, defaultValue float64, validators ...Validator) (float64, error) {
	value := c.GetFloat(key, defaultValue)

	for _, validator := range validators {
		if err := validator.Validate(value, key); err != nil {
			return defaultValue, err
		}
	}

	return value, nil
}

// DurationMs returns a type-safe time.Duration configuration
func (c *Config) DurationMs(key string, defaultValueMs int, validators ...Validator) (time.Duration, error) {
	value := c.GetDurationMs(key, defaultValueMs)

	// Convert to int for validation
	intValue := int(value.Milliseconds())

	for _, validator := range validators {
		if err := validator.Validate(intValue, key); err != nil {
			return time.Duration(defaultValueMs) * time.Millisecond, err
		}
	}

	return value, nil
}

// toString converts any value to string
func toString(value any) (string, bool) {
	if value == nil {
		return "", false
	}

	switch v := value.(type) {
	case string:
		return v, true
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v), true
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), true
	case float32, float64:
		return fmt.Sprintf("%g", v), true
	case bool:
		return strconv.FormatBool(v), true
	case time.Time:
		return v.Format(time.RFC3339), true
	default:
		return fmt.Sprintf("%v", v), true
	}
}

// toFloat64 converts any value to float64
func toFloat64(value any) (float64, error) {
	if str, ok := toString(value); ok {
		return strconv.ParseFloat(str, 64)
	}
	return 0, fmt.Errorf("cannot convert to float64")
}
