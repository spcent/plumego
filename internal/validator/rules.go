package validator

import (
	"encoding/base64"
	"fmt"
	"math"
	"net"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	emailPattern        = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	alphaPattern        = regexp.MustCompile(`^[a-zA-Z]+$`)
	alphaNumPattern     = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	phonePattern        = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	phoneCleanupPattern = regexp.MustCompile(`[\s\-\(\)\.]`)
	uuidPattern         = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	macPattern          = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
	hexPattern          = regexp.MustCompile(`^[0-9A-Fa-f]+$`)
	datePattern         = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	timePattern         = regexp.MustCompile(`^\d{2}:\d{2}(:\d{2})?$`)
)

func invalidRegexRule(code string) Rule {
	return RuleFunc(func(value any) *ValidationError {
		return &ValidationError{Code: code, Message: "invalid regex pattern"}
	})
}

func isFiniteFloat(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

func parseFiniteFloat(value string) (float64, bool) {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	return parsed, err == nil && isFiniteFloat(parsed)
}

func Required() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return &ValidationError{Code: validationCodeRequired, Message: "field is required"}
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return &ValidationError{Code: validationCodeRequired, Message: "field is required"}
			}
			rv = rv.Elem()
		}

		switch rv.Kind() {
		case reflect.String:
			if strings.TrimSpace(rv.String()) == "" {
				return &ValidationError{Code: validationCodeRequired, Message: "field is required"}
			}
		case reflect.Slice, reflect.Map, reflect.Array, reflect.Chan:
			if rv.Len() == 0 {
				return &ValidationError{Code: validationCodeRequired, Message: "field is required"}
			}
		default:
			if rv.IsZero() {
				return &ValidationError{Code: validationCodeRequired, Message: "field is required"}
			}
		}
		return nil
	})
}

// Email performs email format validation
func Email() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeEmail, Message: "must be a string"}
		}

		str = strings.TrimSpace(str)
		if str == "" {
			return nil
		}

		if !emailPattern.MatchString(str) {
			return &ValidationError{Code: validationCodeEmail, Message: "invalid email format"}
		}

		return nil
	})
}

