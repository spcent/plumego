package config

import (
	"bufio"
	"context"
	"os"
	"sync"
	"time"

	"github.com/spcent/plumego/log"
)

// Global configuration instance for package-level convenience functions.
var (
	globalConfig        *Manager
	globalConfigMu      sync.RWMutex
	globalInitialized   bool // protected by globalConfigMu
	globalInitErr       error
)

// LoadEnvFile loads environment variables from a file.
// If overwrite is true, existing environment variables will be overwritten.
// If overwrite is false, existing environment variables will be preserved.
func LoadEnvFile(filepath string, overwrite bool) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, ok := parseEnvLine(scanner.Text())
		if !ok {
			continue
		}
		if overwrite || os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

// LoadEnv loads environment variables from a file (alias for LoadEnvFile).
func LoadEnv(filepath string, overwrite bool) error {
	return LoadEnvFile(filepath, overwrite)
}

// InitDefault initializes the global config with environment and file sources.
// It is idempotent: subsequent calls return the error from the first invocation.
func InitDefault() error {
	// Fast path — already initialized
	globalConfigMu.RLock()
	if globalInitialized {
		err := globalInitErr
		globalConfigMu.RUnlock()
		return err
	}
	globalConfigMu.RUnlock()

	// Slow path — initialize under write lock
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()

	if globalInitialized { // double-checked locking
		return globalInitErr
	}

	logger := log.NewGLogger()
	globalConfig = NewManager(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	globalConfig.AddSource(NewEnvSource(""))

	configFiles := []string{
		".env",
		".env.local",
		"config.env",
		"config.json",
	}

	for _, configFile := range configFiles {
		if info, err := os.Stat(configFile); err == nil && !info.IsDir() {
			if info.Size() > 1024*1024 {
				logger.Warn("Skipping large config file", log.Fields{
					"file": configFile,
					"size": info.Size(),
				})
				continue
			}

			switch {
			case len(configFile) > 5 && configFile[len(configFile)-5:] == ".json":
				globalConfig.AddSource(NewFileSource(configFile, FormatJSON, true))
			case len(configFile) > 4 && configFile[len(configFile)-4:] == ".env":
				globalConfig.AddSource(NewFileSource(configFile, FormatEnv, true))
			}
		}
	}

	globalInitErr = globalConfig.Load(ctx)
	globalInitialized = true
	return globalInitErr
}

// GetGlobalConfig returns the global config instance.
func GetGlobalConfig() *Manager {
	globalConfigMu.RLock()
	cfg := globalConfig
	globalConfigMu.RUnlock()

	if cfg != nil {
		return cfg
	}

	if err := InitDefault(); err != nil {
		logger := log.NewGLogger()
		return NewManager(logger)
	}

	globalConfigMu.RLock()
	cfg = globalConfig
	globalConfigMu.RUnlock()

	if cfg == nil {
		logger := log.NewGLogger()
		return NewManager(logger)
	}

	return cfg
}

// SetGlobalConfig sets a custom global config instance.
// Passing a non-nil manager marks initialization as complete (further calls to
// InitDefault are no-ops). Passing nil resets the state so that the next call
// to InitDefault or GetGlobalConfig will re-initialize from environment/files.
// This is primarily useful for testing.
func SetGlobalConfig(config *Manager) {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()
	globalConfig = config
	globalInitErr = nil
	// nil means "reset" — allow InitDefault to run again
	globalInitialized = config != nil
}

// Package-level convenience functions that use the global config.

// GetString gets a configuration value as string with default.
func GetString(key, defaultValue string) string {
	return GetGlobalConfig().GetString(key, defaultValue)
}

// GetInt gets a configuration value as int with default.
func GetInt(key string, defaultValue int) int {
	return GetGlobalConfig().GetInt(key, defaultValue)
}

// GetBool gets a configuration value as bool with default.
func GetBool(key string, defaultValue bool) bool {
	return GetGlobalConfig().GetBool(key, defaultValue)
}

// GetFloat gets a configuration value as float64 with default.
func GetFloat(key string, defaultValue float64) float64 {
	return GetGlobalConfig().GetFloat(key, defaultValue)
}

// GetDurationMs gets a configuration value as duration (milliseconds) with default.
func GetDurationMs(key string, defaultValueMs int) time.Duration {
	return GetGlobalConfig().GetDurationMs(key, defaultValueMs)
}

// Has reports whether a configuration key exists in the global config.
func Has(key string) bool {
	return GetGlobalConfig().Has(key)
}

// GetStringSlice gets a configuration value as a string slice split by sep.
func GetStringSlice(key, sep string, defaultValue []string) []string {
	return GetGlobalConfig().GetStringSlice(key, sep, defaultValue)
}

// GetDuration gets a configuration value as time.Duration.
// Accepts Go duration strings ("30s", "5m") or plain integer milliseconds.
func GetDuration(key string, defaultValue time.Duration) time.Duration {
	return GetGlobalConfig().GetDuration(key, defaultValue)
}

// Set sets an environment variable and reloads the global config.
func Set(key, value string) error {
	os.Setenv(key, value)

	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()

	if globalConfig != nil {
		return globalConfig.Load(context.Background())
	}

	return nil
}

// Type-safe global accessors with validation.

// GetStringSafe gets a validated string configuration value.
func GetStringSafe(key, defaultValue string, validators ...Validator) (string, error) {
	return GetGlobalConfig().String(key, defaultValue, validators...)
}

// GetIntSafe gets a validated int configuration value.
func GetIntSafe(key string, defaultValue int, validators ...Validator) (int, error) {
	return GetGlobalConfig().Int(key, defaultValue, validators...)
}

// GetBoolSafe gets a validated bool configuration value.
func GetBoolSafe(key string, defaultValue bool, validators ...Validator) (bool, error) {
	return GetGlobalConfig().Bool(key, defaultValue, validators...)
}

// GetFloatSafe gets a validated float64 configuration value.
func GetFloatSafe(key string, defaultValue float64, validators ...Validator) (float64, error) {
	return GetGlobalConfig().Float(key, defaultValue, validators...)
}

// GetDurationMsSafe gets a validated duration configuration value.
func GetDurationMsSafe(key string, defaultValueMs int, validators ...Validator) (time.Duration, error) {
	return GetGlobalConfig().DurationMs(key, defaultValueMs, validators...)
}
