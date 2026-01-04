package config

import (
	"bufio"
	"context"
	"os"
	"strings"
	"sync"
	"time"
)

// Global default config instance for backward compatibility
var (
	globalConfig   *Config
	globalConfigMu sync.RWMutex
	globalInitOnce sync.Once
	globalInitErr  error
)

// LoadEnvFile loads environment variables from a file
// If overwrite=true, existing environment variables will be overwritten.
// If overwrite=false, existing environment variables will not be overwritten.
func LoadEnvFile(filepath string, overwrite bool) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		if overwrite || os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

// LoadEnv loads environment variables from a file (backward compatibility alias)
func LoadEnv(filepath string, overwrite bool) error {
	return LoadEnvFile(filepath, overwrite)
}

// InitDefault initializes the global config with environment and file sources
func InitDefault() error {
	globalInitOnce.Do(func() {
		globalConfigMu.Lock()
		defer globalConfigMu.Unlock()

		// Create new config instance
		globalConfig = New()

		ctx := context.Background()

		// Add environment variable source
		globalConfig.AddSource(NewEnvSource(""))

		// Try to load from common config files
		configFiles := []string{
			".env",
			".env.local",
			"config.env",
			"config.json",
		}

		for _, configFile := range configFiles {
			if info, err := os.Stat(configFile); err == nil && !info.IsDir() {
				switch {
				case strings.HasSuffix(configFile, ".json"):
					globalConfig.AddSource(NewFileSource(configFile, FormatJSON, true))
				case strings.HasSuffix(configFile, ".env"):
					globalConfig.AddSource(NewFileSource(configFile, FormatEnv, true))
				}
			}
		}

		globalInitErr = globalConfig.Load(ctx)
	})

	return globalInitErr
}

// GetGlobalConfig returns the global config instance
func GetGlobalConfig() *Config {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()

	if globalConfig == nil {
		// Auto-initialize if not already done
		if err := InitDefault(); err != nil {
			// Return empty config on error
			return New()
		}
		return globalConfig
	}

	return globalConfig
}

// SetGlobalConfig allows setting a custom global config instance
// This is primarily useful for testing
func SetGlobalConfig(config *Config) {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()
	globalConfig = config
	// Reset the init once to allow re-initialization
	globalInitOnce = sync.Once{}
}

// GetString gets an environment variable as a string, returns default value if not found
func GetString(key, defaultValue string) string {
	return GetGlobalConfig().GetString(key, defaultValue)
}

// GetInt gets an environment variable as an integer, returns default value if not found or invalid
func GetInt(key string, defaultValue int) int {
	return GetGlobalConfig().GetInt(key, defaultValue)
}

// GetBool gets an environment variable as a boolean, returns default value if not found or invalid
func GetBool(key string, defaultValue bool) bool {
	return GetGlobalConfig().GetBool(key, defaultValue)
}

// GetFloat gets an environment variable as a float, returns default value if not found or invalid
func GetFloat(key string, defaultValue float64) float64 {
	return GetGlobalConfig().GetFloat(key, defaultValue)
}

// GetDurationMs gets an environment variable representing milliseconds and returns a time.Duration.
// Returns the default duration if not found or invalid.
func GetDurationMs(key string, defaultValueMs int) time.Duration {
	return GetGlobalConfig().GetDurationMs(key, defaultValueMs)
}

// Set sets an environment variable and reloads the global config
func Set(key, value string) error {
	os.Setenv(key, value)

	// Reload global config to pick up the change
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()

	if globalConfig != nil {
		ctx := context.Background()
		return globalConfig.Load(ctx)
	}

	return nil
}

// Type-safe global accessors with validation

// GetStringSafe gets a validated string configuration value
func GetStringSafe(key, defaultValue string, validators ...Validator) (string, error) {
	return GetGlobalConfig().String(key, defaultValue, validators...)
}

// GetIntSafe gets a validated int configuration value
func GetIntSafe(key string, defaultValue int, validators ...Validator) (int, error) {
	return GetGlobalConfig().Int(key, defaultValue, validators...)
}

// GetBoolSafe gets a validated bool configuration value
func GetBoolSafe(key string, defaultValue bool, validators ...Validator) (bool, error) {
	return GetGlobalConfig().Bool(key, defaultValue, validators...)
}

// GetFloatSafe gets a validated float64 configuration value
func GetFloatSafe(key string, defaultValue float64, validators ...Validator) (float64, error) {
	return GetGlobalConfig().Float(key, defaultValue, validators...)
}

// GetDurationMsSafe gets a validated duration configuration value
func GetDurationMsSafe(key string, defaultValueMs int, validators ...Validator) (time.Duration, error) {
	return GetGlobalConfig().DurationMs(key, defaultValueMs, validators...)
}
