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

// Match performs route matching against the given path parts.
// This is the main matching method that traverses the radix tree.
//
// Parameters:
//   - parts: URL path segments (e.g., ["users", "123"])
//
// Returns:
//   - *matchResult: Match result with handler and parameters, or nil if no match
//
// Example:
//
//	matcher := &routeMatcher{root: root}
//	result := matcher.Match([]string{"users", "123"})
//	if result != nil {
//	    handler := result.Handler
//	    params := result.ParamValues // ["123"]
//	}
func (rm *routeMatcher) Match(parts []string) *matchResult {
	if rm.root == nil {
		return nil
	}

	current := rm.root
	// Use pooled slice for parameter values to avoid per-request allocation
	pvPtr := getParamValues()
	paramValues := *pvPtr

	for i, pathSegment := range parts {
		// Reject empty path segments (e.g., from double slashes /users//123).
		if pathSegment == "" {
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

	return &matchResult{
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
	return findStaticChild(parent, path)
}

// findParamChild finds a param child node (path starting with ":").
func (rm *routeMatcher) findParamChild(parent *node) *node {
	return findParamChild(parent)
}

// findWildChild finds a wildcard child node (path starting with "*").
func (rm *routeMatcher) findWildChild(parent *node) *node {
	return findWildChild(parent)
}
