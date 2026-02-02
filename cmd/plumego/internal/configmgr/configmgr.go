package configmgr

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents the application configuration
type Config struct {
	Config map[string]any    `json:"config" yaml:"config"`
	Source map[string]string `json:"source,omitempty" yaml:"source,omitempty"`
}

// ValidationResult represents configuration validation result
type ValidationResult struct {
	Valid    bool              `json:"valid" yaml:"valid"`
	Errors   []ValidationIssue `json:"errors,omitempty" yaml:"errors,omitempty"`
	Warnings []ValidationIssue `json:"warnings,omitempty" yaml:"warnings,omitempty"`
}

// ValidationIssue represents a single validation issue
type ValidationIssue struct {
	Type    string `json:"type" yaml:"type"`
	Field   string `json:"field" yaml:"field"`
	Message string `json:"message" yaml:"message"`
}

// LoadConfig loads configuration from files and environment
func LoadConfig(dir, envFile string, resolve bool) (*Config, error) {
	config := &Config{
		Config: make(map[string]any),
		Source: make(map[string]string),
	}

	// Load from env file
	envPath := filepath.Join(dir, envFile)
	envVars := make(map[string]string)

	if file, err := os.Open(envPath); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				envVars[key] = value
			}
		}
	}

	// App configuration
	appConfig := make(map[string]any)

	addr := getConfigValue("APP_ADDR", envVars, resolve)
	if addr != "" {
		appConfig["addr"] = addr
		config.Source["app.addr"] = "env:APP_ADDR"
	} else {
		appConfig["addr"] = ":8080"
		config.Source["app.addr"] = "default"
	}

	debug := getConfigValue("APP_DEBUG", envVars, resolve)
	if debug != "" {
		appConfig["debug"] = debug == "true"
		config.Source["app.debug"] = "env:APP_DEBUG"
	} else {
		appConfig["debug"] = false
		config.Source["app.debug"] = "default"
	}

	shutdownTimeout := getConfigValue("APP_SHUTDOWN_TIMEOUT_MS", envVars, resolve)
	if shutdownTimeout != "" {
		appConfig["shutdown_timeout_ms"] = shutdownTimeout
		config.Source["app.shutdown_timeout_ms"] = "env:APP_SHUTDOWN_TIMEOUT_MS"
	} else {
		appConfig["shutdown_timeout_ms"] = 5000
		config.Source["app.shutdown_timeout_ms"] = "default"
	}

	config.Config["app"] = appConfig

	// Security configuration
	securityConfig := make(map[string]any)

	wsSecret := getConfigValue("WS_SECRET", envVars, resolve)
	if wsSecret != "" {
		securityConfig["ws_secret"] = wsSecret
		config.Source["security.ws_secret"] = "env:WS_SECRET"
	}

	jwtExpiry := getConfigValue("JWT_EXPIRY", envVars, resolve)
	if jwtExpiry != "" {
		securityConfig["jwt_expiry"] = jwtExpiry
		config.Source["security.jwt_expiry"] = "env:JWT_EXPIRY"
	} else {
		securityConfig["jwt_expiry"] = "15m"
		config.Source["security.jwt_expiry"] = "default"
	}

	config.Config["security"] = securityConfig

	return config, nil
}

// RedactSensitive redacts sensitive values in configuration
func RedactSensitive(config *Config) *Config {
	redacted := &Config{
		Config: make(map[string]any),
		Source: config.Source,
	}

	for key, value := range config.Config {
		if key == "security" {
			if secMap, ok := value.(map[string]any); ok {
				redactedSec := make(map[string]any)
				for secKey, secValue := range secMap {
					if strings.Contains(secKey, "secret") || strings.Contains(secKey, "key") {
						redactedSec[secKey] = "***REDACTED***"
					} else {
						redactedSec[secKey] = secValue
					}
				}
				redacted.Config[key] = redactedSec
			}
		} else {
			redacted.Config[key] = value
		}
	}

	return redacted
}

