package router

import (
	"strings"
)

// routeMatcher performs efficient trie-based route matching.
// It traverses the radix tree structure to find the best matching route for a given URL path.
//
// Matching strategy:
//  1. Try static path segments first (exact match)
//  2. Try parameter segments (dynamic segments like ":id")
//  3. Try wildcard segments (catch-all segments like "*path")
type routeMatcher struct {
	root *node
}

// newRouteMatcher creates a new route matcher for the given tree root.
func newRouteMatcher(root *node) *routeMatcher {
	return &routeMatcher{
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
//	matcher := newRouteMatcher(root)
//	result := matcher.Match([]string{"users", "123"})
//	if result != nil {
//	    handler := result.Handler
//	    params := result.ParamValues // ["123"]
//	}
func (rm *routeMatcher) Match(parts []string) *MatchResult {
	if rm.root == nil {
		return nil
	}

	current := rm.root
	// Use pooled slice for parameter values to avoid per-request allocation
	pvPtr := getParamValues()
	paramValues := *pvPtr

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
			*pvPtr = paramValues
			putParamValues(pvPtr)
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
		*pvPtr = paramValues
		putParamValues(pvPtr)
		return nil
	}

	// Check if we found a valid handler
	if current == nil || current.handler == nil {
		*pvPtr = paramValues
		putParamValues(pvPtr)
		return nil
	}

	// Copy param values out of the pooled slice so we can return it to the pool.
	// For zero params this is a nil slice (no allocation).
	var resultParams []string
	if len(paramValues) > 0 {
		resultParams = make([]string, len(paramValues))
		copy(resultParams, paramValues)
	}
	*pvPtr = paramValues
	putParamValues(pvPtr)

	return &MatchResult{
		Handler:      current.handler,
		ParamValues:  resultParams,
		ParamKeys:    current.paramKeys,
		RoutePattern: current.fullPath,
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
func (rm *routeMatcher) findChildForPath(parent *node, path string) *node {
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
func (rm *routeMatcher) findChildByByte(parent *node, b byte) *node {
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
func (rm *routeMatcher) findParamChild(parent *node) *node {
	return rm.findChildByByte(parent, ':')
}

// findWildChild finds a wildcard child node (path starting with "*").
func (rm *routeMatcher) findWildChild(parent *node) *node {
	return rm.findChildByByte(parent, '*')
}
