package input

import (
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// IsToken reports whether value is a valid HTTP token (RFC 7230).
//
// HTTP tokens are used in header names, media types, and other HTTP constructs.
// A token consists of ASCII letters, digits, and certain special characters.
//
// Example:
//
//	import "github.com/spcent/plumego/security/input"
//
//	if input.IsToken("Content-Type") {
//		// Valid HTTP token
//	}
//
// Valid tokens include:
//   - Letters (a-z, A-Z)
//   - Digits (0-9)
//   - Special characters: ! # $ % & ' * + - . ^ _ ` | ~
func IsToken(value string) bool {
	if value == "" {
		return false
	}

	for i := 0; i < len(value); i++ {
		ch := value[i]
		if !isTchar(ch) {
			return false
		}
	}

	return true
}

// IsHeaderName reports whether value is safe for use as an HTTP header name.
//
// Header names must be valid HTTP tokens according to RFC 7230.
//
// Example:
//
//	import "github.com/spcent/plumego/security/input"
//
//	if input.IsHeaderName("X-Custom-Header") {
//		// Valid header name
//	}
func IsHeaderName(value string) bool {
	return IsToken(value)
}

// IsHeaderValue reports whether value is safe for use as an HTTP header value.
// It rejects control characters that can lead to response splitting.
//
// Header values must be valid UTF-8 and must not contain:
//   - Carriage return (\r)
//   - Line feed (\n)
//   - Null character (\0)
//
// Example:
//
//	import "github.com/spcent/plumego/security/input"
//
//	if input.IsHeaderValue("some value") {
//		// Safe to use as header value
//	}
func IsHeaderValue(value string) bool {
	if !utf8.ValidString(value) {
		return false
	}

	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '\r', '\n', 0:
			return false
		}
	}

	return true
}

func isTchar(ch byte) bool {
	if ch >= '0' && ch <= '9' {
		return true
	}
	if ch >= 'a' && ch <= 'z' {
		return true
	}
	if ch >= 'A' && ch <= 'Z' {
		return true
	}

	switch ch {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	default:
		return false
	}
}

// ValidateEmail performs security-focused email validation.
//
// Returns true if the email appears valid. This is a basic check for security
// purposes, not a comprehensive email validation.
//
// Example:
//
//	import "github.com/spcent/plumego/security/input"
//
//	if !input.ValidateEmail(userEmail) {
//		// Reject invalid email
//	}
func ValidateEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}

	// Basic length check to prevent DoS
	if len(email) > 254 {
		return false
	}

	// Simple regex check
	// This is intentionally simple for security validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return false
	}

	// Check for dangerous patterns
	if strings.Contains(email, "..") {
		return false
	}

	// Split and validate parts
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	local, domain := parts[0], parts[1]

	// Validate local part
	if len(local) == 0 || len(local) > 64 {
		return false
	}

	// Validate domain part
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}

	// Domain must contain at least one dot
	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}

// ValidateURL performs security-focused URL validation.
//
// Returns true if the URL is valid and safe to use. This function checks for
// common security issues like javascript: URLs, file: URLs, and malformed URLs.
//
// Example:
//
//	import "github.com/spcent/plumego/security/input"
//
//	if !input.ValidateURL(userURL) {
//		// Reject invalid or unsafe URL
//	}
func ValidateURL(rawURL string) bool {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return false
	}

	// Basic length check
	if len(rawURL) > 2083 {
		return false
	}

	// Parse URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// Reject dangerous schemes
	scheme := strings.ToLower(u.Scheme)
	switch scheme {
	case "javascript", "data", "vbscript", "file":
		return false
	}

	// Require http or https for absolute URLs
	if u.IsAbs() {
		if scheme != "http" && scheme != "https" {
			return false
		}
	}

	// Check for null bytes
	if strings.Contains(rawURL, "\x00") {
		return false
	}

	return true
}

