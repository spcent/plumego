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
	current := rm.root
	if current == nil {
		return nil
	}

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
func (rm *RouteMatcher) findChildForPath(parent *node, path string) *node {
	// Check exact match first
	for _, child := range parent.children {
		if child.path == path {
			return child
		}
	}
	return nil
}

// findParamChild finds a param child node if exists
func (rm *RouteMatcher) findParamChild(parent *node) *node {
	for _, child := range parent.children {
		if len(child.path) > 0 && child.path[0] == ':' {
			return child
		}
	}
	return nil
}

// findWildChild finds a wildcard child node if exists
func (rm *RouteMatcher) findWildChild(parent *node) *node {
	for _, child := range parent.children {
		if len(child.path) > 0 && child.path[0] == '*' {
			return child
		}
	}
	return nil
}
