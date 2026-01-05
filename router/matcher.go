package router

import (
	"strings"

	"github.com/spcent/plumego/middleware"
)

// RouteMatcher performs efficient trie-based route matching
type RouteMatcher struct {
	root *node
}

// NewRouteMatcher creates a new route matcher for the given tree root
func NewRouteMatcher(root *node) *RouteMatcher {
	return &RouteMatcher{
		root: root,
	}
}

// MatchResult represents the result of route matching
type MatchResult struct {
	Handler          Handler
	ParamValues      []string
	ParamKeys        []string
	RouteMiddlewares []middleware.Middleware
}

// Match performs route matching against the given path parts
func (rm *RouteMatcher) Match(parts []string) *MatchResult {
	if rm.root == nil {
		return nil
	}

	current := rm.root
	paramValues := make([]string, 0, len(parts))

	for i, pathSegment := range parts {
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

	// Use the paramKeys stored in the node during route registration
	return &MatchResult{
		Handler:          current.handler,
		ParamValues:      paramValues,
		ParamKeys:        current.paramKeys,
		RouteMiddlewares: current.middlewares,
	}
}

// findChildForPath finds a child node that matches the given path segment
// Optimized version with better lookup performance
func (rm *RouteMatcher) findChildForPath(parent *node, path string) *node {
	// Fast path: empty parent or no children
	if parent == nil || len(parent.children) == 0 {
		return nil
	}

	// Use indices for faster lookup when available
	if len(parent.indices) > 0 && len(path) > 0 {
		firstChar := path[0]

		// Binary search in indices for better performance
		idx := strings.IndexByte(parent.indices, firstChar)
		if idx == -1 {
			return nil
		}

		// Check the corresponding child
		// Since indices are sorted, we can find the child more efficiently
		for _, child := range parent.children {
			if child.path == path {
				return child
			}
		}
		return nil
	}

	// Fallback to linear search for small sets or when indices aren't available
	// For very small sets (1-2 children), linear search is faster
	if len(parent.children) <= 2 {
		for _, child := range parent.children {
			if child.path == path {
				return child
			}
		}
		return nil
	}

	// For larger sets, optimize the loop
	for i := range parent.children {
		if parent.children[i].path == path {
			return parent.children[i]
		}
	}
	return nil
}

// findParamChild finds a param child node if exists
// Optimized to avoid unnecessary allocations
func (rm *RouteMatcher) findParamChild(parent *node) *node {
	if parent == nil || len(parent.children) == 0 {
		return nil
	}

	for i := range parent.children {
		child := parent.children[i]
		if len(child.path) > 0 && child.path[0] == ':' {
			return child
		}
	}
	return nil
}

// findWildChild finds a wildcard child node if exists
// Optimized to avoid unnecessary allocations
func (rm *RouteMatcher) findWildChild(parent *node) *node {
	if parent == nil || len(parent.children) == 0 {
		return nil
	}

	for i := range parent.children {
		child := parent.children[i]
		if len(child.path) > 0 && child.path[0] == '*' {
			return child
		}
	}
	return nil
}