// ValidateConfig validates the configuration
func ValidateConfig(dir, envFile string) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Errors:   []ValidationIssue{},
		Warnings: []ValidationIssue{},
	}

	// Check if go.mod exists
	goModPath := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationIssue{
			Type:    "missing_file",
			Field:   "go.mod",
			Message: "go.mod file not found",
		})
	}

	// Check if env file exists
	envPath := filepath.Join(dir, envFile)
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		result.Warnings = append(result.Warnings, ValidationIssue{
			Type:    "missing_file",
			Field:   envFile,
			Message: fmt.Sprintf("%s file not found (optional)", envFile),
		})
	} else {
		// Validate env file content
		envVars := make(map[string]string)
		if file, err := os.Open(envPath); err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(line, "#") || line == "" {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					envVars[parts[0]] = parts[1]
				}
			}
		}

		// Check for required secrets
		requiredSecrets := []string{"WS_SECRET"}
		for _, secret := range requiredSecrets {
			if _, ok := envVars[secret]; !ok {
				if os.Getenv(secret) == "" {
					result.Warnings = append(result.Warnings, ValidationIssue{
						Type:    "missing_secret",
						Field:   secret,
						Message: fmt.Sprintf("%s not set in %s or environment", secret, envFile),
					})
				}
			}
		}
	}

	return result
}

// InitConfig creates default configuration files
func InitConfig(dir string) ([]string, error) {
	created := []string{}

	// Create env.example if it doesn't exist
	envExamplePath := filepath.Join(dir, "env.example")
	if _, err := os.Stat(envExamplePath); os.IsNotExist(err) {
		content := `# Application Configuration
APP_ADDR=:8080
APP_DEBUG=false
APP_SHUTDOWN_TIMEOUT_MS=5000
APP_MAX_BODY_BYTES=10485760
APP_MAX_CONCURRENCY=256

# Security Configuration
WS_SECRET=your-websocket-secret-here-32-bytes-minimum
JWT_SECRET=your-jwt-secret-here
JWT_EXPIRY=15m

# Database Configuration (optional)
# DB_URL=postgres://user:pass@localhost:5432/dbname

# Redis Configuration (optional)
# REDIS_ADDR=localhost:6379
# REDIS_PASSWORD=
# REDIS_DB=0
`
		if err := os.WriteFile(envExamplePath, []byte(content), 0644); err != nil {
			return created, fmt.Errorf("failed to write env.example: %w", err)
		}
		created = append(created, "env.example")
	}

	// Create .plumego.yaml if it doesn't exist
	plumegoConfigPath := filepath.Join(dir, ".plumego.yaml")
	if _, err := os.Stat(plumegoConfigPath); os.IsNotExist(err) {
		content := `project:
  name: myapp
  module: github.com/user/myapp

dev:
  addr: :8080
  watch:
    - "**/*.go"
  exclude:
    - "**/*_test.go"
    - "**/vendor/**"
  reload: true

build:
  output: ./bin/app
  embed_frontend: false
  tags: []

test:
  timeout: 20s
  race: true
  coverage: true
`
		if err := os.WriteFile(plumegoConfigPath, []byte(content), 0644); err != nil {
			return created, fmt.Errorf("failed to write .plumego.yaml: %w", err)
		}
		created = append(created, ".plumego.yaml")
	}

	return created, nil
}

// GetEnvVars returns all environment variables
func GetEnvVars(dir, envFile string) map[string]any {
	result := make(map[string]any)

	envPath := filepath.Join(dir, envFile)
	envVars := make(map[string]string)

	if file, err := os.Open(envPath); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				envVars[parts[0]] = parts[1]
			}
		}
	}

	result["file"] = envVars

	// Get system environment variables
	systemEnv := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			// Only include APP_*, WS_*, JWT_*, DB_*, REDIS_* prefixes
			key := parts[0]
			if strings.HasPrefix(key, "APP_") ||
				strings.HasPrefix(key, "WS_") ||
				strings.HasPrefix(key, "JWT_") ||
				strings.HasPrefix(key, "DB_") ||
				strings.HasPrefix(key, "REDIS_") {
				systemEnv[key] = parts[1]
			}
		}
	}
	result["system"] = systemEnv

	return result
}

func getConfigValue(key string, envVars map[string]string, resolve bool) string {
	// First check environment
	if val := os.Getenv(key); val != "" {
		return val
	}

	// Then check env file
	if resolve {
		if val, ok := envVars[key]; ok {
			return val
		}
	}

	return ""
}
