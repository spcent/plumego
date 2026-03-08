package router

import (
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	"github.com/spcent/plumego/middleware"
)

// HTTP method constants define standard HTTP methods for route registration.
// These constants are used when registering routes to specify which HTTP method
// the route should respond to.
//
// Example:
//
//	router.Get("/users", handler)        // GET /users
//	router.Post("/users", handler)       // POST /users
//	router.Put("/users/:id", handler)    // PUT /users/:id
//	router.Delete("/users/:id", handler) // DELETE /users/:id
//	router.Any("/health", handler)       // All HTTP methods
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
// This type is used for all route handler registrations in the router.
//
// Example:
//
//	import "net/http"
//
//	func myHandler(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Hello, World!"))
//	}
//
//	router.Get("/hello", http.HandlerFunc(myHandler))
//
// Use HandlerFunc when registering inline functions for better readability.
type Handler = http.Handler

// HandlerFunc is an alias to the standard http.HandlerFunc for convenience.
// This is the most common way to register simple route handlers.
//
// Example:
//
//	router.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Hello, World!"))
//	})
type HandlerFunc = http.HandlerFunc

// RouteRegistrar defines an interface for objects that can register routes with a router.
// This interface enables modular route registration and is commonly used for
// organizing routes by feature or domain.
//
// Example:
//
//	type UserRoutes struct{}
//
//	func (ur *UserRoutes) Register(r *router.Router) {
//	    r.Get("/users", ur.ListUsers)
//	    r.Post("/users", ur.CreateUser)
//	    r.Get("/users/:id", ur.GetUser)
//	}
//
//	func main() {
//	    r := router.NewRouter()
//	    r.Register(&UserRoutes{})
//	}
type RouteRegistrar interface {
	Register(r *Router)
}

// MiddlewareManager manages middleware chain for routes and groups.
// It provides thread-safe middleware registration and retrieval operations.
// Middleware is applied in the order they are added, creating a processing chain
// that each request passes through before reaching the final handler.
//
// Example:
//
//	manager := NewMiddlewareManager()
//	manager.AddMiddleware(middleware.Logging())
//	manager.AddMiddleware(middleware.Recovery())
//
//	// Middleware will be executed in order: Logging -> Recovery -> Handler
type MiddlewareManager struct {
	middlewares []middleware.Middleware
	mu          sync.RWMutex
	metrics     metrics.MetricsCollector // Unified metrics collector for monitoring
	version     atomic.Uint64
}

// NewMiddlewareManager creates a new middleware manager with an empty middleware chain.
// The manager is ready to accept middleware registrations immediately after creation.
//
// Example:
//
//	manager := NewMiddlewareManager()
//	manager.AddMiddleware(middleware.Logging())
func NewMiddlewareManager() *MiddlewareManager {
	return &MiddlewareManager{
		middlewares: make([]middleware.Middleware, 0),
	}
}

// AddMiddleware adds a middleware to the manager.
// The middleware will be executed in the order it was added.
// Thread-safe for concurrent access.
//
// Parameters:
//   - m: The middleware function to add
//
// Example:
//
//	manager := NewMiddlewareManager()
//	manager.AddMiddleware(middleware.Logging())
//	manager.AddMiddleware(middleware.Recovery())
func (mm *MiddlewareManager) AddMiddleware(m middleware.Middleware) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.middlewares = append(mm.middlewares, m)
	mm.version.Add(1)
}

// GetMiddlewares returns a copy of all middlewares in the manager.
// The returned slice is a copy, so modifications to it won't affect the manager.
// Thread-safe for concurrent access.
//
// Returns:
//   - []middleware.Middleware: A copy of all registered middlewares
//
// Example:
//
//	middlewares := manager.GetMiddlewares()
//	fmt.Printf("Registered middlewares: %d\n", len(middlewares))
func (mm *MiddlewareManager) GetMiddlewares() []middleware.Middleware {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	result := make([]middleware.Middleware, len(mm.middlewares))
	copy(result, mm.middlewares)
	return result
}

// Version returns the current middleware version for cache invalidation.
// This is a lock-free atomic read.
func (mm *MiddlewareManager) Version() uint64 {
	return mm.version.Load()
}

// MergeMiddlewares merges another middleware manager's middlewares with this manager's middlewares.
// The combined middleware chain will execute this manager's middlewares first,
// followed by the other manager's middlewares.
//
// Parameters:
//   - other: The other MiddlewareManager to merge with
//
// Returns:
//   - []middleware.Middleware: Combined middleware chain
//
// Example:
//
//	parentManager := NewMiddlewareManager()
//	parentManager.AddMiddleware(middleware.Logging())
//
//	childManager := NewMiddlewareManager()
//	childManager.AddMiddleware(middleware.Recovery())
//
//	combined := parentManager.MergeMiddlewares(childManager)
//	// combined will be: [Logging, Recovery]
func (mm *MiddlewareManager) MergeMiddlewares(other *MiddlewareManager) []middleware.Middleware {
	combined := mm.GetMiddlewares()
	otherMiddlewares := other.GetMiddlewares()
	return append(combined, otherMiddlewares...)
}

