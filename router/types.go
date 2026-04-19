package router

import (
	"net/http"
)

// matchResult represents the result of route matching.
type matchResult struct {
	Handler      http.Handler
	ParamValues  []string
	ParamKeys    []string
	RoutePattern string
	RouteMethod  string
}

// RouteMeta describes route metadata used for reverse URL generation.
type RouteMeta struct {
	Name string `json:"name,omitempty"`
}

// RouteInfo describes a registered route with metadata.
type RouteInfo struct {
	Method string    `json:"method"`
	Path   string    `json:"path"`
	Meta   RouteMeta `json:"meta,omitempty"`
}

// NamedRoute stores information for reverse URL generation.
type NamedRoute struct {
	Method   string
	Pattern  string
	ParamPos map[string]int // parameter name -> position in pattern
}