// Min validates minimum numeric value or string length
func Min(min int64) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil // Skip validation for nil values unless required
		}

		switch v := value.(type) {
		case int:
			if int64(v) < min {
				return &ValidationError{Code: validationCodeMin, Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case int8:
			if int64(v) < min {
				return &ValidationError{Code: validationCodeMin, Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case int16:
			if int64(v) < min {
				return &ValidationError{Code: validationCodeMin, Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case int32:
			if int64(v) < min {
				return &ValidationError{Code: validationCodeMin, Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case int64:
			if v < min {
				return &ValidationError{Code: validationCodeMin, Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case uint:
			if min > 0 && uint64(v) < uint64(min) {
				return &ValidationError{Code: validationCodeMin, Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case uint8:
			if min > 0 && uint64(v) < uint64(min) {
				return &ValidationError{Code: validationCodeMin, Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case uint16:
			if min > 0 && uint64(v) < uint64(min) {
				return &ValidationError{Code: validationCodeMin, Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case uint32:
			if min > 0 && uint64(v) < uint64(min) {
				return &ValidationError{Code: validationCodeMin, Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case uint64:
			if min > 0 && v < uint64(min) {
				return &ValidationError{Code: validationCodeMin, Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case float32:
			if float64(v) < float64(min) {
				return &ValidationError{Code: validationCodeMin, Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case float64:
			if v < float64(min) {
				return &ValidationError{Code: validationCodeMin, Message: fmt.Sprintf("must be at least %d", min)}
			}
			return nil
		case string:
			trimmed := strings.TrimSpace(v)
			if parsed, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
				if parsed < min {
					return &ValidationError{Code: validationCodeMin, Message: fmt.Sprintf("must be at least %d", min)}
				}
				return nil
			} else if parsed, err := strconv.ParseUint(trimmed, 10, 64); err == nil {
				if min > 0 && parsed < uint64(min) {
					return &ValidationError{Code: validationCodeMin, Message: fmt.Sprintf("must be at least %d", min)}
				}
				return nil
			} else if parsed, err := strconv.ParseFloat(trimmed, 64); err == nil {
				if parsed < float64(min) {
					return &ValidationError{Code: validationCodeMin, Message: fmt.Sprintf("must be at least %d", min)}
				}
				return nil
			} else {
				// For non-numeric strings, check length for backward compatibility
				if int64(utf8.RuneCountInString(v)) < min {
					return &ValidationError{Code: validationCodeMin, Message: fmt.Sprintf("must be at least %d characters", min)}
				}
			}
		}

		return nil
	})
}

// Max validates maximum numeric value or string length
func Max(max int64) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		switch v := value.(type) {
		case int:
			if int64(v) > max {
				return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case int8:
			if int64(v) > max {
				return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case int16:
			if int64(v) > max {
				return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case int32:
			if int64(v) > max {
				return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case int64:
			if v > max {
				return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case uint:
			if max < 0 {
				return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
			}
			if uint64(v) > uint64(max) {
				return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case uint8:
			if max < 0 {
				return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
			}
			if uint64(v) > uint64(max) {
				return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case uint16:
			if max < 0 {
				return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
			}
			if uint64(v) > uint64(max) {
				return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case uint32:
			if max < 0 {
				return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
			}
			if uint64(v) > uint64(max) {
				return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case uint64:
			if max < 0 {
				return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
			}
			if v > uint64(max) {
				return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case float32:
			if float64(v) > float64(max) {
				return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case float64:
			if v > float64(max) {
				return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
			}
			return nil
		case string:
			trimmed := strings.TrimSpace(v)
			if parsed, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
				if parsed > max {
					return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
				}
				return nil
			} else if parsed, err := strconv.ParseUint(trimmed, 10, 64); err == nil {
				if max < 0 || parsed > uint64(max) {
					return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
				}
				return nil
			} else if parsed, err := strconv.ParseFloat(trimmed, 64); err == nil {
				if parsed > float64(max) {
					return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d", max)}
				}
				return nil
			} else {
				// For non-numeric strings, check length for backward compatibility
				if int64(utf8.RuneCountInString(v)) > max {
					return &ValidationError{Code: validationCodeMax, Message: fmt.Sprintf("must be at most %d characters", max)}
				}
			}
		}

		return nil
	})
}

// MinLength validates minimum string length
func MinLength(min int) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return nil
		}

		if utf8.RuneCountInString(str) < min {
			return &ValidationError{Code: validationCodeMinLength, Message: fmt.Sprintf("must be at least %d characters", min)}
		}

		return nil
	})
}

// MaxLength validates maximum string length
func MaxLength(max int) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return nil
		}

		if utf8.RuneCountInString(str) > max {
			return &ValidationError{Code: validationCodeMaxLength, Message: fmt.Sprintf("must be at most %d characters", max)}
		}

		return nil
	})
}

// Numeric validates that value contains only digits
func Numeric() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeNumeric, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if _, ok := parseFiniteFloat(str); !ok {
			return &ValidationError{Code: validationCodeNumeric, Message: "must be a valid number"}
		}

		return nil
	})
}

// Alpha validates that string contains only alphabetic characters
func Alpha() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeAlpha, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !alphaPattern.MatchString(str) {
			return &ValidationError{Code: validationCodeAlpha, Message: "must contain only alphabetic characters"}
		}

		return nil
	})
}

// AlphaNum validates that string contains only alphanumeric characters
func AlphaNum() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeAlphaNum, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !alphaNumPattern.MatchString(str) {
			return &ValidationError{Code: validationCodeAlphaNum, Message: "must contain only alphanumeric characters"}
		}

		return nil
	})
}

// URL validates that string is a valid URL
func URL() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeURL, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if _, err := url.ParseRequestURI(str); err != nil {
			return &ValidationError{Code: validationCodeURL, Message: "must be a valid URL"}
		}

		return nil
	})
}

// Phone validates that string is a valid phone number (basic validation)
func Phone() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodePhone, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		// Remove common phone number characters
		cleaned := phoneCleanupPattern.ReplaceAllString(str, "")

		if !phonePattern.MatchString(cleaned) {
			return &ValidationError{Code: validationCodePhone, Message: "must be a valid phone number"}
		}

		return nil
	})
}

// Regex validates that string matches the given regular expression
func Regex(pattern string) Rule {
	if pattern == "" {
		// Return a rule that always passes
		return RuleFunc(func(value any) *ValidationError {
			return nil
		})
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return invalidRegexRule(validationCodeRegex)
	}

	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeRegex, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !regex.MatchString(str) {
			return &ValidationError{Code: validationCodeRegex, Message: "does not match required pattern"}
		}

		return nil
	})
}

