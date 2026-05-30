package diagnostics

import (
	"regexp"
)

// compileRegexp compiles a regular expression. Returns an error if invalid.
func compileRegexp(pattern string) (*regexp.Regexp, error) {
	return regexp.Compile(pattern)
}

// regexpQuote escapes special regex characters in a string using regexp.QuoteMeta.
func regexpQuote(s string) string {
	return regexp.QuoteMeta(s)
}
