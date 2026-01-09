package router

import (
	"github.com/spcent/plumego/middleware"
)

// MatchResult represents the result of route matching
type MatchResult struct {
	Handler          Handler
	ParamValues      []string
	ParamKeys        []string
	RouteMiddlewares []middleware.Middleware // Now using direct type instead of any
}

// RouteMeta describes route metadata for debugging/observability.
type RouteMeta struct {
	Name string   `json:"name,omitempty"`
	Tags []string `json:"tags,omitempty"`
}

// RouteInfo describes a registered route with metadata.
type RouteInfo struct {
	Method string    `json:"method"`
	Path   string    `json:"path"`
	Meta   RouteMeta `json:"meta,omitempty"`
}
