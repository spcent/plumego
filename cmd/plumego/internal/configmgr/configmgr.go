package configmgr

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
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

var requiredSecrets = []string{"WS_SECRET", "JWT_SECRET"}

// RequiredSecrets returns the stable secret keys checked by config diagnostics.
func RequiredSecrets() []string {
	out := make([]string, len(requiredSecrets))
	copy(out, requiredSecrets)
	return out
}

// ParseEnvFile reads a .env-style file and returns key-value pairs.
// Lines starting with '#' and blank lines are skipped.
func ParseEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	vars := make(map[string]string)
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid env line %d: missing '='", lineNum)
		}
		key := strings.TrimSpace(parts[0])
		if key == "" {
			return nil, fmt.Errorf("invalid env line %d: empty key", lineNum)
		}
		value, err := parseEnvValue(strings.TrimSpace(parts[1]), lineNum)
		if err != nil {
			return nil, err
		}
		vars[key] = value
	}
	return vars, scanner.Err()
}

// LoadConfig loads configuration from files and environment
func LoadConfig(dir, envFile string, resolve bool) (*Config, error) {
	config := &Config{
		Config: make(map[string]any),
		Source: make(map[string]string),
	}

	// Load from env file
	envPath := filepath.Join(dir, envFile)
	envVars, err := parseOptionalEnvFile(envPath)
	if err != nil {
		return nil, err
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

// RedactSensitive redacts sensitive values in configuration.
func RedactSensitive(config *Config) *Config {
	if config == nil {
		return nil
	}
	redacted := &Config{
		Config: make(map[string]any),
		Source: config.Source,
	}

	for key, value := range config.Config {
		redacted.Config[key] = redactValue(key, value)
	}

	return redacted
}

func redactValue(key string, value any) any {
	if isSensitiveKey(key) {
		return "***REDACTED***"
	}

	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for childKey, childValue := range typed {
			out[childKey] = redactValue(childKey, childValue)
		}
		return out
	case map[string]string:
		out := make(map[string]string, len(typed))
		for childKey, childValue := range typed {
			if isSensitiveKey(childKey) {
				out[childKey] = "***REDACTED***"
			} else {
				out[childKey] = childValue
			}
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = redactValue("", item)
		}
		return out
	default:
		return value
	}
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
		envVars, err := ParseEnvFile(envPath)
		if err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationIssue{
				Type:    "invalid_env_file",
				Field:   envFile,
				Message: err.Error(),
			})
			return result
		}
		if envVars == nil {
			envVars = make(map[string]string)
		}

		// Check for required secrets
		for _, secret := range RequiredSecrets() {
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

	return created, nil
}

// GetEnvVars returns all environment variables, with sensitive values redacted.
func GetEnvVars(dir, envFile string) map[string]any {
	result := make(map[string]any)

	envPath := filepath.Join(dir, envFile)
	envVars, _ := parseOptionalEnvFile(envPath)
	if envVars == nil {
		envVars = make(map[string]string)
	}

	redactedFile := make(map[string]string, len(envVars))
	for k, v := range envVars {
		if isSensitiveKey(k) {
			redactedFile[k] = "***REDACTED***"
		} else {
			redactedFile[k] = v
		}
	}
	result["file"] = redactedFile

	// Get system environment variables
	systemEnv := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			if strings.HasPrefix(key, "APP_") ||
				strings.HasPrefix(key, "WS_") ||
				strings.HasPrefix(key, "JWT_") ||
				strings.HasPrefix(key, "DB_") ||
				strings.HasPrefix(key, "REDIS_") {
				if isSensitiveKey(key) {
					systemEnv[key] = "***REDACTED***"
				} else {
					systemEnv[key] = parts[1]
				}
			}
		}
	}
	result["system"] = systemEnv

	return result
}

func parseOptionalEnvFile(path string) (map[string]string, error) {
	envVars, err := ParseEnvFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return envVars, nil
}

func parseEnvValue(value string, lineNum int) (string, error) {
	if value == "" {
		return "", nil
	}
	if value[0] == '\'' || value[0] == '"' {
		return parseQuotedEnvValue(value, lineNum)
	}
	return stripInlineComment(value), nil
}

func parseQuotedEnvValue(value string, lineNum int) (string, error) {
	quote := rune(value[0])
	var b strings.Builder
	escaped := false
	for i, r := range value[1:] {
		if escaped {
			b.WriteRune(r)
			escaped = false
			continue
		}
		if quote == '"' && r == '\\' {
			escaped = true
			continue
		}
		if r == quote {
			tail := strings.TrimSpace(value[i+2:])
			if tail != "" && !strings.HasPrefix(tail, "#") {
				return "", fmt.Errorf("invalid env line %d: unexpected content after quoted value", lineNum)
			}
			return b.String(), nil
		}
		b.WriteRune(r)
	}
	return "", fmt.Errorf("invalid env line %d: unterminated quoted value", lineNum)
}

func stripInlineComment(value string) string {
	var prev rune
	for i, r := range value {
		if r == '#' && (i == 0 || unicode.IsSpace(prev)) {
			return strings.TrimSpace(value[:i])
		}
		prev = r
	}
	return strings.TrimSpace(value)
}

// isSensitiveKey returns true if the env var key likely holds a secret.
func isSensitiveKey(key string) bool {
	upper := strings.ToUpper(key)
	return strings.Contains(upper, "SECRET") ||
		strings.Contains(upper, "PASSWORD") ||
		strings.Contains(upper, "TOKEN") ||
		strings.Contains(upper, "KEY") ||
		strings.Contains(upper, "DB_URL")
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
