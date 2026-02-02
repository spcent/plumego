// Package stringsx provides string utility functions.
//
// This package offers helpful string manipulation utilities that complement
// the standard library's strings package. Functions focus on common use cases
// like secret masking, safe string handling, and text processing.
//
// Features:
//   - MaskSecret: Secure display of sensitive data
//   - (Additional utilities to be added)
//
// Example usage:
//
//	import "github.com/spcent/plumego/utils/stringsx"
//
//	// Mask a secret for logging
//	apiKey := "myapikey_1234567890abcdefghijklmn"
//	masked := stringsx.MaskSecret(apiKey)  // "my**************************mn"
//
//	// Safe for empty strings
//	empty := stringsx.MaskSecret("")       // ""
package stringsx

import "strings"

// MaskSecret masks a secret for display.
// Rules:
// - empty => ""
// - len<=4 => "****" (same length)
// - otherwise => keep first 2 and last 2, middle masked
func MaskSecret(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	n := len(s)
	if n <= 4 {
		return strings.Repeat("*", n)
	}
	return s[:2] + strings.Repeat("*", n-4) + s[n-2:]
}