// ValidatePhone performs basic phone number validation.
//
// Returns true if the phone number appears valid. This is a basic check that
// accepts E.164 format and common phone number patterns.
//
// Example:
//
//	import "github.com/spcent/plumego/security/input"
//
//	if !input.ValidatePhone(userPhone) {
//		// Reject invalid phone
//	}
func ValidatePhone(phone string) bool {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return false
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
	}, phone)

	// Check if any unexpected characters were found
	if strings.Contains(cleaned, string(rune(0))) {
		return false
	}

	// Basic length check (E.164 allows up to 15 digits)
	if len(cleaned) < 7 || len(cleaned) > 16 {
		return false
	}

	// Must start with + or digit
	if !strings.HasPrefix(cleaned, "+") && (cleaned[0] < '0' || cleaned[0] > '9') {
		return false
	}

	// If starts with +, must have at least one digit
	if strings.HasPrefix(cleaned, "+") && len(cleaned) < 2 {
		return false
	}

	return true
}

// SanitizeHTML removes potentially dangerous HTML tags and attributes.
//
// This is a basic sanitizer that removes script tags, event handlers, and
// other dangerous HTML constructs. For production use with untrusted HTML,
// consider using a dedicated HTML sanitization library.
//
// Example:
//
//	import "github.com/spcent/plumego/security/input"
//
//	safe := input.SanitizeHTML(userInput)
func SanitizeHTML(s string) string {
	// Remove script tags and content
	scriptRe := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	s = scriptRe.ReplaceAllString(s, "")

	// Remove event handlers (onclick, onload, etc.)
	eventRe := regexp.MustCompile(`(?i)\s*on\w+\s*=\s*["'][^"']*["']`)
	s = eventRe.ReplaceAllString(s, "")

	// Remove javascript: URLs
	jsRe := regexp.MustCompile(`(?i)javascript:`)
	s = jsRe.ReplaceAllString(s, "")

	// Remove data: URLs (can be used for XSS)
	dataRe := regexp.MustCompile(`(?i)data:`)
	s = dataRe.ReplaceAllString(s, "")

	return s
}

// SanitizeSQL removes common SQL injection patterns.
//
// WARNING: This is NOT a substitute for parameterized queries. Always use
// parameterized queries for database operations. This function is only for
// defense-in-depth scenarios.
//
// Example:
//
//	import "github.com/spcent/plumego/security/input"
//
//	// Still use parameterized queries!
//	cleaned := input.SanitizeSQL(userInput)
func SanitizeSQL(s string) string {
	// Remove SQL comments
	s = regexp.MustCompile(`--.*`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`/\*.*?\*/`).ReplaceAllString(s, "")

	// Remove common SQL injection patterns
	dangerous := []string{
		";", "UNION", "SELECT", "INSERT", "UPDATE", "DELETE",
		"DROP", "CREATE", "ALTER", "EXEC", "EXECUTE",
	}

	for _, pattern := range dangerous {
		s = strings.ReplaceAll(s, pattern, "")
		s = strings.ReplaceAll(s, strings.ToLower(pattern), "")
	}

	return s
}

// StripControlChars removes ASCII control characters except newline and tab.
//
// This function removes characters that could cause display issues or
// be used for terminal injection attacks.
//
// Example:
//
//	import "github.com/spcent/plumego/security/input"
//
//	clean := input.StripControlChars(userInput)
func StripControlChars(s string) string {
	return strings.Map(func(r rune) rune {
		// Keep newline and tab
		if r == '\n' || r == '\t' {
			return r
		}
		// Remove other control characters
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
}

// ContainsDangerousChars checks for characters commonly used in injection attacks.
//
// Returns true if the string contains potentially dangerous characters.
//
// Example:
//
//	import "github.com/spcent/plumego/security/input"
//
//	if input.ContainsDangerousChars(userInput) {
//		// Handle potentially dangerous input
//	}
func ContainsDangerousChars(s string) bool {
	dangerous := []rune{0, '\r', '<', '>', '"', '\'', '`', '\\', '{', '}'}
	for _, r := range s {
		for _, d := range dangerous {
			if r == d {
				return true
			}
		}
	}
	return false
}

// TrimWhitespace removes leading and trailing whitespace and normalizes internal whitespace.
//
// This function collapses multiple consecutive whitespace characters into a
// single space and removes leading/trailing whitespace.
//
// Example:
//
//	import "github.com/spcent/plumego/security/input"
//
//	normalized := input.TrimWhitespace("  hello    world  ")
//	// Returns: "hello world"
func TrimWhitespace(s string) string {
	s = strings.TrimSpace(s)

	// Replace multiple spaces with single space
	spaceRe := regexp.MustCompile(`\s+`)
	s = spaceRe.ReplaceAllString(s, " ")

	return s
}