// Int validates that value is an integer
func Int() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		switch v := value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return nil
		case string:
			if v == "" {
				return nil
			}
			if _, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err != nil {
				return &ValidationError{Code: validationCodeInt, Message: "must be an integer"}
			}
			return nil
		default:
			return &ValidationError{Code: validationCodeInt, Message: "must be an integer"}
		}
	})
}

// Float validates that value is a float
func Float() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		switch v := value.(type) {
		case float32:
			if !isFiniteFloat(float64(v)) {
				return &ValidationError{Code: validationCodeFloat, Message: "must be a finite float"}
			}
			return nil
		case float64:
			if !isFiniteFloat(v) {
				return &ValidationError{Code: validationCodeFloat, Message: "must be a finite float"}
			}
			return nil
		case string:
			if v == "" {
				return nil
			}
			if _, ok := parseFiniteFloat(v); !ok {
				return &ValidationError{Code: validationCodeFloat, Message: "must be a float"}
			}
			return nil
		default:
			return &ValidationError{Code: validationCodeFloat, Message: "must be a float"}
		}
	})
}

// Bool validates that value is a boolean
func Bool() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		switch v := value.(type) {
		case bool:
			return nil
		case string:
			if v == "" {
				return nil
			}
			lower := strings.ToLower(v)
			if lower == "true" || lower == "false" || lower == "1" || lower == "0" {
				return nil
			}
			return &ValidationError{Code: validationCodeBool, Message: "must be a boolean"}
		default:
			return &ValidationError{Code: validationCodeBool, Message: "must be a boolean"}
		}
	})
}

// UUID validates that string is a valid UUID
func UUID() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeUUID, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !uuidPattern.MatchString(strings.ToLower(str)) {
			return &ValidationError{Code: validationCodeUUID, Message: "must be a valid UUID"}
		}

		return nil
	})
}

// IPv4 validates that string is a valid IPv4 address
func IPv4() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeIPv4, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		ip := net.ParseIP(str)
		if ip == nil || ip.To4() == nil {
			return &ValidationError{Code: validationCodeIPv4, Message: "must be a valid IPv4 address"}
		}

		return nil
	})
}

// IPv6 validates that string is a valid IPv6 address
func IPv6() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeIPv6, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		ip := net.ParseIP(str)
		// IPv6 addresses contain colons, IPv4 addresses don't
		if ip == nil || !strings.Contains(str, ":") {
			return &ValidationError{Code: validationCodeIPv6, Message: "must be a valid IPv6 address"}
		}

		return nil
	})
}

// IP validates that string is a valid IP address (IPv4 or IPv6)
func IP() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeIP, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		ip := net.ParseIP(str)
		if ip == nil {
			return &ValidationError{Code: validationCodeIP, Message: "must be a valid IP address"}
		}

		return nil
	})
}

// MAC validates that string is a valid MAC address
func MAC() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeMAC, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !macPattern.MatchString(str) {
			return &ValidationError{Code: validationCodeMAC, Message: "must be a valid MAC address"}
		}

		return nil
	})
}

// Hex validates that string contains only hexadecimal characters
func Hex() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeHex, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !hexPattern.MatchString(str) {
			return &ValidationError{Code: validationCodeHex, Message: "must be a valid hexadecimal string"}
		}

		return nil
	})
}

// Base64 validates that string is a valid Base64 encoded string
func Base64() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeBase64, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		// Try standard encoding first
		if _, err := base64.StdEncoding.DecodeString(str); err == nil {
			return nil
		}

		// Try URL encoding (no padding)
		if _, err := base64.RawURLEncoding.DecodeString(str); err == nil {
			return nil
		}

		return &ValidationError{Code: validationCodeBase64, Message: "must be a valid Base64 string"}
	})
}

