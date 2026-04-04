package router

import (
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
)

// HTTP method constants define standard HTTP methods for route registration.
const (
	GET     = "GET"     // HTTP GET method - retrieve resources
	POST    = "POST"    // HTTP POST method - create resources
	PUT     = "PUT"     // HTTP PUT method - update/replace resources
	DELETE  = "DELETE"  // HTTP DELETE method - remove resources
	PATCH   = "PATCH"   // HTTP PATCH method - partial updates
	OPTIONS = "OPTIONS" // HTTP OPTIONS method - describe communication options
	HEAD    = "HEAD"    // HTTP HEAD method - same as GET but without response body
	ANY     = "ANY"     // Any HTTP method - catch-all route
)

// Configuration defaults
const (
	DefaultCacheCapacity = 100 // Default route cache capacity
	DefaultMaxParams     = 8   // Default maximum parameters per route
	DefaultPoolSliceCap  = 4   // Default capacity for pooled slices
	DefaultPathPartsCap  = 8   // Default capacity for path parts slices
)

// Handler is an alias to the standard http.Handler for route handlers.
type Handler = http.Handler

// HandlerFunc is an alias to the standard http.HandlerFunc for convenience.
type HandlerFunc = http.HandlerFunc

// middlewareManager manages the middleware chain for a router or group.
// It is an internal type; callers interact with it only through Router.Use().
type middlewareManager struct {
	middlewares []middleware.Middleware
	mu          sync.RWMutex
	version     atomic.Uint64
}

func newMiddlewareManager() *middlewareManager {
	return &middlewareManager{
		middlewares: make([]middleware.Middleware, 0),
	}
}

func (mm *middlewareManager) addMiddleware(m middleware.Middleware) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.middlewares = append(mm.middlewares, m)
	mm.version.Add(1)
}

func (mm *middlewareManager) getMiddlewares() []middleware.Middleware {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	result := make([]middleware.Middleware, len(mm.middlewares))
	copy(result, mm.middlewares)
	return result
}

func (mm *middlewareManager) getVersion() uint64 {
	return mm.version.Load()
}

// segment represents a path segment with type information.
type segment struct {
	raw       string
	isParam   bool
	isWild    bool
	paramName string
}

// node represents a node in the prefix trie.
type node struct {
	path        string
	paramName   string
	fullPath    string
	indices     string
	children    []*node
	handler     Handler
	paramKeys   []string
	middlewares []middleware.Middleware
	validation  *RouteValidation
}

type route struct {
	Method string
	Path   string
}

// routerState is the shared mutable state for the root router and all groups.
// It does not carry application-layer concerns (logger, metrics); those live
// on the Router struct itself and are not shared across groups.
type routerState struct {
	trees            map[string]*node
	routes           map[string][]route
	frozen           bool
	mu               sync.RWMutex
	routeCache       *RouteCache
	routeValidations map[string]map[string]*RouteValidation
	routeMeta        map[string]map[string]RouteMeta
	namedRoutes      map[string]*NamedRoute
	methodNotAllowed atomic.Bool
}

// Router represents an HTTP router with path-based routing and middleware support.
// It implements http.Handler, making it compatible with any Go HTTP server.
//
// Example:
//
//	r := router.NewRouter()
//	r.Get("/users", listUsersHandler)
//	r.Post("/users", createUserHandler)
//
//	api := r.Group("/api/v1")
//	api.Get("/users/:id", getUserHandler)
//
//	http.ListenAndServe(":8080", r)
type Router struct {
	prefix            string
	parent            *Router
	middlewareManager *middlewareManager
	state             *routerState

	// logger is stored directly on the Router (not in shared routerState) so
	// that groups can carry the same logger reference without coupling the
	// routing state to application-layer concerns.
	logger log.StructuredLogger
}

// RouterOption defines a function type for router configuration options.
type RouterOption func(*Router)

// RouteOption defines an option for route metadata.
type RouteOption func(*RouteMeta)

// WithRouteName sets a route name for reverse URL generation.
//
// Example:
//
//	r.AddRouteWithOptions("GET", "/users/:id", handler, router.WithRouteName("user.show"))
//	url := r.URL("user.show", "id", "123") // → "/users/123"
func WithRouteName(name string) RouteOption {
	return func(meta *RouteMeta) {
		meta.Name = name
	}
}

