package router

import (
	"strings"
)

// RouteMatcher performs efficient trie-based route matching.
// This is a lightweight matcher used during HTTP request processing for
// non-cached route lookups. It traverses the radix tree structure to find
// the best matching route for a given URL path.
//
// Matching strategy:
//  1. Try static path segments first (exact match)
//  2. Try parameter segments (dynamic segments like ":id")
//  3. Try wildcard segments (catch-all segments like "*path")
//
// Example:
//
//	tree := &node{...} // Radix tree structure
//	matcher := NewRouteMatcher(tree)
//	result := matcher.Match([]string{"users", "123"})
//	// result.Handler contains the matched handler
//	// result.ParamValues contains ["123"]
type RouteMatcher struct {
	root *node
}

// NewRouteMatcher creates a new route matcher for the given tree root.
// The matcher is lightweight and can be reused for multiple matching operations.
// This is primarily used internally by the Router for route matching.
//
// Parameters:
//   - root: Root node of the radix tree
//
// Returns:
//   - *RouteMatcher: A new route matcher instance
//
// Example:
//
//	// Internal usage within Router.handleRouteMatch:
//	matcher := NewRouteMatcher(tree)
//	result := matcher.Match([]string{"users", "123"})
func NewRouteMatcher(root *node) *RouteMatcher {
	return &RouteMatcher{
		root: root,
	}
}

// Match performs route matching against the given path parts.
// This is the main matching method that traverses the radix tree.
//
// Parameters:
//   - parts: URL path segments (e.g., ["users", "123"])
//
// Returns:
//   - *MatchResult: Match result with handler and parameters, or nil if no match
//
// Example:
//
//	matcher := NewRouteMatcher(root)
//	result := matcher.Match([]string{"users", "123"})
//	if result != nil {
//	    handler := result.Handler
//	    params := result.ParamValues // ["123"]
//	}
func (rm *RouteMatcher) Match(parts []string) *MatchResult {
	if rm.root == nil {
		return nil
	}

	current := rm.root
	paramValues := make([]string, 0, len(parts))

	for i, pathSegment := range parts {
		// Reject empty path segments (e.g., from double slashes /users//123)
		// Empty segments are only valid for wildcards
		if pathSegment == "" {
			// Try wildcard match for empty segment
			if wildChild := rm.findWildChild(current); wildChild != nil {
				wildValue := strings.Join(parts[i:], "/")
				paramValues = append(paramValues, wildValue)
				current = wildChild
				break
			}
			// Empty segment with no wildcard = no match
			return nil
		}

		// Try to find exact match first
		if child := rm.findChildForPath(current, pathSegment); child != nil {
			current = child
			continue
		}

		// Try param match
		if paramChild := rm.findParamChild(current); paramChild != nil {
			paramValues = append(paramValues, pathSegment)
			current = paramChild
			continue
		}

		// Try wildcard match
		if wildChild := rm.findWildChild(current); wildChild != nil {
			wildValue := strings.Join(parts[i:], "/")
			paramValues = append(paramValues, wildValue)
			current = wildChild
			break
		}

		// No match found
		return nil
	}

	// Check if we found a valid handler
	if current == nil || current.handler == nil {
		return nil
	}

	// Return match result with direct middleware slice
	return &MatchResult{
		Handler:          current.handler,
		ParamValues:      paramValues,
		ParamKeys:        current.paramKeys,
		RouteMiddlewares: current.middlewares,
		RoutePattern:     current.fullPath,
	}
}

// findChildForPath finds a child node that matches the given path segment.
// This method uses multiple optimization strategies for different scenarios:
//   - Uses indices for fast lookup when available
//   - Falls back to linear search for small sets
//   - Optimized loop for larger sets
//
// Parameters:
//   - parent: Parent node to search in
//   - path: Path segment to match
//
// Returns:
//   - *node: Matching child node, or nil if not found
func (rm *RouteMatcher) findChildForPath(parent *node, path string) *node {
	// Fast path: empty parent or no children
	if parent == nil || len(parent.children) == 0 {
		return nil
	}

	numChildren := len(parent.children)

	// For very small sets (1-2 children), linear search is faster than index lookup
	if numChildren <= 2 {
		for i := 0; i < numChildren; i++ {
			if parent.children[i].path == path {
				return parent.children[i]
			}
		}
		return nil
	}

	// Use indices for faster lookup when available
	if len(parent.indices) > 0 && len(path) > 0 {
		firstChar := path[0]

		// Find the first occurrence of the character in indices
		idx := strings.IndexByte(parent.indices, firstChar)
		if idx == -1 {
			return nil
		}

		// Only scan children that start with the same character
		// Since indices are sorted, all matching chars are consecutive
		for i := idx; i < numChildren && parent.indices[i] == firstChar; i++ {
			if parent.children[i].path == path {
				return parent.children[i]
			}
		}
		return nil
	}

	// Fallback to linear search when indices aren't available
	for i := 0; i < numChildren; i++ {
		if parent.children[i].path == path {
			return parent.children[i]
		}
	}
	return nil
}

// findChildByByte finds a child node whose path starts with the given byte.
// Uses indices for O(1) lookup when available, falls back to linear search.
func (rm *RouteMatcher) findChildByByte(parent *node, b byte) *node {
	if parent == nil || len(parent.children) == 0 {
		return nil
	}

	// Use indices for fast lookup if available
	if len(parent.indices) > 0 {
		idx := strings.IndexByte(parent.indices, b)
		if idx >= 0 && idx < len(parent.children) {
			return parent.children[idx]
		}
		return nil
	}

	// Fallback to linear search
	for i := range parent.children {
		if len(parent.children[i].path) > 0 && parent.children[i].path[0] == b {
			return parent.children[i]
		}
	}
	return nil
}

// findParamChild finds a param child node (path starting with ":").
func (rm *RouteMatcher) findParamChild(parent *node) *node {
	return rm.findChildByByte(parent, ':')
}

// findWildChild finds a wildcard child node (path starting with "*").
func (rm *RouteMatcher) findWildChild(parent *node) *node {
	return rm.findChildByByte(parent, '*')
}