// Date validates that string is a valid date (YYYY-MM-DD)
func Date() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeDate, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !datePattern.MatchString(str) {
			return &ValidationError{Code: validationCodeDate, Message: "must be a valid date (YYYY-MM-DD)"}
		}

		if _, err := time.Parse("2006-01-02", str); err != nil {
			return &ValidationError{Code: validationCodeDate, Message: "must be a valid date"}
		}

		return nil
	})
}

// Time validates that string is a valid time (HH:MM:SS or HH:MM)
func Time() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeTime, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !timePattern.MatchString(str) {
			return &ValidationError{Code: validationCodeTime, Message: "must be a valid time (HH:MM:SS or HH:MM)"}
		}

		if _, err := time.Parse("15:04:05", str); err != nil {
			if _, err := time.Parse("15:04", str); err != nil {
				return &ValidationError{Code: validationCodeTime, Message: "must be a valid time"}
			}
		}

		return nil
	})
}

// DateTime validates that string is a valid datetime (YYYY-MM-DD HH:MM:SS or ISO8601)
func DateTime() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeDateTime, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		// Try common datetime formats
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02T15:04:05.999Z07:00",
			time.RFC3339,
			time.RFC3339Nano,
		}

		for _, format := range formats {
			if _, err := time.Parse(format, str); err == nil {
				return nil
			}
		}

		return &ValidationError{Code: validationCodeDateTime, Message: "must be a valid datetime"}
	})
}

// Array validates that value is an array or slice
func Array() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil
			}
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return &ValidationError{Code: validationCodeArray, Message: "must be an array or slice"}
		}

		return nil
	})
}

// Object validates that value is a struct or map
func Object() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil
			}
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Struct && rv.Kind() != reflect.Map {
			return &ValidationError{Code: validationCodeObject, Message: "must be an object"}
		}

		return nil
	})
}

// Validator holds validation configuration