// segment represents a path segment with type information
type segment struct {
	raw       string
	isParam   bool
	isWild    bool
	paramName string
}

// node represents a node in the prefix trie
type node struct {
	path        string                  // path segment
	paramName   string                  // param or wildcard name for this segment
	fullPath    string                  // full path for this node (only set for nodes with handlers)
	indices     string                  // string of child path starts (optimized for lookup)
	children    []*node                 // child nodes
	handler     Handler                 // handler for this node
	paramKeys   []string                // parameter keys for this node
	middlewares []middleware.Middleware // middlewares specific to the route
	validation  *RouteValidation        // route parameter validation
}

type route struct {
	Method string
	Path   string
}

// routerState is the shared mutable state for the root router and all groups.
// Group routers carry only local prefix and middleware layering; all route trees,
// named routes, metadata, validations, and runtime caches live here.
type routerState struct {
	trees            map[string]*node
	registrars       []RouteRegistrar
	routes           map[string][]route
	frozen           bool
	mu               sync.RWMutex
	logger           log.StructuredLogger
	routeCache       *RouteCache
	routeValidations map[string]map[string]*RouteValidation
	routeMeta        map[string]map[string]RouteMeta
	namedRoutes      map[string]*NamedRoute
	// methodNotAllowed and metricsCollector use atomic access so the hot-path
	// ServeHTTP can read them without acquiring the global mutex.
	methodNotAllowed atomic.Bool
	metricsCollector atomic.Pointer[metrics.MetricsCollector]
}

// Router represents an HTTP router with path-based routing and middleware support.
// It implements the http.Handler interface, making it compatible with any Go HTTP server.
//
// Features:
//   - Path-based routing with support for parameters and wildcards
//   - Middleware chain support for request processing
//   - Route grouping with shared prefixes and middleware
//   - Route caching for improved performance
//   - Route metadata for documentation and API generation
//   - Named routes with reverse URL generation
//   - Thread-safe concurrent access
//
// Example:
//
//	r := router.NewRouter()
//	r.Get("/users", listUsersHandler)
//	r.Post("/users", createUserHandler)
//
//	// Group routes with shared prefix
//	api := r.Group("/api/v1")
//	api.Get("/users", listUsersHandler)
//	api.Get("/users/:id", getUserHandler)
//
//	// Named routes for reverse URL generation
//	r.Get("/users/:id", getUserHandler, router.WithRouteName("user.show"))
//	url := r.URL("user.show", "id", "123") // returns "/users/123"
//
//	// Start HTTP server
//	http.ListenAndServe(":8080", r)
type Router struct {
	prefix            string
	parent            *Router
	middlewareManager *MiddlewareManager
	state             *routerState
}

// RouterOption defines a function type for router configuration options.
// This functional option pattern allows flexible router configuration without
// requiring a complex configuration struct.
//
// Example:
//
//	r := router.NewRouter(
//	    router.WithLogger(customLogger),
//	    router.WithMetricsCollector(metricsCollector),
//	)
type RouterOption func(*Router)

// RouteOption defines an option for route metadata.
// This functional option pattern allows attaching metadata to routes for
// documentation, API generation, and other purposes.
//
// Example:
//
//	r.Get("/users", handler,
//	    router.WithRouteName("list_users"),
//	    router.WithRouteTags("users", "api"),
//	)
type RouteOption func(*RouteMeta)

// WithRouteName sets a route name for documentation and API generation.
// Route names should be unique and descriptive.
//
// Parameters:
//   - name: The name to assign to the route
//
// Example:
//
//	r.Get("/users", handler, router.WithRouteName("list_users"))
func WithRouteName(name string) RouteOption {
	return func(meta *RouteMeta) {
		meta.Name = name
	}
}

// WithRouteTags sets route tags for categorization and filtering.
// Tags are useful for organizing routes by feature, domain, or API version.
//
// Parameters:
//   - tags: Variable number of tag strings
//
// Example:
//
//	r.Get("/users", handler,
//	    router.WithRouteTags("users", "api", "v1"),
//	)
func WithRouteTags(tags ...string) RouteOption {
	return func(meta *RouteMeta) {
		meta.Tags = append([]string(nil), tags...)
	}
}

