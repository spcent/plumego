package websocket

import (
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

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
	// (0x00-0x1F and 0x7F, unless newlines or tabs are explicitly allowed).
	// Use SanitizeForLogging before logging accepted message content.
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
		AllowedNewlines:         false,
		AllowedTabs:             false,
	}
}

// ValidateTextMessage validates a text WebSocket message against the configured rules.
//
// This is transport-level text validation. It does not inspect business payloads
// for XSS, SQL, or application-specific content rules.
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

	// Check for control characters according to the configured transport policy.
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

			// Reject ASCII control characters (0x00-0x1F and 0x7F).
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

	// Replace all control characters, including newlines and tabs, with spaces.
	// Reuse a pooled strings.Builder to reduce allocator pressure.
	cleaned := sanitizerBuilderPool.Get().(*strings.Builder)
	cleaned.Reset()
	cleaned.Grow(len(s))
	for _, r := range s {
		if unicode.IsControl(r) {
			cleaned.WriteRune(' ')
			continue
		}
		cleaned.WriteRune(r)
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