func Range(min, max int64) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		switch v := value.(type) {
		case int:
			if int64(v) < min || int64(v) > max {
				return &ValidationError{Code: validationCodeRange, Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case int8:
			if int64(v) < min || int64(v) > max {
				return &ValidationError{Code: validationCodeRange, Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case int16:
			if int64(v) < min || int64(v) > max {
				return &ValidationError{Code: validationCodeRange, Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case int32:
			if int64(v) < min || int64(v) > max {
				return &ValidationError{Code: validationCodeRange, Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case int64:
			if v < min || v > max {
				return &ValidationError{Code: validationCodeRange, Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case uint:
			if min > 0 && (uint64(v) < uint64(min) || uint64(v) > uint64(max)) {
				return &ValidationError{Code: validationCodeRange, Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case uint8:
			if min > 0 && (uint64(v) < uint64(min) || uint64(v) > uint64(max)) {
				return &ValidationError{Code: validationCodeRange, Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case uint16:
			if min > 0 && (uint64(v) < uint64(min) || uint64(v) > uint64(max)) {
				return &ValidationError{Code: validationCodeRange, Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case uint32:
			if min > 0 && (uint64(v) < uint64(min) || uint64(v) > uint64(max)) {
				return &ValidationError{Code: validationCodeRange, Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case uint64:
			if min > 0 && (v < uint64(min) || v > uint64(max)) {
				return &ValidationError{Code: validationCodeRange, Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case float32:
			if float64(v) < float64(min) || float64(v) > float64(max) {
				return &ValidationError{Code: validationCodeRange, Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case float64:
			if v < float64(min) || v > float64(max) {
				return &ValidationError{Code: validationCodeRange, Message: fmt.Sprintf("must be between %d and %d", min, max)}
			}
			return nil
		case string:
			trimmed := strings.TrimSpace(v)
			if parsed, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
				if parsed < min || parsed > max {
					return &ValidationError{Code: validationCodeRange, Message: fmt.Sprintf("must be between %d and %d", min, max)}
				}
				return nil
			} else if parsed, err := strconv.ParseUint(trimmed, 10, 64); err == nil {
				if min > 0 && (parsed < uint64(min) || parsed > uint64(max)) {
					return &ValidationError{Code: validationCodeRange, Message: fmt.Sprintf("must be between %d and %d", min, max)}
				}
				return nil
			} else if parsed, err := strconv.ParseFloat(trimmed, 64); err == nil {
				if parsed < float64(min) || parsed > float64(max) {
					return &ValidationError{Code: validationCodeRange, Message: fmt.Sprintf("must be between %d and %d", min, max)}
				}
				return nil
			}
		}

		return nil
	})
}

// MinFloat validates minimum float value
func MinFloat(min float64) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		switch v := value.(type) {
		case float32:
			if !isFiniteFloat(float64(v)) {
				return &ValidationError{Code: validationCodeMinFloat, Message: "must be a finite number"}
			}
			if float64(v) < min {
				return &ValidationError{Code: validationCodeMinFloat, Message: fmt.Sprintf("must be at least %f", min)}
			}
			return nil
		case float64:
			if !isFiniteFloat(v) {
				return &ValidationError{Code: validationCodeMinFloat, Message: "must be a finite number"}
			}
			if v < min {
				return &ValidationError{Code: validationCodeMinFloat, Message: fmt.Sprintf("must be at least %f", min)}
			}
			return nil
		case string:
			if v == "" {
				return nil
			}
			if parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
				if !isFiniteFloat(parsed) {
					return &ValidationError{Code: validationCodeMinFloat, Message: "must be a finite number"}
				}
				if parsed < min {
					return &ValidationError{Code: validationCodeMinFloat, Message: fmt.Sprintf("must be at least %f", min)}
				}
				return nil
			}
		}

		return nil
	})
}

// MaxFloat validates maximum float value
func MaxFloat(max float64) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		switch v := value.(type) {
		case float32:
			if !isFiniteFloat(float64(v)) {
				return &ValidationError{Code: validationCodeMaxFloat, Message: "must be a finite number"}
			}
			if float64(v) > max {
				return &ValidationError{Code: validationCodeMaxFloat, Message: fmt.Sprintf("must be at most %f", max)}
			}
			return nil
		case float64:
			if !isFiniteFloat(v) {
				return &ValidationError{Code: validationCodeMaxFloat, Message: "must be a finite number"}
			}
			if v > max {
				return &ValidationError{Code: validationCodeMaxFloat, Message: fmt.Sprintf("must be at most %f", max)}
			}
			return nil
		case string:
			if v == "" {
				return nil
			}
			if parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
				if !isFiniteFloat(parsed) {
					return &ValidationError{Code: validationCodeMaxFloat, Message: "must be a finite number"}
				}
				if parsed > max {
					return &ValidationError{Code: validationCodeMaxFloat, Message: fmt.Sprintf("must be at most %f", max)}
				}
				return nil
			}
		}

		return nil
	})
}

// RangeFloat validates that float value is within a range [min, max]
func RangeFloat(min, max float64) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		switch v := value.(type) {
		case float32:
			if !isFiniteFloat(float64(v)) {
				return &ValidationError{Code: validationCodeRangeFloat, Message: "must be a finite number"}
			}
			if float64(v) < min || float64(v) > max {
				return &ValidationError{Code: validationCodeRangeFloat, Message: fmt.Sprintf("must be between %f and %f", min, max)}
			}
			return nil
		case float64:
			if !isFiniteFloat(v) {
				return &ValidationError{Code: validationCodeRangeFloat, Message: "must be a finite number"}
			}
			if v < min || v > max {
				return &ValidationError{Code: validationCodeRangeFloat, Message: fmt.Sprintf("must be between %f and %f", min, max)}
			}
			return nil
		case string:
			if v == "" {
				return nil
			}
			if parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
				if !isFiniteFloat(parsed) {
					return &ValidationError{Code: validationCodeRangeFloat, Message: "must be a finite number"}
				}
				if parsed < min || parsed > max {
					return &ValidationError{Code: validationCodeRangeFloat, Message: fmt.Sprintf("must be between %f and %f", min, max)}
				}
				return nil
			}
		}

		return nil
	})
}

// EmailList validates that string is a comma-separated list of valid emails
func EmailList() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeEmailList, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		emails := strings.Split(str, ",")
		for _, email := range emails {
			email = strings.TrimSpace(email)
			if email == "" {
				continue
			}
			if !emailPattern.MatchString(email) {
				return &ValidationError{Code: validationCodeEmailList, Message: "contains invalid email format"}
			}
		}

		return nil
	})
}

// URLList validates that string is a comma-separated list of valid URLs
func URLList() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeURLList, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		urls := strings.Split(str, ",")
		for _, urlStr := range urls {
			urlStr = strings.TrimSpace(urlStr)
			if urlStr == "" {
				continue
			}
			if _, err := url.ParseRequestURI(urlStr); err != nil {
				return &ValidationError{Code: validationCodeURLList, Message: "contains invalid URL"}
			}
		}

		return nil
	})
}