// WithRouteDescription sets a description for the route.
// This is useful for API documentation generation.
//
// Parameters:
//   - desc: Description text
//
// Example:
//
//	r.Get("/users", handler,
//	    router.WithRouteDescription("List all users with pagination"),
//	)
func WithRouteDescription(desc string) RouteOption {
	return func(meta *RouteMeta) {
		meta.Description = desc
	}
}

// WithRouteDeprecated marks the route as deprecated.
// This is useful for API versioning and documentation.
//
// Example:
//
//	r.Get("/v1/users", handler,
//	    router.WithRouteDeprecated(true),
//	)
func WithRouteDeprecated(deprecated bool) RouteOption {
	return func(meta *RouteMeta) {
		meta.Deprecated = deprecated
	}
}

// WithLogger sets a custom logger for the router.
// The logger is used for request logging and error reporting.
//
// Parameters:
//   - logger: A structured logger implementation
//
// Example:
//
//	import "github.com/spcent/plumego/log"
//
//	customLogger := log.NewGLogger()
//	r := router.NewRouter(router.WithLogger(customLogger))
func WithLogger(logger log.StructuredLogger) RouterOption {
	return func(r *Router) {
		if logger != nil {
			r.state.logger = logger
		}
	}
}

// WithMethodNotAllowed enables returning 405 with Allow header when path matches another method.
func WithMethodNotAllowed(enabled bool) RouterOption {
	return func(r *Router) {
		r.state.methodNotAllowed.Store(enabled)
	}
}

