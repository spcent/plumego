package proxy

import (
	"net/http"
	"strings"
)

// PathRewriteFunc defines a function that rewrites the request path
type PathRewriteFunc func(path string) string

// StripPrefix returns a PathRewriteFunc that removes the given prefix from the path
//
// Example:
//
//	rewriter := StripPrefix("/api/v1")
//	// /api/v1/users -> /users
func StripPrefix(prefix string) PathRewriteFunc {
	prefix = strings.TrimSuffix(prefix, "/")
	return func(path string) string {
		return strings.TrimPrefix(path, prefix)
	}
}

// AddPrefix returns a PathRewriteFunc that adds the given prefix to the path
//
// Example:
//
//	rewriter := AddPrefix("/internal")
//	// /users -> /internal/users
func AddPrefix(prefix string) PathRewriteFunc {
	prefix = strings.TrimSuffix(prefix, "/")
	return func(path string) string {
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		return prefix + path
	}
}

// ReplacePrefix returns a PathRewriteFunc that replaces oldPrefix with newPrefix
//
// Example:
//
//	rewriter := ReplacePrefix("/api/v1", "/api/v2")
//	// /api/v1/users -> /api/v2/users
func ReplacePrefix(oldPrefix, newPrefix string) PathRewriteFunc {
	oldPrefix = strings.TrimSuffix(oldPrefix, "/")
	newPrefix = strings.TrimSuffix(newPrefix, "/")
	return func(path string) string {
		if strings.HasPrefix(path, oldPrefix) {
			return newPrefix + strings.TrimPrefix(path, oldPrefix)
		}
		return path
	}
}

// RewriteMap returns a PathRewriteFunc that uses a map for rewriting
//
// Example:
//
//	rewriter := RewriteMap(map[string]string{
//		"/old-api": "/new-api",
//		"/legacy":  "/v2",
//	})
func RewriteMap(rules map[string]string) PathRewriteFunc {
	return func(path string) string {
		for old, new := range rules {
			if strings.HasPrefix(path, old) {
				return new + strings.TrimPrefix(path, old)
			}
		}
		return path
	}
}

// RewriteRegex returns a PathRewriteFunc that uses regular expressions
// Note: This requires importing "regexp" package
//
// Example:
//
//	rewriter := RewriteRegex(regexp.MustCompile(`^/api/v(\d+)`), "/internal/v$1")
//	// /api/v1/users -> /internal/v1/users
type RegexRewriter interface {
	ReplaceAllString(src, repl string) string
}

func RewriteRegex(pattern RegexRewriter, replacement string) PathRewriteFunc {
	return func(path string) string {
		return pattern.ReplaceAllString(path, replacement)
	}
}

// Chain chains multiple PathRewriteFunc together
//
// Example:
//
//	rewriter := Chain(
//		StripPrefix("/api"),
//		AddPrefix("/internal"),
//	)
//	// /api/users -> /internal/users
func Chain(rewriters ...PathRewriteFunc) PathRewriteFunc {
	return func(path string) string {
		for _, rewriter := range rewriters {
			path = rewriter(path)
		}
		return path
	}
}

// applyPathRewrite applies path rewriting to the request
func applyPathRewrite(req *http.Request, rewriter PathRewriteFunc) {
	if rewriter == nil {
		return
	}

	originalPath := req.URL.Path
	newPath := rewriter(originalPath)

	if newPath != originalPath {
		req.URL.Path = newPath
		if req.URL.RawPath != "" {
			// Update RawPath if it was set
			req.URL.RawPath = newPath
		}
	}
}
