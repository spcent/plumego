package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

// If overwrite=true, existing environment variables will be overwritten.
// If overwrite=false, existing environment variables will not be overwritten.
func LoadEnv(filepath string, overwrite bool) error {
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

// GetString gets an environment variable as a string, returns default value if not found
func GetString(key, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	return value
}

// GetInt gets an environment variable as an integer, returns default value if not found or invalid
func GetInt(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// GetBool gets an environment variable as a boolean, returns default value if not found or invalid
func GetBool(key string, defaultValue bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	switch strings.ToLower(value) {
	case "1", "true", "yes", "y", "on", "t":
		return true
	case "0", "false", "no", "n", "off", "f":
		return false
	default:
		return defaultValue
	}
}

// GetFloat gets an environment variable as a float, returns default value if not found or invalid
func GetFloat(key string, defaultValue float64) float64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}

	return floatValue
}

// GetDurationMs gets an environment variable representing milliseconds and returns a time.Duration.
// Returns the default duration if not found or invalid.
func GetDurationMs(key string, defaultValueMs int) time.Duration {
	return time.Duration(GetInt(key, defaultValueMs)) * time.Millisecond
}

// Set sets an environment variable
func Set(key, value string) {
	os.Setenv(key, value)
}
