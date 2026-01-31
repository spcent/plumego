package config

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/spcent/plumego/log"
)

// Type conversion utilities for configuration values.

// toString converts any value to its string representation.
func toString(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// toInt converts any value to int, returning defaultValue on failure.
func toInt(value any, defaultValue int) int {
	switch v := value.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint8:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case bool:
		if v {
			return 1
		}
		return 0
	case string:
		if v := strings.TrimSpace(v); v != "" {
			if intValue, err := strconv.Atoi(v); err == nil {
				return intValue
			}
		}
	}
	return defaultValue
}

// toFloat64 converts any value to float64, returning defaultValue on failure.
func toFloat64(value any, defaultValue float64) float64 {
	switch v := value.(type) {
	case float32:
		return float64(v)
	case float64:
		return v
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case bool:
		if v {
			return 1.0
		}
		return 0.0
	case string:
		if v := strings.TrimSpace(v); v != "" {
			if floatValue, err := strconv.ParseFloat(v, 64); err == nil {
				return floatValue
			}
		}
	}
	return defaultValue
}

// toBool converts any value to bool, returning defaultValue on failure.
func toBool(value any, defaultValue bool) bool {
	switch v := value.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case int8:
		return v != 0
	case int16:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case uint:
		return v != 0
	case uint8:
		return v != 0
	case uint16:
		return v != 0
	case uint32:
		return v != 0
	case uint64:
		return v != 0
	case float32:
		return v != 0
	case float64:
		return v != 0
	case string:
		if v := strings.TrimSpace(v); v != "" {
			return parseBool(v, defaultValue)
		}
	}
	return defaultValue
}

// parseBool parses a string to bool with common true/false representations.
func parseBool(value string, defaultValue bool) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on", "t":
		return true
	case "0", "false", "no", "n", "off", "f":
		return false
	default:
		return defaultValue
	}
}

// valuesEqual compares two configuration values for equality.
func valuesEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch av := a.(type) {
	case string:
		if bv, ok := b.(string); ok {
			return av == bv
		}
	case int:
		if bv, ok := b.(int); ok {
			return av == bv
		}
	case float64:
		if bv, ok := b.(float64); ok {
			return av == bv
		}
	case bool:
		if bv, ok := b.(bool); ok {
			return av == bv
		}
	}

	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// Key normalization utilities.

// normalizeKey converts a configuration key to normalized form (lowercase snake_case).
func normalizeKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	return strings.ToLower(toSnakeCase(key))
}

// toSnakeCase converts CamelCase to snake_case.
func toSnakeCase(s string) string {
	if s == "" {
		return s
	}

	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			if s[i-1] >= 'a' && s[i-1] <= 'z' {
				result = append(result, '_')
			}
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

// normalizeData normalizes all keys in a configuration map.
func normalizeData(data map[string]any, logger log.StructuredLogger) map[string]any {
	if len(data) == 0 {
		return map[string]any{}
	}

	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	normalized := make(map[string]any, len(data))
	seen := make(map[string][]string, len(data))

	for _, key := range keys {
		norm := normalizeKey(key)
		if norm == "" {
			continue
		}
		seen[norm] = append(seen[norm], key)
		normalized[norm] = data[key]
	}

	if logger != nil {
		for norm, originals := range seen {
			if len(originals) > 1 {
				logger.Warn("Config key collision after normalization", log.Fields{
					"key":       norm,
					"originals": originals,
				})
			}
		}
	}

	return normalized
}

// lookupValue looks up a configuration value with key normalization.
func lookupValue(data map[string]any, key string) (any, bool) {
	if len(data) == 0 {
		return nil, false
	}

	normalized := normalizeKey(key)
	if normalized != "" {
		if value, exists := data[normalized]; exists {
			return value, true
		}
	}

	if value, exists := data[key]; exists {
		return value, true
	}

	return nil, false
}
