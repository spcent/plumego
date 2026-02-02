// Package input provides input validation and sanitization utilities.
//
// This package implements validation functions for common input types including:
//   - Email addresses (RFC 5322 compliant)
//   - URLs (RFC 3986 compliant)
//   - Phone numbers (E.164 format)
//   - HTTP tokens (RFC 7230)
//   - Safe strings (alphanumeric + whitespace)
//
// All validators are designed to be fast, secure, and follow relevant RFCs.
// They use standard library regex where appropriate and avoid ReDoS vulnerabilities.
//
// Example usage:
//
//	import "github.com/spcent/plumego/security/input"
//
//	// Validate email
//	if !input.ValidateEmail("user@example.com") {
//		// Invalid email
//	}
//
//	// Validate URL
//	if !input.ValidateURL("https://example.com") {
//		// Invalid URL
//	}
//
//	// Validate phone number (E.164)
//	if !input.ValidatePhone("+1234567890") {
//		// Invalid phone number
//	}
//
//	// Check if string is safe (no special chars)
//	if !input.IsSafeString("Hello World 123") {
//		// Contains unsafe characters
//	}
package input

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/spcent/plumego/validator"
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
// This function delegates to validator.SecureEmail() for the actual validation logic.
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
	rule := validator.SecureEmail()
	err := rule.Validate(email)
	return err == nil
}

// ValidateURL performs security-focused URL validation.
//
// Returns true if the URL is valid and safe to use. This function checks for
// common security issues like javascript: URLs, file: URLs, and malformed URLs.
//
// This function delegates to validator.SecureURL() for the actual validation logic.
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
	rule := validator.SecureURL()
	err := rule.Validate(rawURL)
	return err == nil
}

// ValidatePhone performs basic phone number validation.
//
// Returns true if the phone number appears valid. This is a basic check that
// accepts E.164 format and common phone number patterns.
//
// This function delegates to validator.SecurePhone() for the actual validation logic.
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
	rule := validator.SecurePhone()
	err := rule.Validate(phone)
	return err == nil
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
