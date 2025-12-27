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