// WithLogger sets a logger on the router. The logger is not used by routing
// logic itself; it is available via Logger() so that components registering
// routes can obtain the application logger without an extra dependency.
func WithLogger(logger log.StructuredLogger) RouterOption {
	return func(r *Router) {
		r.logger = logger
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
//	r.Get("/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Hello, World!"))
//	}))
func NewRouter(opts ...RouterOption) *Router {
	r := &Router{
		prefix:            "",
		parent:            nil,
		middlewareManager: newMiddlewareManager(),
		state: &routerState{
			trees:            make(map[string]*node),
			routes:           make(map[string][]route),
			routeValidations: make(map[string]map[string]*RouteValidation),
			routeMeta:        make(map[string]map[string]RouteMeta),
			namedRoutes:      make(map[string]*NamedRoute),
			routeCache:       NewRouteCache(DefaultCacheCapacity),
		},
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// SetLogger sets the logger returned by Logger(). The router does not use the
// logger internally; this is a convenience for component code that receives a
// *Router and needs to obtain the application logger.
func (r *Router) SetLogger(logger log.StructuredLogger) {
	r.logger = logger
}

// Logger returns the logger associated with this router (or nil if unset).
// Components that need to pass a logger to contract.AdaptCtxHandler can use
// this to obtain the one configured via WithLogger or SetLogger.
func (r *Router) Logger() log.StructuredLogger {
	return r.logger
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

// Use adds middlewares to the router or group.
// Middlewares execute in the order they are added, for all routes in this group and its children.
//
// Example:
//
//	r := router.NewRouter()
//	r.Use(middleware.Logging())
//	r.Use(middleware.Recovery())
func (r *Router) Use(middlewares ...middleware.Middleware) {
	for _, m := range middlewares {
		r.middlewareManager.addMiddleware(m)
	}
}

// CacheStats returns statistics about the route cache.
func (r *Router) CacheStats() CacheStats {
	if r.state.routeCache == nil {
		return CacheStats{}
	}
	return r.state.routeCache.Stats()
}

// ClearCache clears all cached route matching results.
func (r *Router) ClearCache() {
	if r.state.routeCache != nil {
		r.state.routeCache.Clear()
	}
}

// CacheSize returns the current number of cached entries.
func (r *Router) CacheSize() int {
	if r.state.routeCache == nil {
		return 0
	}
	return r.state.routeCache.Size()
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

func (r *Router) routeMiddlewares() []middleware.Middleware {
	if r.parent == nil {
		return nil
	}

	var layers [][]middleware.Middleware
	for current := r; current != nil && current.parent != nil; current = current.parent {
		mws := current.middlewareManager.getMiddlewares()
		if len(mws) > 0 {
			layers = append(layers, mws)
		}
	}

	if len(layers) == 0 {
		return nil
	}

	total := 0
	for i := len(layers) - 1; i >= 0; i-- {
		total += len(layers[i])
	}

	combined := make([]middleware.Middleware, 0, total)
	for i := len(layers) - 1; i >= 0; i-- {
		combined = append(combined, layers[i]...)
	}

	return combined
}

func (r *Router) fullPath(path string) string {
	fullPath := r.prefix + strings.TrimRight(path, "/")
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

func (r *Router) validationFor(method, pattern string) *RouteValidation {
	pattern = normalizeStoredPattern(pattern)
	if byMethod, ok := r.state.routeValidations[method]; ok {
		return byMethod[pattern]
	}
	return nil
}

func (r *Router) setValidation(method, pattern string, validation *RouteValidation) {
	pattern = normalizeStoredPattern(pattern)
	if r.state.routeValidations == nil {
		r.state.routeValidations = make(map[string]map[string]*RouteValidation)
	}
	if r.state.routeValidations[method] == nil {
		r.state.routeValidations[method] = make(map[string]*RouteValidation)
	}
	r.state.routeValidations[method][pattern] = validation
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
//	r.Get("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	    id := router.Param(r, "id")
//	    fmt.Fprintf(w, "User: %s", id)
//	}))
func Param(r *http.Request, name string) string {
	rc := contract.RequestContextFrom(r.Context())
	return rc.Params[name]
}
