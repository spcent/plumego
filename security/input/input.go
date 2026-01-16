package input

import "unicode/utf8"

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