// OneOf validates that value is one of the allowed values
func OneOf(allowed ...string) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeOneOf, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		for _, allowedValue := range allowed {
			if str == allowedValue {
				return nil
			}
		}

		return &ValidationError{Code: validationCodeOneOf, Message: fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", "))}
	})
}

// NotEmpty validates that string is not empty
func NotEmpty() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return &ValidationError{Code: validationCodeNotEmpty, Message: "field is required"}
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeNotEmpty, Message: "must be a string"}
		}

		if str == "" {
			return &ValidationError{Code: validationCodeNotEmpty, Message: "field cannot be empty"}
		}

		return nil
	})
}

// NotZero validates that value is not the zero value for its type
func NotZero() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return &ValidationError{Code: validationCodeNotZero, Message: "field is required"}
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return &ValidationError{Code: validationCodeNotZero, Message: "field is required"}
			}
			rv = rv.Elem()
		}

		if rv.IsZero() {
			return &ValidationError{Code: validationCodeNotZero, Message: "field cannot be zero"}
		}

		return nil
	})
}

// AfterDate validates that date is after a specified date
func AfterDate(after string) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeAfterDate, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		parsedAfter, err := time.Parse("2006-01-02", after)
		if err != nil {
			return &ValidationError{Code: validationCodeAfterDate, Message: "invalid reference date"}
		}

		parsedValue, err := time.Parse("2006-01-02", str)
		if err != nil {
			return &ValidationError{Code: validationCodeAfterDate, Message: "must be a valid date"}
		}

		if !parsedValue.After(parsedAfter) {
			return &ValidationError{Code: validationCodeAfterDate, Message: fmt.Sprintf("must be after %s", after)}
		}

		return nil
	})
}

// BeforeDate validates that date is before a specified date
func BeforeDate(before string) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeBeforeDate, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		parsedBefore, err := time.Parse("2006-01-02", before)
		if err != nil {
			return &ValidationError{Code: validationCodeBeforeDate, Message: "invalid reference date"}
		}

		parsedValue, err := time.Parse("2006-01-02", str)
		if err != nil {
			return &ValidationError{Code: validationCodeBeforeDate, Message: "must be a valid date"}
		}

		if !parsedValue.Before(parsedBefore) {
			return &ValidationError{Code: validationCodeBeforeDate, Message: fmt.Sprintf("must be before %s", before)}
		}

		return nil
	})
}

// MinLengthBytes validates minimum byte length
func MinLengthBytes(min int) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return nil
		}

		if len(str) < min {
			return &ValidationError{Code: validationCodeMinLengthBytes, Message: fmt.Sprintf("must be at least %d bytes", min)}
		}

		return nil
	})
}

// MaxLengthBytes validates maximum byte length
func MaxLengthBytes(max int) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return nil
		}

		if len(str) > max {
			return &ValidationError{Code: validationCodeMaxLengthBytes, Message: fmt.Sprintf("must be at most %d bytes", max)}
		}

		return nil
	})
}

// Contains validates that string contains a substring
func Contains(substring string) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeContains, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !strings.Contains(str, substring) {
			return &ValidationError{Code: validationCodeContains, Message: fmt.Sprintf("must contain '%s'", substring)}
		}

		return nil
	})
}

// HasPrefix validates that string starts with a prefix
func HasPrefix(prefix string) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeHasPrefix, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !strings.HasPrefix(str, prefix) {
			return &ValidationError{Code: validationCodeHasPrefix, Message: fmt.Sprintf("must start with '%s'", prefix)}
		}

		return nil
	})
}

// HasSuffix validates that string ends with a suffix
func HasSuffix(suffix string) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeHasSuffix, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !strings.HasSuffix(str, suffix) {
			return &ValidationError{Code: validationCodeHasSuffix, Message: fmt.Sprintf("must end with '%s'", suffix)}
		}

		return nil
	})
}

