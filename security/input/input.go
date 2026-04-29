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
//	// Check if string contains dangerous characters
//	if input.ContainsDangerousChars(userInput) {
//		// Contains dangerous characters
//	}
package input

import (
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	htmlScriptTagRe     = regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script\s*>`)
	htmlEventHandlerRe  = regexp.MustCompile(`(?i)\s+on[a-z0-9_-]+\s*=\s*(?:"[^"]*"|'[^']*'|[^\s>]+)`)
	htmlJavaScriptURLRe = regexp.MustCompile(`(?i)javascript:`)
	htmlDataURLRe       = regexp.MustCompile(`(?i)data:`)
	sqlLineCommentRe    = regexp.MustCompile(`--.*`)
	sqlBlockCommentRe   = regexp.MustCompile(`/\*.*?\*/`)
	sqlKeywordRe        = regexp.MustCompile(`(?i)\b(?:UNION|SELECT|INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|EXECUTE|EXEC)\b`)
	whitespaceRe        = regexp.MustCompile(`\s+`)
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
// Header values must be valid UTF-8 and must not contain ASCII control
// characters except horizontal tab.
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
		ch := value[i]
		if ch == '\t' {
			continue
		}
		if ch < 0x20 || ch == 0x7f {
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
	if len(email) > 254 || strings.ContainsRune(email, 0) || !utf8.ValidString(email) {
		return false
	}

	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return false
	}

	local, domain, ok := strings.Cut(email, "@")
	if !ok || local == "" || domain == "" {
		return false
	}
	if len(local) > 64 || strings.Contains(local, "..") || strings.Contains(domain, "..") {
		return false
	}
	if strings.HasPrefix(local, ".") || strings.HasSuffix(local, ".") {
		return false
	}

	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return false
	}
	for _, label := range labels {
		if !isEmailDomainLabel(label) {
			return false
		}
	}
	return true
}

func isEmailDomainLabel(label string) bool {
	if label == "" || len(label) > 63 {
		return false
	}
	if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
		return false
	}
	for i := 0; i < len(label); i++ {
		ch := label[i]
		if ch >= 'a' && ch <= 'z' {
			continue
		}
		if ch >= 'A' && ch <= 'Z' {
			continue
		}
		if ch >= '0' && ch <= '9' {
			continue
		}
		if ch == '-' {
			continue
		}
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
	if len(rawURL) > 2048 || strings.ContainsRune(rawURL, 0) || !utf8.ValidString(rawURL) {
		return false
	}

	lowerURL := strings.ToLower(rawURL)
	for _, blocked := range []string{"javascript:", "data:", "file:", "vbscript:"} {
		if strings.HasPrefix(lowerURL, blocked) {
			return false
		}
	}

	if strings.HasPrefix(rawURL, "/") {
		return !strings.HasPrefix(rawURL, "//")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	if parsed.User != nil {
		return false
	}
	host := parsed.Hostname()
	if host == "" {
		return false
	}
	if parsed.Port() != "" {
		if _, err := net.LookupPort(parsed.Scheme, parsed.Port()); err != nil {
			return false
		}
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
	if !utf8.ValidString(phone) {
		return false
	}

	digits := make([]rune, 0, len(phone))
	for i, r := range phone {
		switch {
		case r >= '0' && r <= '9':
			digits = append(digits, r)
		case r == '+':
			if i != 0 {
				return false
			}
		case r == ' ' || r == '-' || r == '.' || r == '(' || r == ')':
		default:
			return false
		}
	}

	return len(digits) >= 7 && len(digits) <= 15
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
	s = htmlScriptTagRe.ReplaceAllString(s, "")

	// Remove event handlers (onclick, onload, etc.)
	s = htmlEventHandlerRe.ReplaceAllString(s, "")

	// Remove javascript: URLs
	s = htmlJavaScriptURLRe.ReplaceAllString(s, "")

	// Remove data: URLs (can be used for XSS)
	s = htmlDataURLRe.ReplaceAllString(s, "")

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
	s = sqlLineCommentRe.ReplaceAllString(s, "")
	s = sqlBlockCommentRe.ReplaceAllString(s, "")

	s = strings.ReplaceAll(s, ";", "")
	s = sqlKeywordRe.ReplaceAllString(s, "")

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
	s = whitespaceRe.ReplaceAllString(s, " ")

	return s
}
