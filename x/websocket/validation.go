package websocket

import (
	"strings"
	"sync"
	"unicode/utf8"
)

const (
	MaxRoomNameLength  = 128
	RoomPasswordHeader = "X-Room-Password"
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

// ValidateRoomName validates a room identifier accepted by the handshake.
func ValidateRoomName(room string) error {
	if room == "" {
		return ErrInvalidRoomName
	}
	if len(room) > MaxRoomNameLength {
		return ErrInvalidRoomName
	}
	for i := 0; i < len(room); i++ {
		c := room[i]
		if c >= 'a' && c <= 'z' {
			continue
		}
		if c >= 'A' && c <= 'Z' {
			continue
		}
		if c >= '0' && c <= '9' {
			continue
		}
		switch c {
		case '-', '_', '.', ':':
			continue
		default:
			return ErrInvalidRoomName
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

	// Replace control characters with spaces, including newlines and tabs, so a
	// sanitized value always stays on one log line.
	// Reuse a pooled strings.Builder to reduce allocator pressure.
	cleaned := sanitizerBuilderPool.Get().(*strings.Builder)
	cleaned.Reset()
	cleaned.Grow(len(s))
	for _, r := range s {
		if r >= 0x20 && r != 0x7F {
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
