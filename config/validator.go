package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spcent/plumego/contract"
)

// Validator interface for configuration validation
type Validator interface {
	Validate(value any, key string) error
	Name() string
}

// ConfigSchemaEntry represents a configuration field with its metadata
type ConfigSchemaEntry struct {
	Key         string
	Type        string
	Required    bool
	Default     any
	Description string
	Validators  []Validator
}

// ConfigSchemaManager manages configuration schemas with validation and documentation
type ConfigSchemaManager struct {
	schemas map[string]ConfigSchemaEntry
}

// NewConfigSchemaManager creates a new configuration schema manager
func NewConfigSchemaManager() *ConfigSchemaManager {
	return &ConfigSchemaManager{
		schemas: make(map[string]ConfigSchemaEntry),
	}
}

// Register adds a configuration field schema
func (csm *ConfigSchemaManager) Register(key string, entry ConfigSchemaEntry) {
	entry.Key = key
	csm.schemas[key] = entry
}

// ValidateAll validates all configuration against registered schemas
func (csm *ConfigSchemaManager) ValidateAll(config map[string]any) []contract.APIError {
	var errors []contract.APIError

	for key, schema := range csm.schemas {
		value, exists := config[key]

		// Check required fields
		if schema.Required && !exists {
			errors = append(errors, contract.NewErrorBuilder().
				Status(400).
				Category(contract.CategoryValidation).
				Type(contract.ErrTypeRequired).
				Code("CONFIG_REQUIRED").
				Message(fmt.Sprintf("required configuration '%s' is missing", key)).
				Detail("key", key).
				Build())
			continue
		}

		// Apply default if missing
		if !exists && schema.Default != nil {
			config[key] = schema.Default
			continue
		}

		// Run validators
		if exists {
			for _, validator := range schema.Validators {
				if err := validator.Validate(value, key); err != nil {
					// Convert to APIError
					apiErr := contract.NewErrorBuilder().
						Status(400).
						Category(contract.CategoryValidation).
						Type(contract.ErrTypeValidation).
						Code("CONFIG_VALIDATION_FAILED").
						Message(err.Error()).
						Detail("key", key).
						Detail("value", value).
						Build()
					errors = append(errors, apiErr)
				}
			}
		}
	}

	return errors
}

// GenerateDocumentation generates markdown documentation for all schemas
func (csm *ConfigSchemaManager) GenerateDocumentation() string {
	var builder strings.Builder
	builder.WriteString("# Configuration Documentation\n\n")
	builder.WriteString("| Key | Type | Required | Default | Description | Validators |\n")
	builder.WriteString("|-----|------|----------|---------|-------------|------------|\n")

	for _, schema := range csm.schemas {
		defaultStr := "nil"
		if schema.Default != nil {
			defaultStr = fmt.Sprintf("%v", schema.Default)
		}
		requiredStr := "No"
		if schema.Required {
			requiredStr = "Yes"
		}

		validatorNames := make([]string, 0, len(schema.Validators))
		for _, v := range schema.Validators {
			validatorNames = append(validatorNames, v.Name())
		}
		validatorStr := strings.Join(validatorNames, ", ")
		if validatorStr == "" {
			validatorStr = "-"
		}

		builder.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			schema.Key, schema.Type, requiredStr, defaultStr, schema.Description, validatorStr))
	}

	return builder.String()
}

// GetSchema returns the schema for a specific key
func (csm *ConfigSchemaManager) GetSchema(key string) (ConfigSchemaEntry, bool) {
	schema, exists := csm.schemas[key]
	return schema, exists
}

// ListSchemas returns all registered schemas
func (csm *ConfigSchemaManager) ListSchemas() []ConfigSchemaEntry {
	schemas := make([]ConfigSchemaEntry, 0, len(csm.schemas))
	for _, schema := range csm.schemas {
		schemas = append(schemas, schema)
	}
	return schemas
}

// ValidateAndApplyDefaults validates configuration and applies defaults
func (csm *ConfigSchemaManager) ValidateAndApplyDefaults(config map[string]any) (map[string]any, []contract.APIError) {
	result := make(map[string]any)
	for k, v := range config {
		result[k] = v
	}

	errors := csm.ValidateAll(result)
	return result, errors
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

	// Convert to string for checking
	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		strValue = fmt.Sprintf("%d", v)
	case float32, float64:
		strValue = fmt.Sprintf("%g", v)
	case bool:
		strValue = strconv.FormatBool(v)
	default:
		strValue = fmt.Sprintf("%v", v)
	}

	if strings.TrimSpace(strValue) == "" {
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
	str := configToString(value)
	if utf8.RuneCountInString(str) < m.Min {
		return fmt.Errorf("config %s must be at least %d characters, got %d", key, m.Min, utf8.RuneCountInString(str))
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
	str := configToString(value)
	if utf8.RuneCountInString(str) > m.Max {
		return fmt.Errorf("config %s must be at most %d characters, got %d", key, m.Max, utf8.RuneCountInString(str))
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

	str := configToString(value)
	if !p.regex.MatchString(str) {
		return fmt.Errorf("config %s must match pattern %s, got %s", key, p.Pattern, str)
	}
	return nil
}

func (p *Pattern) Name() string {
	return "pattern"
}

// URL validator ensures string is a valid URL
type URL struct{}

func (u *URL) Validate(value any, key string) error {
	str := configToString(value)
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
	num, err := configToFloat64(value)
	if err != nil {
		return err
	}
	if num < r.Min || num > r.Max {
		return fmt.Errorf("config %s must be between %g and %g, got %g", key, r.Min, r.Max, num)
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
	str := configToString(value)
	for _, allowed := range o.Values {
		if str == allowed {
			return nil
		}
	}
	return fmt.Errorf("config %s must be one of %v, got %s", key, o.Values, str)
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

// configToString converts any value to string (helper for validators)
func configToString(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// configToFloat64 converts any value to float64 (helper for validators)
func configToFloat64(value any) (float64, error) {
	str := configToString(value)
	if str == "" {
		return 0, fmt.Errorf("cannot convert empty value to float64")
	}
	return strconv.ParseFloat(str, 64)
}
