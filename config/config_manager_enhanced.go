package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/spcent/plumego/contract"
	log "github.com/spcent/plumego/log"
)

// EnhancedConfigManager extends ConfigManager with validation and documentation capabilities
type EnhancedConfigManager struct {
	*ConfigManager
	schemaManager *ConfigSchemaManager
}

// NewEnhancedConfigManager creates a new enhanced configuration manager
func NewEnhancedConfigManager(logger log.StructuredLogger) *EnhancedConfigManager {
	return &EnhancedConfigManager{
		ConfigManager: NewConfigManager(logger),
		schemaManager: NewConfigSchemaManager(),
	}
}

// RegisterSchema registers a configuration schema for validation and documentation
func (ecm *EnhancedConfigManager) RegisterSchema(key string, entry ConfigSchemaEntry) {
	ecm.schemaManager.Register(key, entry)
}

// LoadWithValidation loads configuration and validates against registered schemas
func (ecm *EnhancedConfigManager) LoadWithValidation(ctx context.Context) error {
	// Load configuration
	if err := ecm.ConfigManager.Load(ctx); err != nil {
		return err
	}

	// Validate configuration
	config := ecm.ConfigManager.GetAll()
	errors := ecm.schemaManager.ValidateAll(config)

	if len(errors) > 0 {
		// Log validation errors
		for _, err := range errors {
			ecm.ConfigManager.logger.Error("Configuration validation failed", log.Fields{
				"error": err.Error(),
			})
		}
		return fmt.Errorf("configuration validation failed with %d errors", len(errors))
	}

	return nil
}

// LoadWithValidationAndDefaults loads configuration, applies defaults, and validates
func (ecm *EnhancedConfigManager) LoadWithValidationAndDefaults(ctx context.Context) error {
	// Load configuration
	if err := ecm.ConfigManager.Load(ctx); err != nil {
		return err
	}

	// Get current config and apply defaults
	config := ecm.ConfigManager.GetAll()
	updatedConfig, errors := ecm.schemaManager.ValidateAndApplyDefaults(config)

	// Update config manager with defaults applied
	for key, value := range updatedConfig {
		ecm.ConfigManager.mu.Lock()
		ecm.ConfigManager.data[key] = value
		ecm.ConfigManager.mu.Unlock()
	}

	if len(errors) > 0 {
		// Log validation errors
		for _, err := range errors {
			ecm.ConfigManager.logger.Error("Configuration validation failed", log.Fields{
				"error": err.Error(),
			})
		}
		return fmt.Errorf("configuration validation failed with %d errors", len(errors))
	}

	return nil
}

// GenerateDocumentation generates comprehensive configuration documentation
func (ecm *EnhancedConfigManager) GenerateDocumentation() string {
	var builder strings.Builder

	// Header
	builder.WriteString("# Configuration Documentation\n\n")
	builder.WriteString("This document describes all available configuration options for the application.\n\n")

	// Configuration Table
	builder.WriteString("## Configuration Options\n\n")
	builder.WriteString(ecm.schemaManager.GenerateDocumentation())

	// Usage Examples
	builder.WriteString("\n## Usage Examples\n\n")
	builder.WriteString("### Environment Variables\n\n")
	builder.WriteString("```bash\n")
	for _, schema := range ecm.schemaManager.ListSchemas() {
		envKey := strings.ToUpper(schema.Key)
		if schema.Default != nil {
			builder.WriteString(fmt.Sprintf("export %s=%v  # %s\n", envKey, schema.Default, schema.Description))
		} else {
			builder.WriteString(fmt.Sprintf("export %s=your_value  # %s\n", envKey, schema.Description))
		}
	}
	builder.WriteString("```\n\n")

	// Validation Rules
	builder.WriteString("## Validation Rules\n\n")
	builder.WriteString("- **Required fields**: Must be provided or have a default value\n")
	builder.WriteString("- **Type validation**: Values must match the specified type\n")
	builder.WriteString("- **Range constraints**: Numeric values must be within specified ranges\n")
	builder.WriteString("- **Pattern matching**: String values must match regex patterns\n")
	builder.WriteString("- **Length limits**: String values must respect min/max length constraints\n\n")

	return builder.String()
}

// GetConfigSchema returns the schema manager for programmatic access
func (ecm *EnhancedConfigManager) GetConfigSchema() *ConfigSchemaManager {
	return ecm.schemaManager
}

// ValidateConfig validates the current configuration against schemas
func (ecm *EnhancedConfigManager) ValidateConfig() []contract.StructuredError {
	config := ecm.ConfigManager.GetAll()
	return ecm.schemaManager.ValidateAll(config)
}

// GetValidationReport returns a detailed validation report
func (ecm *EnhancedConfigManager) GetValidationReport() string {
	errors := ecm.ValidateConfig()
	
	if len(errors) == 0 {
		return "✓ Configuration is valid"
	}

	var report strings.Builder
	report.WriteString(fmt.Sprintf("✗ Configuration validation failed with %d errors:\n\n", len(errors)))
	
	for i, err := range errors {
		report.WriteString(fmt.Sprintf("%d. %s\n", i+1, err.Error()))
		if err.Detail.Field != "" {
			report.WriteString(fmt.Sprintf("   Field: %s\n", err.Detail.Field))
		}
		if err.Detail.Value != nil {
			report.WriteString(fmt.Sprintf("   Value: %v\n", err.Detail.Value))
		}
		if len(err.Detail.Constraints) > 0 {
			report.WriteString(fmt.Sprintf("   Constraints: %v\n", err.Detail.Constraints))
		}
		report.WriteString("\n")
	}
	
	return report.String()
}

// LoadBestEffortWithValidation loads configuration from all sources with validation
func (ecm *EnhancedConfigManager) LoadBestEffortWithValidation(ctx context.Context) error {
	// Load best effort
	if err := ecm.ConfigManager.LoadBestEffort(ctx); err != nil {
		return err
	}

	// Apply defaults and validate
	config := ecm.ConfigManager.GetAll()
	updatedConfig, errors := ecm.schemaManager.ValidateAndApplyDefaults(config)

	// Update config with defaults
	for key, value := range updatedConfig {
		ecm.ConfigManager.mu.Lock()
		ecm.ConfigManager.data[key] = value
		ecm.ConfigManager.mu.Unlock()
	}

	if len(errors) > 0 {
		// Log but don't fail - this is best effort
		for _, err := range errors {
			ecm.ConfigManager.logger.Warn("Configuration validation warning", log.Fields{
				"error": err.Error(),
			})
		}
	}

	return nil
}

// MustLoadWithValidation loads configuration with validation and panics on error
// Useful for application startup where configuration is critical
func (ecm *EnhancedConfigManager) MustLoadWithValidation(ctx context.Context) {
	if err := ecm.LoadWithValidationAndDefaults(ctx); err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}
}