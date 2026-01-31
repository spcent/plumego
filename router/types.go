package router

import (
	"sync/atomic"

	"github.com/spcent/plumego/middleware"
)

// MatchResult represents the result of route matching
type MatchResult struct {
	Handler          Handler
	ParamValues      []string
	ParamKeys        []string
	RouteMiddlewares []middleware.Middleware // Now using direct type instead of any
	RoutePattern     string
	RouteMethod      string
	cache            atomic.Value
}

type cachedHandler struct {
	version uint64
	handler Handler
}

func (mr *MatchResult) loadCached(version uint64) (Handler, bool) {
	if mr == nil {
		return nil, false
	}
	val := mr.cache.Load()
	if val == nil {
		return nil, false
	}
	entry, ok := val.(cachedHandler)
	if !ok || entry.handler == nil {
		return nil, false
	}
	if entry.version != version {
		return nil, false
	}
	return entry.handler, true
}

func (mr *MatchResult) storeCached(version uint64, handler Handler) {
	if mr == nil || handler == nil {
		return
	}
	mr.cache.Store(cachedHandler{version: version, handler: handler})
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
