package logging

import (
	"io"
	"regexp"
	"strings"
)

// RedactingWriter wraps an io.Writer and redacts sensitive information from log lines.
// It scans each line for patterns like Bearer tokens, API keys, passwords, etc.
// and replaces them with <redacted> placeholders.
type RedactingWriter struct {
	writer io.Writer
}

// NewRedactingWriter creates a new RedactingWriter that wraps the given writer.
func NewRedactingWriter(w io.Writer) *RedactingWriter {
	return &RedactingWriter{writer: w}
}

// Write implements io.Writer. It redacts sensitive patterns before writing to the underlying writer.
func (rw *RedactingWriter) Write(p []byte) (n int, err error) {
	redacted := redactLogLine(string(p))
	return rw.writer.Write([]byte(redacted))
}

// sensitivePatterns defines regex patterns for sensitive data that should be redacted.
var sensitivePatterns = []*regexp.Regexp{
	// Bearer tokens
	regexp.MustCompile(`(?i)bearer\s+[a-z0-9\-._~+/]+=*`),
	// API keys and tokens
	regexp.MustCompile(`(?i)(api[_-]?key|api[_-]?token|access[_-]?key|secret[_-]?key|auth[_-]?token)\s*[=:]\s*[a-z0-9\-._~+/]+=*`),
	// Passwords
	regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[=:]\s*[^\s]+`),
	// Email addresses (optional, can be noisy)
	// regexp.MustCompile(`[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}`),
}

// redactLogLine redacts sensitive information from a log line.
func redactLogLine(line string) string {
	result := line

	for _, pattern := range sensitivePatterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			// Find where the sensitive value starts (after key=, key:, or bearer)
			lower := strings.ToLower(match)

			// Handle bearer tokens
			if strings.Contains(lower, "bearer") {
				parts := strings.SplitN(match, " ", 2)
				if len(parts) == 2 {
					return parts[0] + " <redacted>"
				}
				return "Bearer <redacted>"
			}

			// Handle key=value or key: value patterns
			for _, sep := range []string{"=", ": ", " "} {
				if idx := strings.Index(match, sep); idx >= 0 {
					return match[:idx+len(sep)] + "<redacted>"
				}
			}

			return "<redacted>"
		})
	}

	return result
}
