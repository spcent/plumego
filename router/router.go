package router

import (
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/spcent/plumego/contract"
)

const methodAny = "ANY"

// Configuration defaults (unexported; callers configure via RouterOption).
const (
	defaultCacheCapacity = 100
	defaultMaxParams     = 8
	defaultPoolSliceCap  = 4
	defaultPathPartsCap  = 8
)

// segment represents a path segment with type information.
type segment struct {
	raw       string
	isParam   bool
	isWild    bool
	paramName string
}

// node represents a node in the prefix trie.
type node struct {
	path      string
	paramName string
	fullPath  string
	indices   string
	children  []*node
	handler   http.Handler
	paramKeys []string
}

type route struct {
	Method string
	Path   string
}

// routerState is the shared mutable state for the root router and all groups.
// It excludes application-layer concerns so the router stays a pure
// route-structure primitive.
type routerState struct {
	trees            map[string]*node
	routes           map[string][]route
	frozen           bool
	mu               sync.RWMutex
	matchCache       *matchCache
	routeMeta        map[string]map[string]RouteMeta
	namedRoutes      map[string]*NamedRoute
	methodNotAllowed atomic.Bool
}

// Router represents an HTTP router with path-based routing.
// It implements http.Handler, making it compatible with any Go HTTP server.
//
// Example:
//
//	r := router.NewRouter()
//	if err := r.AddRoute(http.MethodGet, "/users", listUsersHandler); err != nil {
//	    return err
//	}
//	if err := r.AddRoute(http.MethodPost, "/users", createUserHandler); err != nil {
//	    return err
//	}
//
//	api := r.Group("/api/v1")
//	if err := api.AddRoute(http.MethodGet, "/users/:id", getUserHandler); err != nil {
//	    return err
//	}
//
//	http.ListenAndServe(":8080", r)
type Router struct {
	prefix string
	parent *Router
	state  *routerState
}

// RouterOption defines a function type for router configuration options.
type RouterOption func(*Router)

// RouteOption defines an option for route metadata.
type RouteOption func(*RouteMeta)

// WithRouteName sets a route name for reverse URL generation.
//
// Example:
//
//	r.AddRoute(http.MethodGet, "/users/:id", handler, router.WithRouteName("user.show"))
//	url := r.URL("user.show", "id", "123") // → "/users/123"
func WithRouteName(name string) RouteOption {
	return func(meta *RouteMeta) {
		meta.Name = name
	}
}

// WithMethodNotAllowed enables returning 405 with Allow header when path matches another method.
func WithMethodNotAllowed(enabled bool) RouterOption {
	return func(r *Router) {
		r.state.methodNotAllowed.Store(enabled)
	}
}

// NewRouter creates a new Router instance with default configuration.
//
// Example:
//
//	r := router.NewRouter()
//	err := r.AddRoute(http.MethodGet, "/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Hello, World!"))
//	}))
func NewRouter(opts ...RouterOption) *Router {
	r := &Router{
		prefix: "",
		parent: nil,
		state: &routerState{
			trees:       make(map[string]*node),
			routes:      make(map[string][]route),
			routeMeta:   make(map[string]map[string]RouteMeta),
			namedRoutes: make(map[string]*NamedRoute),
			matchCache:  newMatchCache(defaultCacheCapacity),
		},
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// SetMethodNotAllowed toggles 405 responses when another method matches the path.
func (r *Router) SetMethodNotAllowed(enabled bool) {
	r.state.methodNotAllowed.Store(enabled)
}

// MethodNotAllowedEnabled reports whether 405 handling is enabled.
func (r *Router) MethodNotAllowedEnabled() bool {
	return r.state.methodNotAllowed.Load()
}

// Freeze prevents the router from accepting new route registrations.
func (r *Router) Freeze() {
	r.state.mu.Lock()
	defer r.state.mu.Unlock()
	r.state.frozen = true
}

// findChild finds a child node with the exact given path segment (static nodes only).
func (r *Router) findChild(parent *node, path string) *node {
	for _, child := range parent.children {
		if child.path == path {
			return child
		}
	}
	return nil
}

func (r *Router) findParamChild(parent *node) *node {
	if parent == nil || len(parent.children) == 0 {
		return nil
	}
	for _, child := range parent.children {
		if len(child.path) > 0 && child.path[0] == ':' {
			return child
		}
	}
	return nil
}

func (r *Router) findWildChild(parent *node) *node {
	if parent == nil || len(parent.children) == 0 {
		return nil
	}
	for _, child := range parent.children {
		if len(child.path) > 0 && child.path[0] == '*' {
			return child
		}
	}
	return nil
}

// insertChild inserts a child node keeping indices sorted by first byte.
func (r *Router) insertChild(parent *node, child *node) {
	i := 0
	for i < len(parent.indices) && parent.indices[i] <= child.path[0] {
		i++
	}

	parent.indices = parent.indices[:i] + string(child.path[0]) + parent.indices[i:]

	parent.children = append(parent.children, nil)
	copy(parent.children[i+1:], parent.children[i:])
	parent.children[i] = child
}

func (r *Router) fullPath(path string) string {
	fullPath := joinRoutePath(r.prefix, path)
	if fullPath == "" {
		return "/"
	}
	return fullPath
}

func joinRoutePath(prefix, path string) string {
	prefix = strings.TrimRight(prefix, "/")
	if path != "" && path[0] != '/' {
		path = "/" + path
	}
	path = strings.TrimRight(path, "/")

	fullPath := prefix + path
	if fullPath == "" {
		return "/"
	}
	return fullPath
}

func (r *Router) metaFor(method, pattern string) RouteMeta {
	pattern = normalizeStoredPattern(pattern)
	if byMethod, ok := r.state.routeMeta[method]; ok {
		return byMethod[pattern]
	}
	return RouteMeta{}
}

func (r *Router) setMeta(method, pattern string, meta RouteMeta) {
	pattern = normalizeStoredPattern(pattern)
	if r.state.routeMeta == nil {
		r.state.routeMeta = make(map[string]map[string]RouteMeta)
	}
	if r.state.routeMeta[method] == nil {
		r.state.routeMeta[method] = make(map[string]RouteMeta)
	}
	r.state.routeMeta[method][pattern] = meta
}

func normalizeStoredPattern(pattern string) string {
	pattern = strings.TrimRight(pattern, "/")
	if pattern == "" {
		return "/"
	}
	return pattern
}

// Param returns the value of the named path parameter from the request context.
// Returns an empty string if the parameter is not found.
//
// Example:
//
//	r.AddRoute(http.MethodGet, "/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	    id := router.Param(r, "id")
//	    fmt.Fprintf(w, "User: %s", id)
//	}))
func Param(r *http.Request, name string) string {
	rc := contract.RequestContextFromContext(r.Context())
	return rc.Params[name]
}
