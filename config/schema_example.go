package config

// DefaultConfigSchema creates and returns the default configuration schema
// This demonstrates how to use the enhanced validation and documentation system
func DefaultConfigSchema() *ConfigSchemaManager {
	csm := NewConfigSchemaManager()

	// Core Application Configuration
	csm.Register("APP_ADDR", ConfigSchemaEntry{
		Type:        "string",
		Required:    false,
		Default:     ":8080",
		Description: "Server listen address",
		Validators: []Validator{
			&Pattern{Pattern: `^:\d+$|^\d+\.\d+\.\d+\.\d+:\d+$`},
		},
	})

	csm.Register("APP_DEBUG", ConfigSchemaEntry{
		Type:        "bool",
		Required:    false,
		Default:     false,
		Description: "Enable debug mode",
		Validators:  []Validator{},
	})

	csm.Register("APP_ENV_FILE", ConfigSchemaEntry{
		Type:        "string",
		Required:    false,
		Default:     ".env",
		Description: "Path to environment file",
		Validators: []Validator{
			&MinLength{Min: 1},
		},
	})

	// Timeouts Configuration
	csm.Register("APP_SHUTDOWN_TIMEOUT_MS", ConfigSchemaEntry{
		Type:        "int",
		Required:    false,
		Default:     5000,
		Description: "Graceful shutdown timeout in milliseconds",
		Validators: []Validator{
			&Range{Min: 1000, Max: 30000},
		},
	})

	csm.Register("APP_READ_TIMEOUT_MS", ConfigSchemaEntry{
		Type:        "int",
		Required:    false,
		Default:     30000,
		Description: "HTTP read timeout in milliseconds",
		Validators: []Validator{
			&Range{Min: 1000, Max: 60000},
		},
	})

	csm.Register("APP_WRITE_TIMEOUT_MS", ConfigSchemaEntry{
		Type:        "int",
		Required:    false,
		Default:     30000,
		Description: "HTTP write timeout in milliseconds",
		Validators: []Validator{
			&Range{Min: 1000, Max: 60000},
		},
	})

	// Request Limits
	csm.Register("APP_MAX_BODY_BYTES", ConfigSchemaEntry{
		Type:        "int",
		Required:    false,
		Default:     10485760, // 10MB
		Description: "Maximum request body size in bytes",
		Validators: []Validator{
			&Range{Min: 1024, Max: 1073741824}, // 1KB to 1GB
		},
	})

	csm.Register("APP_MAX_CONCURRENCY", ConfigSchemaEntry{
		Type:        "int",
		Required:    false,
		Default:     256,
		Description: "Maximum concurrent requests",
		Validators: []Validator{
			&Range{Min: 1, Max: 10000},
		},
	})

	// WebSocket Configuration
	csm.Register("WS_SECRET", ConfigSchemaEntry{
		Type:        "string",
		Required:    false,
		Default:     "",
		Description: "WebSocket JWT signing key (minimum 32 bytes)",
		Validators: []Validator{
			&MinLength{Min: 32},
		},
	})

	csm.Register("WS_ROUTE_PATH", ConfigSchemaEntry{
		Type:        "string",
		Required:    false,
		Default:     "/ws",
		Description: "WebSocket endpoint path",
		Validators: []Validator{
			&Pattern{Pattern: `^/[\w\-/]*$`},
		},
	})

	csm.Register("WS_BROADCAST_ENABLED", ConfigSchemaEntry{
		Type:        "bool",
		Required:    false,
		Default:     true,
		Description: "Enable WebSocket broadcast endpoint",
		Validators:  []Validator{},
	})

	// Webhook Configuration
	csm.Register("WEBHOOK_TRIGGER_TOKEN", ConfigSchemaEntry{
		Type:        "string",
		Required:    false,
		Default:     "",
		Description: "Token for triggering outbound webhooks",
		Validators: []Validator{
			&MinLength{Min: 8},
		},
	})

	csm.Register("GITHUB_WEBHOOK_SECRET", ConfigSchemaEntry{
		Type:        "string",
		Required:    false,
		Default:     "",
		Description: "GitHub webhook secret for signature verification",
		Validators: []Validator{
			&MinLength{Min: 16},
		},
	})

	csm.Register("STRIPE_WEBHOOK_SECRET", ConfigSchemaEntry{
		Type:        "string",
		Required:    false,
		Default:     "",
		Description: "Stripe webhook secret for signature verification",
		Validators: []Validator{
			&MinLength{Min: 16},
		},
	})

	csm.Register("WEBHOOK_MAX_BODY_BYTES", ConfigSchemaEntry{
		Type:        "int",
		Required:    false,
		Default:     1048576, // 1MB
		Description: "Maximum webhook body size in bytes",
		Validators: []Validator{
			&Range{Min: 1024, Max: 10485760}, // 1KB to 10MB
		},
	})

	csm.Register("WEBHOOK_TOPIC_PREFIX_GITHUB", ConfigSchemaEntry{
		Type:        "string",
		Required:    false,
		Default:     "in.github.",
		Description: "Topic prefix for GitHub webhook events",
		Validators: []Validator{
			&Pattern{Pattern: `^[\w\-.]+$`},
		},
	})

	csm.Register("WEBHOOK_TOPIC_PREFIX_STRIPE", ConfigSchemaEntry{
		Type:        "string",
		Required:    false,
		Default:     "in.stripe.",
		Description: "Topic prefix for Stripe webhook events",
		Validators: []Validator{
			&Pattern{Pattern: `^[\w\-.]+$`},
		},
	})

	// PubSub Configuration
	csm.Register("PUBSUB_DEBUG_ENABLED", ConfigSchemaEntry{
		Type:        "bool",
		Required:    false,
		Default:     false,
		Description: "Enable PubSub debug endpoints",
		Validators:  []Validator{},
	})

	csm.Register("PUBSUB_DEBUG_PATH", ConfigSchemaEntry{
		Type:        "string",
		Required:    false,
		Default:     "/_debug/pubsub",
		Description: "PubSub debug endpoint path",
		Validators: []Validator{
			&Pattern{Pattern: `^/[\w\-/]*$`},
		},
	})

	// TLS Configuration
	csm.Register("TLS_ENABLED", ConfigSchemaEntry{
		Type:        "bool",
		Required:    false,
		Default:     false,
		Description: "Enable TLS/HTTPS",
		Validators:  []Validator{},
	})

	csm.Register("TLS_CERT_FILE", ConfigSchemaEntry{
		Type:        "string",
		Required:    false,
		Default:     "",
		Description: "Path to TLS certificate file",
		Validators: []Validator{
			&MinLength{Min: 1},
		},
	})

	csm.Register("TLS_KEY_FILE", ConfigSchemaEntry{
		Type:        "string",
		Required:    false,
		Default:     "",
		Description: "Path to TLS private key file",
		Validators: []Validator{
			&MinLength{Min: 1},
		},
	})

	// Security Configuration
	csm.Register("AUTH_TOKEN", ConfigSchemaEntry{
		Type:        "string",
		Required:    false,
		Default:     "",
		Description: "Simple authentication token",
		Validators: []Validator{
			&MinLength{Min: 8},
		},
	})

	csm.Register("SECURITY_HEADERS_ENABLED", ConfigSchemaEntry{
		Type:        "bool",
		Required:    false,
		Default:     true,
		Description: "Enable security headers middleware",
		Validators:  []Validator{},
	})

	csm.Register("ABUSE_GUARD_ENABLED", ConfigSchemaEntry{
		Type:        "bool",
		Required:    false,
		Default:     true,
		Description: "Enable abuse guard middleware",
		Validators:  []Validator{},
	})

	return csm
}

// GetExampleConfig returns an example configuration map with all defaults
func GetExampleConfig() map[string]any {
	schema := DefaultConfigSchema()
	config := make(map[string]any)

	// Apply all defaults
	for _, entry := range schema.ListSchemas() {
		if entry.Default != nil {
			config[entry.Key] = entry.Default
		}
	}

	return config
}

// ValidateExampleConfig demonstrates validation with the schema
func ValidateExampleConfig(config map[string]any) (bool, []string) {
	schema := DefaultConfigSchema()
	errors := schema.ValidateAll(config)

	if len(errors) == 0 {
		return true, nil
	}

	messages := make([]string, len(errors))
	for i, err := range errors {
		messages[i] = err.Error()
	}

	return false, messages
}
