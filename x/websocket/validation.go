package websocket

import (
	"bytes"
	"strings"
	"sync"
	"unicode/utf8"
)

// sqlInjectionPatterns is a pre-allocated slice of lowercase SQL injection indicator
// patterns used by ContainsDangerousPatterns. Defined once at package level to avoid
// a heap allocation on every call.
var sqlInjectionPatterns = [][]byte{
	[]byte("union select"), []byte("union all select"),
	[]byte("; drop "), []byte("; delete "), []byte("; update "), []byte("; insert "),
	[]byte("' or '1'='1"), []byte("\" or \"1\"=\"1"),
	[]byte("'; --"), []byte("\"; --"),
}

// sanitizerBuilderPool reuses strings.Builder instances in SanitizeForLogging to
// reduce GC pressure in high-log-volume scenarios.
var sanitizerBuilderPool = sync.Pool{
	New: func() any { return new(strings.Builder) },
}

// MessageValidationConfig defines validation rules for WebSocket messages.
type MessageValidationConfig struct {
	// MaxLength is the maximum allowed message length in bytes.
	// 0 means no limit.
	MaxLength int

	// AllowEmpty allows empty messages.
	AllowEmpty bool

	// RejectControlCharacters rejects messages containing ASCII control characters
	// (0x00-0x1F except newline/tab, and 0x7F).
	// This prevents log injection, terminal manipulation, and other attacks.
	RejectControlCharacters bool

	// RequireValidUTF8 requires all text messages to be valid UTF-8.
	RequireValidUTF8 bool

	// AllowedNewlines allows newline characters (\n, \r) even when RejectControlCharacters is true.
	AllowedNewlines bool

	// AllowedTabs allows tab characters (\t) even when RejectControlCharacters is true.
	AllowedTabs bool
}

// DefaultMessageValidationConfig returns a secure default configuration.
func DefaultMessageValidationConfig() MessageValidationConfig {
	return MessageValidationConfig{
		MaxLength:               1 << 20, // 1MB
		AllowEmpty:              false,
		RejectControlCharacters: true,
		RequireValidUTF8:        true,
		AllowedNewlines:         true,
		AllowedTabs:             true,
	}
}

// ValidateTextMessage validates a text WebSocket message against the configured rules.
//
// This helps prevent:
// - XSS attacks (by rejecting dangerous control characters)
// - Log injection attacks (by rejecting newlines/control characters in logs)
// - Terminal manipulation attacks (by rejecting ANSI escape sequences)
// - Invalid UTF-8 that could cause parsing errors
//
// Example:
//
//	cfg := DefaultMessageValidationConfig()
//	if err := ValidateTextMessage(msg, cfg); err != nil {
//	    return err // reject message
//	}
func ValidateTextMessage(data []byte, cfg MessageValidationConfig) error {
	// Check length
	if cfg.MaxLength > 0 && len(data) > cfg.MaxLength {
		return ErrMessageTooLong
	}

	// Check empty
	if !cfg.AllowEmpty && len(data) == 0 {
		return ErrEmptyMessage
	}

	// Check UTF-8 validity
	if cfg.RequireValidUTF8 && !utf8.Valid(data) {
		return ErrInvalidUTF8
	}

	// Check for dangerous control characters
	if cfg.RejectControlCharacters {
		for i := 0; i < len(data); i++ {
			c := data[i]

			// Allow specific characters based on configuration
			if c == '\n' || c == '\r' {
				if cfg.AllowedNewlines {
					continue
				}
			}
			if c == '\t' {
				if cfg.AllowedTabs {
					continue
				}
			}

			// Reject ASCII control characters (0x00-0x1F and 0x7F)
			// These can be used for:
			// - ANSI escape sequences (terminal manipulation)
			// - Log injection (newlines in logs)
			// - Null byte injection
			// - Other protocol-level attacks
			if c < 0x20 || c == 0x7F {
				return ErrControlCharacters
			}
		}
	}

	return nil
}

// SanitizeForLogging sanitizes a message for safe logging.
//
// This function:
// - Truncates long messages
// - Removes/replaces control characters
// - Ensures valid UTF-8
//
// Use this before logging user-provided WebSocket messages to prevent log injection.
//
// Example:
//
//	safe := SanitizeForLogging(msg, 200)
//	log.Printf("Received message: %s", safe)
func SanitizeForLogging(data []byte, maxLen int) string {
	if maxLen <= 0 {
		maxLen = 200
	}

	// Remember if we truncated
	originalLen := len(data)
	truncated := originalLen > maxLen

	// Truncate if needed
	if truncated {
		data = data[:maxLen]
	}

	// Ensure valid UTF-8
	s := string(data)
	if !utf8.ValidString(s) {
		// Replace invalid UTF-8 sequences
		s = strings.ToValidUTF8(s, "�")
	}

	// Replace control characters with spaces (except newlines and tabs).
	// Reuse a pooled strings.Builder to reduce allocator pressure.
	cleaned := sanitizerBuilderPool.Get().(*strings.Builder)
	cleaned.Reset()
	cleaned.Grow(len(s))
	for _, r := range s {
		// Keep printable characters, newlines, and tabs
		if r >= 0x20 && r != 0x7F {
			cleaned.WriteRune(r)
		} else if r == '\n' || r == '\t' {
			cleaned.WriteRune(r)
		} else {
			cleaned.WriteRune(' ')
		}
	}

	// Append truncation marker inside the builder before extracting the string
	// so we avoid a second string allocation from result += "...".
	if truncated {
		cleaned.WriteString("...")
	}
	result := cleaned.String()
	sanitizerBuilderPool.Put(cleaned)
	return result
}

// ContainsDangerousPatterns checks if a message contains patterns that could indicate
// an attack attempt.
//
// Detects:
// - ANSI escape sequences
// - HTML/XML tags (potential XSS if message is rendered)
// - JavaScript event handlers
// - SQL keywords (if message is used in queries)
//
// This is a heuristic check and may have false positives.
// Use for additional security layers, not as the only validation.
func ContainsDangerousPatterns(data []byte) bool {
	// Check for ANSI escape sequences directly on raw bytes (no allocation).
	if bytes.Contains(data, []byte("\x1b[")) {
		return true
	}

	// Lowercase once for all case-insensitive checks below.
	// bytes.ToLower allocates a single copy — cheaper than string(data)+strings.ToLower.
	lower := bytes.ToLower(data)

	// Check for HTML/XML tags and dangerous JS patterns (case-insensitive)
	if bytes.Contains(lower, []byte("<script")) || bytes.Contains(lower, []byte("</script")) ||
		bytes.Contains(lower, []byte("<iframe")) || bytes.Contains(lower, []byte("javascript:")) ||
		bytes.Contains(lower, []byte("onerror=")) || bytes.Contains(lower, []byte("onload=")) {
		return true
	}

	// Check for SQL injection patterns (basic detection)
	for _, keyword := range sqlInjectionPatterns {
		if bytes.Contains(lower, keyword) {
			return true
		}
	}

	return false
}