// CaseInsensitive validates that string matches pattern case-insensitively
func CaseInsensitive(pattern string) Rule {
	if pattern == "" {
		return RuleFunc(func(value any) *ValidationError {
			return nil
		})
	}

	regex, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return invalidRegexRule(validationCodeCaseInsensitive)
	}

	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeCaseInsensitive, Message: "must be a string"}
		}

		if str == "" {
			return nil
		}

		if !regex.MatchString(str) {
			return &ValidationError{Code: validationCodeCaseInsensitive, Message: "does not match required pattern"}
		}

		return nil
	})
}

// MinItems validates minimum array/slice length
func MinItems(min int) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil
			}
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return nil
		}

		if rv.Len() < min {
			return &ValidationError{Code: validationCodeMinItems, Message: fmt.Sprintf("must have at least %d items", min)}
		}

		return nil
	})
}

// MaxItems validates maximum array/slice length
func MaxItems(max int) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil
			}
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return nil
		}

		if rv.Len() > max {
			return &ValidationError{Code: validationCodeMaxItems, Message: fmt.Sprintf("must have at most %d items", max)}
		}

		return nil
	})
}

// Unique validates that array/slice contains unique elements
func Unique() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil
			}
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return nil
		}

		seen := make(map[any]bool)
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i).Interface()
			if seen[elem] {
				return &ValidationError{Code: validationCodeUnique, Message: "array must contain unique elements"}
			}
			seen[elem] = true
		}

		return nil
	})
}

// MinMapKeys validates minimum map keys count
func MinMapKeys(min int) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil
			}
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Map {
			return nil
		}

		if rv.Len() < min {
			return &ValidationError{Code: validationCodeMinMapKeys, Message: fmt.Sprintf("must have at least %d keys", min)}
		}

		return nil
	})
}

// MaxMapKeys validates maximum map keys count
func MaxMapKeys(max int) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				return nil
			}
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Map {
			return nil
		}

		if rv.Len() > max {
			return &ValidationError{Code: validationCodeMaxMapKeys, Message: fmt.Sprintf("must have at most %d keys", max)}
		}

		return nil
	})
}

// CustomRule creates a custom validation rule from a function
func CustomRule(name string, fn func(value any) bool) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		if !fn(value) {
			return &ValidationError{Code: name, Message: fmt.Sprintf("failed custom validation: %s", name)}
		}

		return nil
	})
}

// Optional marks a rule as optional (allows nil or empty values)
func Optional(rule Rule) Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if ok && str == "" {
			return nil
		}

		return rule.Validate(value)
	})
}

// WithMessage creates a rule with a custom error message
func WithMessage(rule Rule, message string) Rule {
	return RuleFunc(func(value any) *ValidationError {
		err := rule.Validate(value)
		if err != nil {
			err.Message = message
		}
		return err
	})
}

// WithCode creates a rule with a custom error code
func WithCode(rule Rule, code string) Rule {
	return RuleFunc(func(value any) *ValidationError {
		err := rule.Validate(value)
		if err != nil {
			err.Code = code
		}
		return err
	})
}

// Security-focused validation rules

// SecureEmail performs security-focused email validation with additional checks.
//
// This rule adds security checks beyond basic email format validation:
// - Length limits (max 254 characters for email, max 64 for local part, max 253 for domain)
// - Rejects double dots (..)
// - Validates domain has at least one dot
//
// Example:
//
//	type UserInput struct {
//		Email string `validate:"secureEmail"`
//	}
func SecureEmail() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeSecureEmail, Message: "must be a string"}
		}

		str = strings.TrimSpace(str)
		if str == "" {
			return nil
		}

		// Basic length check to prevent DoS
		if len(str) > 254 {
			return &ValidationError{Code: validationCodeSecureEmail, Message: "email too long (max 254 characters)"}
		}

		// Check regex format
		if !emailPattern.MatchString(str) {
			return &ValidationError{Code: validationCodeSecureEmail, Message: "invalid email format"}
		}

		// Check for dangerous patterns
		if strings.Contains(str, "..") {
			return &ValidationError{Code: validationCodeSecureEmail, Message: "email contains consecutive dots"}
		}

		// Split and validate parts
		parts := strings.Split(str, "@")
		if len(parts) != 2 {
			return &ValidationError{Code: validationCodeSecureEmail, Message: "invalid email format"}
		}

		local, domain := parts[0], parts[1]

		// Validate local part
		if len(local) == 0 || len(local) > 64 {
			return &ValidationError{Code: validationCodeSecureEmail, Message: "email local part invalid length"}
		}

		// Validate domain part
		if len(domain) == 0 || len(domain) > 253 {
			return &ValidationError{Code: validationCodeSecureEmail, Message: "email domain invalid length"}
		}

		// Domain must contain at least one dot
		if !strings.Contains(domain, ".") {
			return &ValidationError{Code: validationCodeSecureEmail, Message: "email domain must contain a dot"}
		}

		return nil
	})
}