// NewRouter creates a new Router instance with default configuration.
// The router is ready to accept route registrations immediately after creation.
// Route caching is enabled by default with DefaultCacheCapacity.
//
// Parameters:
//   - opts: Optional router configuration options
//
// Returns:
//   - *Router: A new router instance
//
// Example:
//
//	r := router.NewRouter()
//	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Hello, World!"))
//	})
//
//	// With options
//	r := router.NewRouter(
//	    router.WithLogger(customLogger),
//	)
func NewRouter(opts ...RouterOption) *Router {
	r := &Router{
		prefix:            "",
		parent:            nil,
		middlewareManager: NewMiddlewareManager(),
		state: &routerState{
			trees:            make(map[string]*node),
			routes:           make(map[string][]route),
			logger:           log.NewGLogger(),
			routeValidations: make(map[string]map[string]*RouteValidation),
			routeMeta:        make(map[string]map[string]RouteMeta),
			namedRoutes:      make(map[string]*NamedRoute),
			routeCache:       NewRouteCache(DefaultCacheCapacity),
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(r)
	}

	return r
}

// SetLogger configures the logger used by context-aware handlers.
// This method is thread-safe and can be called at any time.
//
// Parameters:
//   - logger: The structured logger to use
//
// Example:
//
//	r := router.NewRouter()
//	r.SetLogger(customLogger)
func (r *Router) SetLogger(logger log.StructuredLogger) {
	r.state.mu.Lock()
	defer r.state.mu.Unlock()

	r.state.logger = logger
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
// This is useful for ensuring route stability after initialization and
// for detecting accidental route registrations after startup.
//
// Example:
//
//	r := router.NewRouter()
//	r.Get("/users", handler)
//	r.Freeze() // No more routes can be added
func (r *Router) Freeze() {
	r.state.mu.Lock()
	defer r.state.mu.Unlock()
	r.state.frozen = true
}

// Register adds route registrars to the router.
// This enables modular route registration where different components
// can register their own routes.
//
// Parameters:
//   - registrars: One or more RouteRegistrar implementations
//
// Example:
//
//	type UserRoutes struct{}
//
//	func (ur *UserRoutes) Register(r *router.Router) {
//	    r.Get("/users", ur.ListUsers)
//	    r.Post("/users", ur.CreateUser)
//	}
//
//	r := router.NewRouter()
//	r.Register(&UserRoutes{})
func (r *Router) Register(registrars ...RouteRegistrar) {
	// Deduplicate within the caller's input list first (no lock needed).
	var unique []RouteRegistrar
	localSeen := make(map[RouteRegistrar]bool, len(registrars))
	for _, reg := range registrars {
		if !localSeen[reg] {
			localSeen[reg] = true
			unique = append(unique, reg)
		}
	}

	for _, registrar := range unique {
		// Atomically check-and-append to prevent duplicate registration even
		// when Register is called concurrently. We must NOT hold the lock while
		// calling registrar.Register(r) because it invokes AddRoute, which also
		// acquires the lock.
		r.state.mu.Lock()
		alreadyRegistered := false
		for _, existing := range r.state.registrars {
			if existing == registrar {
				alreadyRegistered = true
				break
			}
		}
		if !alreadyRegistered {
			r.state.registrars = append(r.state.registrars, registrar)
		}
		r.state.mu.Unlock()

		if !alreadyRegistered {
			registrar.Register(r)
		}
	}
}

func (r *Router) routeMiddlewares() []middleware.Middleware {
	if r.parent == nil {
		return nil
	}

	var layers [][]middleware.Middleware
	for current := r; current != nil && current.parent != nil; current = current.parent {
		mws := current.middlewareManager.GetMiddlewares()
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

// Use adds middlewares to the router group.
// Middlewares are applied in the order they are added and will be executed
// for all routes registered in this group and its children.
//
// Parameters:
//   - middlewares: One or more middleware functions
//
// Example:
//
//	r := router.NewRouter()
//	r.Use(middleware.Logging())
//	r.Use(middleware.Recovery())
//
//	// All routes in this group will have logging and recovery middleware
//	api := r.Group("/api")
//	api.Get("/users", handler) // Logging -> Recovery -> Handler
func (r *Router) Use(middlewares ...middleware.Middleware) {
	for _, middleware := range middlewares {
		r.middlewareManager.AddMiddleware(middleware)
	}
}

// SetMetricsCollector sets the unified metrics collector for the router.
// The metrics collector is used to track route performance, request counts,
// and other metrics for monitoring and observability.
//
// Parameters:
//   - collector: Metrics collector implementation
//
// Example:
//
//	import "github.com/spcent/plumego/metrics"
//
//	prometheusCollector := metrics.NewPrometheusCollector()
//	r.SetMetricsCollector(prometheusCollector)
func (r *Router) SetMetricsCollector(collector metrics.MetricsCollector) {
	if collector == nil {
		r.state.metricsCollector.Store(nil)
	} else {
		r.state.metricsCollector.Store(&collector)
	}
}

// GetMetricsCollector returns the current metrics collector.
// Returns nil if no metrics collector has been set.
//
// Returns:
//   - metrics.MetricsCollector: The current metrics collector or nil
//
// Example:
//
//	collector := r.GetMetricsCollector()
//	if collector != nil {
//	    collector.IncrementCounter("route_requests", map[string]string{"route": "/users"})
//	}
func (r *Router) GetMetricsCollector() metrics.MetricsCollector {
	p := r.state.metricsCollector.Load()
	if p == nil {
		return nil
	}
	return *p
}

// CacheStats returns statistics about the route cache.
// This is useful for monitoring cache effectiveness and tuning capacity.
//
// Returns:
//   - CacheStats: Current cache statistics
//
// Example:
//
//	stats := r.CacheStats()
//	fmt.Printf("Cache hit rate: %.2f%%\n", stats.HitRate*100)
//	fmt.Printf("Exact entries: %d, Pattern entries: %d\n",
//	    stats.ExactEntries, stats.PatternEntries)
func (r *Router) CacheStats() CacheStats {
	if r.state.routeCache == nil {
		return CacheStats{}
	}
	return r.state.routeCache.Stats()
}

// ClearCache clears all cached route matching results.
// This is useful when routes are modified after initial registration
// or for testing purposes.
//
// Example:
//
//	r.ClearCache()
func (r *Router) ClearCache() {
	if r.state.routeCache != nil {
		r.state.routeCache.Clear()
	}
}

// CacheSize returns the current number of cached entries.
//
// Returns:
//   - int: Total number of cached entries (exact + pattern)
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

// insertChild inserts a child node into the parent's children list,
// keeping the indices string and children slice sorted by first byte.
func (r *Router) insertChild(parent *node, child *node) {
	// Find insertion point to keep indices sorted.
	i := 0
	for i < len(parent.indices) && parent.indices[i] <= child.path[0] {
		i++
	}

	// Insert the index character without allocating a temp string.
	parent.indices = parent.indices[:i] + string(child.path[0]) + parent.indices[i:]

	// Insert the child node using copy to avoid the temporary slice created by
	// the double-append pattern: append(a[:i], append([]T{x}, a[i:]...)...).
	parent.children = append(parent.children, nil)        // grow by one
	copy(parent.children[i+1:], parent.children[i:])      // shift right
	parent.children[i] = child
}

func (r *Router) addCtxRoute(method, path string, h contract.CtxHandlerFunc) error {
	if err := contract.ValidateCtxHandler(h); err != nil {
		return contract.WrapError(err, "add_ctx_route", "router", map[string]any{
			"method": method,
			"path":   path,
		})
	}

	r.state.mu.RLock()
	logger := r.state.logger
	r.state.mu.RUnlock()
	return r.AddRoute(method, path, contract.AdaptCtxHandler(h, logger))
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