// SecureURL performs security-focused URL validation.
//
// This rule adds security checks to prevent common URL-based attacks:
// - Rejects dangerous schemes (javascript:, data:, file:, vbscript:)
// - Requires http/https for absolute URLs
// - Checks for null bytes
// - Length limit (max 2083 characters)
//
// Example:
//
//	type UserInput struct {
//		WebsiteURL string `validate:"secureUrl"`
//	}
func SecureURL() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeSecureURL, Message: "must be a string"}
		}

		str = strings.TrimSpace(str)
		if str == "" {
			return nil
		}

		// Basic length check
		if len(str) > 2083 {
			return &ValidationError{Code: validationCodeSecureURL, Message: "URL too long (max 2083 characters)"}
		}

		// Parse URL
		u, err := url.Parse(str)
		if err != nil {
			return &ValidationError{Code: validationCodeSecureURL, Message: "invalid URL format"}
		}

		// Reject dangerous schemes
		scheme := strings.ToLower(u.Scheme)
		switch scheme {
		case "javascript", "data", "vbscript", "file":
			return &ValidationError{Code: validationCodeSecureURL, Message: "dangerous URL scheme not allowed"}
		}

		// Require http or https for absolute URLs
		if u.IsAbs() {
			if scheme != "http" && scheme != "https" {
				return &ValidationError{Code: validationCodeSecureURL, Message: "only http and https schemes allowed"}
			}
		}

		// Check for null bytes
		if strings.Contains(str, "\x00") {
			return &ValidationError{Code: validationCodeSecureURL, Message: "URL contains null bytes"}
		}

		return nil
	})
}

// SecurePhone performs security-focused phone number validation.
//
// This rule validates phone numbers with security considerations:
// - Supports E.164 format (up to 15 digits)
// - Allows common formatting characters (spaces, dashes, parentheses, dots)
// - Rejects unexpected characters that could indicate injection attempts
// - Length validation (7-16 characters after formatting)
//
// Example:
//
//	type UserInput struct {
//		Phone string `validate:"securePhone"`
//	}
func SecurePhone() Rule {
	return RuleFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return &ValidationError{Code: validationCodeSecurePhone, Message: "must be a string"}
		}

		str = strings.TrimSpace(str)
		if str == "" {
			return nil
		}

		// Remove common formatting characters
		cleaned := strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' || r == '+' {
				return r
			}
			if r == ' ' || r == '-' || r == '(' || r == ')' || r == '.' {
				return -1
			}
			// Reject unexpected characters
			return 0
		}, str)

		// Check if any unexpected characters were found
		if strings.Contains(cleaned, string(rune(0))) {
			return &ValidationError{Code: validationCodeSecurePhone, Message: "phone contains invalid characters"}
		}

		// Basic length check (E.164 allows up to 15 digits)
		if len(cleaned) < 7 || len(cleaned) > 16 {
			return &ValidationError{Code: validationCodeSecurePhone, Message: "phone number length invalid"}
		}

		// Must start with + or digit
		if !strings.HasPrefix(cleaned, "+") && (cleaned[0] < '0' || cleaned[0] > '9') {
			return &ValidationError{Code: validationCodeSecurePhone, Message: "phone must start with + or digit"}
		}

		// If starts with +, must have at least one digit
		if strings.HasPrefix(cleaned, "+") && len(cleaned) < 2 {
			return &ValidationError{Code: validationCodeSecurePhone, Message: "phone number too short"}
		}

		return nil
	})
}
