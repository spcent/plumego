package router

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/spcent/plumego/contract"
	log "github.com/spcent/plumego/log"
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
	GET    = "GET"    // HTTP GET method - retrieve resources
	POST   = "POST"   // HTTP POST method - create resources
	PUT    = "PUT"    // HTTP PUT method - update/replace resources
	DELETE = "DELETE" // HTTP DELETE method - remove resources
	PATCH  = "PATCH"  // HTTP PATCH method - partial updates
	ANY    = "ANY"    // Any HTTP method - catch-all route
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
	priority    int                     // priority for this node (higher = more specific)
	middlewares []middleware.Middleware // middlewares specific to the route
}

type route struct {
	Method string
	Path   string
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
//	// Start HTTP server
//	http.ListenAndServe(":8080", r)
type Router struct {
	prefix            string                      // Group prefix for route grouping
	trees             map[string]*node            // Method -> root node (Radix tree)
	registrars        []RouteRegistrar            // Route registrars for modular registration
	routes            map[string][]route          // Registered routes for debugging
	frozen            bool                        // Whether router is frozen (no new routes)
	mu                sync.RWMutex                // Mutex for concurrent access
	parent            *Router                     // Parent router for groups
	middlewareManager *MiddlewareManager          // Middleware management
	logger            log.StructuredLogger        // Logger for contextual handlers
	routeCache        *RouteCache                 // Route matching cache for performance
	routeValidations  map[string]*RouteValidation // Route parameter validations
	validationIndex   map[string][]validationEntry
	routeMeta         map[string]RouteMeta
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
			r.logger = logger
		}
	}
}

// NewRouter creates a new Router instance with default configuration.
// The router is ready to accept route registrations immediately after creation.
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
		trees:             make(map[string]*node),
		routes:            make(map[string][]route),
		prefix:            "",
		parent:            nil,
		middlewareManager: NewMiddlewareManager(),
		logger:            log.NewGLogger(),
		routeValidations:  make(map[string]*RouteValidation),
		validationIndex:   make(map[string][]validationEntry),
		routeMeta:         make(map[string]RouteMeta),
		routeCache:        NewRouteCache(100), // Initialize with default cache capacity
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
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logger = logger
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
	r.mu.Lock()
	defer r.mu.Unlock()
	r.frozen = true
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
	// Get unique registrars without holding the lock
	// This prevents deadlock when reg.Register(r) calls back into AddRoute
	var newRegistrars []RouteRegistrar

	r.mu.RLock()
	seen := make(map[RouteRegistrar]bool)
	for _, registrar := range r.registrars {
		seen[registrar] = true
	}
	r.mu.RUnlock()

	// Find new registrars
	for _, registrar := range registrars {
		if !seen[registrar] {
			newRegistrars = append(newRegistrars, registrar)
			seen[registrar] = true
		}
	}

	// Register new registrars
	for _, registrar := range newRegistrars {
		// Add to registrars list
		r.mu.Lock()
		r.registrars = append(r.registrars, registrar)
		r.mu.Unlock()

		// Call Register without holding the lock
		// This allows registrar.Register(r) to call AddRoute safely
		registrar.Register(r)
	}
}

// Group creates a new router group with the given prefix.
// Groups allow you to share a common path prefix and middleware across multiple routes.
// Child groups inherit the parent's prefix and can add their own middleware.
//
// Parameters:
//   - prefix: The path prefix for the group (e.g., "/api/v1")
//
// Returns:
//   - *Router: A new router group
//
// Example:
//
//	r := router.NewRouter()
//
//	// Create API group with shared prefix
//	api := r.Group("/api/v1")
//	api.Get("/users", listUsersHandler)
//	api.Get("/users/:id", getUserHandler)
//
//	// Create admin group with additional middleware
//	admin := api.Group("/admin")
//	admin.Use(middleware.AuthRequired())
//	admin.Get("/dashboard", dashboardHandler)
func (r *Router) Group(prefix string) *Router {
	// Create full prefix by combining with parent's prefix
	fullPrefix := r.prefix + prefix

	return &Router{
		prefix:            fullPrefix,
		trees:             r.trees,
		routes:            r.routes,
		parent:            r,
		frozen:            r.frozen,
		middlewareManager: NewMiddlewareManager(),
		logger:            r.logger,
		routeValidations:  r.routeValidations,
		validationIndex:   r.validationIndex,
		routeMeta:         r.routeMeta,
	}
}

// AddRoute adds a route to the router with the given method, path and handler.
// This is the core method for route registration. Routes are stored in a Radix tree
// for efficient matching. Path parameters are denoted with ":" (e.g., "/users/:id")
// and wildcards with "*" (e.g., "/files/*path").
//
// Parameters:
//   - method: HTTP method (GET, POST, PUT, DELETE, PATCH, ANY)
//   - path: URL path (can include parameters)
//   - handler: HTTP handler for the route
//
// Returns:
//   - error: Error if route registration fails (duplicate, frozen, etc.)
//
// Example:
//
//	r := router.NewRouter()
//	r.AddRoute("GET", "/users", listUsersHandler)
//	r.AddRoute("GET", "/users/:id", getUserHandler)
//	r.AddRoute("GET", "/files/*path", fileHandler)
//
//	// Or use convenience methods
//	r.Get("/users", listUsersHandler)
//	r.Get("/users/:id", getUserHandler)
func (r *Router) AddRoute(method, path string, handler Handler) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.frozen {
		return contract.WrapError(
			fmt.Errorf("router is frozen, cannot add route after freeze"),
			"add_route",
			"router",
			map[string]any{
				"method": method,
				"path":   path,
			},
		)
	}

	// Combine with group prefix
	fullPath := r.fullPath(path)
	if fullPath == "" {
		fullPath = "/"
	}

	if r.trees[method] == nil {
		r.trees[method] = &node{}
	}

	current := r.trees[method]
	current.priority++

	// Start with root path
	if fullPath == "/" {
		if current.handler != nil {
			return contract.WrapError(
				fmt.Errorf("duplicate route registration: %s /", method),
				"add_route",
				"router",
				map[string]any{
					"method": method,
					"path":   fullPath,
				},
			)
		}
		current.handler = handler
		current.fullPath = fullPath
		r.routes[method] = append(r.routes[method], route{Method: method, Path: fullPath})
		return nil
	}

	segments := compilePathSegments(fullPath)
	paramKeys := make([]string, 0, len(segments))

	// Add each segment to the trie
	for _, seg := range segments {
		if seg.isParam {
			paramKeys = append(paramKeys, seg.paramName)
			child := r.findParamChild(current)
			if child != nil {
				if child.paramName == "" {
					child.paramName = seg.paramName
				} else if child.paramName != seg.paramName {
					return contract.WrapError(
						fmt.Errorf("route conflict: parameter name mismatch. Existing: %s, New: %s", child.paramName, seg.paramName),
						"add_route",
						"router",
						map[string]any{
							"method": method,
							"path":   fullPath,
						},
					)
				}
				current = child
				current.priority++
				continue
			}

			child = &node{
				path:      ":",
				paramName: seg.paramName,
			}
			r.insertChild(current, child)
			current = child
			current.priority++
			continue
		}

		if seg.isWild {
			paramKeys = append(paramKeys, seg.paramName)
			child := r.findWildChild(current)
			if child != nil {
				if child.paramName == "" {
					child.paramName = seg.paramName
				} else if child.paramName != seg.paramName {
					return contract.WrapError(
						fmt.Errorf("route conflict: wildcard parameter name mismatch. Existing: %s, New: %s", child.paramName, seg.paramName),
						"add_route",
						"router",
						map[string]any{
							"method": method,
							"path":   fullPath,
						},
					)
				}
				current = child
				current.priority++
				continue
			}

			child = &node{
				path:      "*",
				paramName: seg.paramName,
			}
			r.insertChild(current, child)
			current = child
			current.priority++
			continue
		}

		// Find or create static child node
		child := r.findChild(current, seg.raw)
		if child == nil {
			child = &node{
				path: seg.raw,
			}
			r.insertChild(current, child)
		}

		current = child
		current.priority++
	}

	// Set handler for the final node
	if current.handler != nil {
		return contract.WrapError(
			fmt.Errorf("duplicate route registration: %s %s", method, fullPath),
			"add_route",
			"router",
			map[string]any{
				"method": method,
				"path":   fullPath,
			},
		)
	}
	current.handler = handler
	current.paramKeys = paramKeys
	current.fullPath = fullPath
	current.middlewares = r.routeMiddlewares()

	r.routes[method] = append(r.routes[method], route{Method: method, Path: fullPath})
	return nil
}

// AddRouteWithOptions adds a route and attaches metadata options.
// This method combines route registration with metadata attachment in a single call.
//
// Parameters:
//   - method: HTTP method
//   - path: URL path
//   - handler: HTTP handler
//   - opts: Route metadata options
//
// Returns:
//   - error: Error if route registration fails
//
// Example:
//
//	r.AddRouteWithOptions(
//	    "GET",
//	    "/users",
//	    listUsersHandler,
//	    router.WithRouteName("list_users"),
//	    router.WithRouteTags("users", "api"),
//	)
func (r *Router) AddRouteWithOptions(method, path string, handler Handler, opts ...RouteOption) error {
	if err := r.AddRoute(method, path, handler); err != nil {
		return err
	}
	if len(opts) == 0 {
		return nil
	}
	meta := RouteMeta{}
	for _, opt := range opts {
		if opt != nil {
			opt(&meta)
		}
	}
	r.SetRouteMeta(method, path, meta)
	return nil
}

// SetRouteMeta sets metadata for a route.
// This is useful for attaching documentation, tags, or other metadata to routes
// for API generation, documentation, or monitoring purposes.
//
// Parameters:
//   - method: HTTP method
//   - path: URL path
//   - meta: Route metadata
//
// Example:
//
//	meta := RouteMeta{
//	    Name: "list_users",
//	    Tags: []string{"users", "api"},
//	}
//	r.SetRouteMeta("GET", "/users", meta)
func (r *Router) SetRouteMeta(method, path string, meta RouteMeta) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.routeMeta == nil {
		r.routeMeta = make(map[string]RouteMeta)
	}
	key := r.routeKey(method, path)
	r.routeMeta[key] = meta
}

// Routes returns a snapshot of all registered routes with metadata.
// This method is useful for debugging, documentation generation, and monitoring.
// The returned slice is sorted by method and path for consistent output.
//
// Returns:
//   - []RouteInfo: List of all registered routes with metadata
//
// Example:
//
//	routes := r.Routes()
//	for _, route := range routes {
//	    fmt.Printf("%s %s %v\n", route.Method, route.Path, route.Meta.Tags)
//	}
func (r *Router) Routes() []RouteInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]RouteInfo, 0)
	for _, routes := range r.routes {
		for _, entry := range routes {
			key := entry.Method + " " + entry.Path
			meta := r.routeMeta[key]
			infos = append(infos, RouteInfo{
				Method: entry.Method,
				Path:   entry.Path,
				Meta:   meta,
			})
		}
	}

	sort.Slice(infos, func(i, j int) bool {
		if infos[i].Method == infos[j].Method {
			return infos[i].Path < infos[j].Path
		}
		return infos[i].Method < infos[j].Method
	})

	return infos
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
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewareManager.metrics = collector
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
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.middlewareManager.metrics
}

// findChild finds a child node with the given path segment
func (r *Router) findChild(parent *node, path string) *node {
	// Check if it's a wildcard segment
	if len(path) == 1 {
		switch path {
		case ":":
			// Check for param child - return nil to indicate no existing param child found
			// This forces creation of a new param child node
			return nil
		case "*":
			// Check for wild child - return nil to indicate no existing wild child found
			// This forces creation of a new wild child node
			return nil
		default:
			// Check for exact match
			for _, child := range parent.children {
				if child.path == path {
					return child
				}
			}
		}
	} else {
		// Check for exact match for longer paths
		for _, child := range parent.children {
			if child.path == path {
				return child
			}
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

// insertChild inserts a child node into the parent's children list
func (r *Router) insertChild(parent *node, child *node) {
	// Find insertion point to keep indices sorted
	var i int
	for i = 0; i < len(parent.indices); i++ {
		if parent.indices[i] > child.path[0] {
			break
		}
	}

	// Insert index
	parent.indices = parent.indices[:i] + string(child.path[0]) + parent.indices[i:]

	// Insert child
	parent.children = append(parent.children[:i], append([]*node{child}, parent.children[i:]...)...)
}

func (r *Router) addCtxRoute(method, path string, h contract.CtxHandlerFunc) error {
	if err := contract.ValidateCtxHandler(h); err != nil {
		return contract.WrapError(err, "add_ctx_route", "router", map[string]any{
			"method": method,
			"path":   path,
		})
	}

	return r.AddRoute(method, path, contract.AdaptCtxHandler(h, r.logger))
}

func (r *Router) routeKey(method, path string) string {
	return method + " " + r.fullPath(path)
}

func (r *Router) fullPath(path string) string {
	fullPath := r.prefix + strings.TrimRight(path, "/")
	if fullPath == "" {
		return "/"
	}
	return fullPath
}

// ServeHTTP implements http.Handler and handles incoming HTTP requests.
// This method is the entry point for all HTTP requests and performs route matching,
// parameter extraction, validation, and middleware execution.
//
// Parameters:
//   - w: HTTP response writer
//   - req: HTTP request
//
// Example:
//
//	r := router.NewRouter()
//	r.Get("/users", listUsersHandler)
//
//	// Start HTTP server
//	http.ListenAndServe(":8080", r)
//
//	// Or use with other handlers
//	http.Handle("/", r)
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Check route cache first for better performance
	cacheKey := req.Method + ":" + req.Host + ":" + req.URL.Path
	if cachedResult, exists := r.routeCache.Get(cacheKey); exists {
		// Use cached result without holding the lock
		r.handleCachedRouteMatch(w, req, cachedResult)
		return
	}

	// Acquire read lock for route matching
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Find appropriate route tree for the request method
	tree, found := r.findRouteTree(req.Method)
	if !found {
		http.NotFound(w, req)
		return
	}

	// Parse and normalize request path
	path := r.normalizePath(req.URL.Path)

	// Try to match route and handle request
	r.handleRouteMatch(w, req, tree, path, cacheKey)
}

// findRouteTree finds the appropriate route tree for the given HTTP method
func (r *Router) findRouteTree(method string) (*node, bool) {
	tree := r.trees[method]
	if tree == nil {
		tree = r.trees[ANY]
		if tree == nil {
			return nil, false
		}
	}
	return tree, true
}

// normalizePath normalizes the request path by trimming trailing slashes
func (r *Router) normalizePath(path string) string {
	if path == "/" || path == "" {
		return "/"
	}
	return strings.Trim(path, "/")
}

// handleRouteMatch processes the matched route and applies middleware
func (r *Router) handleRouteMatch(w http.ResponseWriter, req *http.Request, tree *node, path string, cacheKey string) {
	// Handle root path specially
	if path == "/" {
		r.handleRootRequest(w, req, tree, cacheKey)
		return
	}

	// Perform route matching for non-root paths
	parts := strings.Split(path, "/")
	matcher := NewRouteMatcher(tree)
	result := matcher.Match(parts)

	// Try ANY method if specific method didn't match
	if result == nil && req.Method != ANY {
		if anyTree := r.trees[ANY]; anyTree != nil {
			anyMatcher := NewRouteMatcher(anyTree)
			result = anyMatcher.Match(parts)
		}
	}

	if result == nil {
		http.NotFound(w, req)
		return
	}

	// Build parameter map
	params := r.buildParamMap(result.ParamValues, result.ParamKeys)

	// Validate parameters if validations are registered
	// Use the full path with prefix for validation lookup
	fullPath := r.prefix + path
	if params != nil {
		if err := r.validateRouteParams(req.Method, fullPath, params); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Cache the matching result for future requests
	r.routeCache.Set(cacheKey, result)

	r.applyMiddlewareAndServe(w, req, params, result.Handler, result.RouteMiddlewares)
}

// handleCachedRouteMatch handles requests using cached route matching results
func (r *Router) handleCachedRouteMatch(w http.ResponseWriter, req *http.Request, result *MatchResult) {
	// Build parameter map
	params := r.buildParamMap(result.ParamValues, result.ParamKeys)

	// Validate parameters if validations are registered
	if params != nil {
		path := r.normalizePath(req.URL.Path)
		fullPath := r.prefix + path
		if err := r.validateRouteParams(req.Method, fullPath, params); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	r.applyMiddlewareAndServe(w, req, params, result.Handler, result.RouteMiddlewares)
}

// handleRootRequest handles requests to the root path "/"
func (r *Router) handleRootRequest(w http.ResponseWriter, req *http.Request, tree *node, cacheKey string) {
	if tree.handler != nil {
		// Create a MatchResult for the root path with direct middleware slice
		result := &MatchResult{
			Handler:          tree.handler,
			ParamValues:      nil,
			ParamKeys:        nil,
			RouteMiddlewares: tree.middlewares,
		}

		// Cache the result for future requests
		r.routeCache.Set(cacheKey, result)

		r.applyMiddlewareAndServe(w, req, nil, tree.handler, result.RouteMiddlewares)
		return
	}
	if req.Method != ANY {
		if anyTree := r.trees[ANY]; anyTree != nil && anyTree.handler != nil {
			result := &MatchResult{
				Handler:          anyTree.handler,
				ParamValues:      nil,
				ParamKeys:        nil,
				RouteMiddlewares: anyTree.middlewares,
			}
			r.routeCache.Set(cacheKey, result)
			r.applyMiddlewareAndServe(w, req, nil, anyTree.handler, result.RouteMiddlewares)
			return
		}
	}
	http.NotFound(w, req)
}

// buildParamMap creates a parameter map from values and keys
func (r *Router) buildParamMap(paramValues []string, paramKeys []string) map[string]string {
	if paramValues == nil || paramKeys == nil {
		return nil
	}
	if len(paramValues) == 0 || len(paramKeys) == 0 {
		return nil
	}

	params := make(map[string]string)
	minLen := len(paramValues)
	if len(paramKeys) < minLen {
		minLen = len(paramKeys)
	}

	for i := 0; i < minLen; i++ {
		params[paramKeys[i]] = paramValues[i]
	}

	return params
}

// applyMiddlewareAndServe applies middleware chain to the handler and serves the request
func (r *Router) applyMiddlewareAndServe(w http.ResponseWriter, req *http.Request, params map[string]string, handler http.Handler, routeMiddlewares []middleware.Middleware) {
	reqWithParams := req
	ctx := req.Context()

	if len(params) > 0 {
		ctx = context.WithValue(ctx, contract.ParamsContextKey{}, params)
	}

	// Always install a RequestContext so downstream code has a predictable place to read/write
	// request-scoped data without custom type assertions.
	existingRC, ok := ctx.Value(contract.RequestContextKey{}).(contract.RequestContext)
	if !ok {
		existingRC = contract.RequestContext{}
	}

	if len(params) > 0 {
		existingRC.Params = params
	}

	ctx = context.WithValue(ctx, contract.RequestContextKey{}, existingRC)
	if ctx != req.Context() {
		reqWithParams = req.WithContext(ctx)
	}

	// Combine middleware slices directly
	allMiddlewares := r.middlewareManager.GetMiddlewares()
	combined := make([]middleware.Middleware, 0, len(allMiddlewares)+len(routeMiddlewares))
	combined = append(combined, allMiddlewares...)
	combined = append(combined, routeMiddlewares...)

	if len(combined) == 0 {
		handler.ServeHTTP(w, reqWithParams)
		return
	}

	chain := middleware.NewChain(combined...)
	wrappedHandler := chain.Apply(handler)
	wrappedHandler.ServeHTTP(w, reqWithParams)
}

// compilePathSegments parses a URL path into route segments
func compilePathSegments(path string) []segment {
	if path == "/" {
		return nil
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	segments := make([]segment, 0, len(parts))
	for _, segmentPart := range parts {
		if strings.HasPrefix(segmentPart, ":") {
			segments = append(segments, segment{
				raw:       segmentPart,
				isParam:   true,
				paramName: segmentPart[1:],
			})
		} else if strings.HasPrefix(segmentPart, "*") {
			segments = append(segments, segment{
				raw:       segmentPart,
				isWild:    true,
				paramName: segmentPart[1:],
			})
		} else {
			segments = append(segments, segment{raw: segmentPart})
		}
	}
	return segments
}

// --- Helper API ---

// addRoute is a generic method for registering routes
func (r *Router) addRoute(method, path string, handler Handler) {
	r.AddRoute(method, path, handler)
}

// HTTP method-specific route registration
// These convenience methods provide a fluent API for route registration.

// Get registers a GET route with the given path and handler.
// GET requests are used to retrieve resources.
//
// Parameters:
//   - path: URL path (can include parameters like "/users/:id")
//   - handler: HTTP handler
//
// Example:
//
//	r.Get("/users", listUsersHandler)
//	r.Get("/users/:id", getUserHandler)
func (r *Router) Get(path string, handler Handler) { r.addRoute(GET, path, handler) }

// Post registers a POST route with the given path and handler.
// POST requests are used to create new resources.
//
// Parameters:
//   - path: URL path
//   - handler: HTTP handler
//
// Example:
//
//	r.Post("/users", createUserHandler)
func (r *Router) Post(path string, handler Handler) { r.addRoute(POST, path, handler) }

// Put registers a PUT route with the given path and handler.
// PUT requests are used to replace existing resources.
//
// Parameters:
//   - path: URL path (typically includes resource ID like "/users/:id")
//   - handler: HTTP handler
//
// Example:
//
//	r.Put("/users/:id", updateUserHandler)
func (r *Router) Put(path string, handler Handler) { r.addRoute(PUT, path, handler) }

// Delete registers a DELETE route with the given path and handler.
// DELETE requests are used to remove resources.
//
// Parameters:
//   - path: URL path (typically includes resource ID like "/users/:id")
//   - handler: HTTP handler
//
// Example:
//
//	r.Delete("/users/:id", deleteUserHandler)
func (r *Router) Delete(path string, handler Handler) { r.addRoute(DELETE, path, handler) }

// Patch registers a PATCH route with the given path and handler.
// PATCH requests are used for partial updates to resources.
//
// Parameters:
//   - path: URL path (typically includes resource ID like "/users/:id")
//   - handler: HTTP handler
//
// Example:
//
//	r.Patch("/users/:id", patchUserHandler)
func (r *Router) Patch(path string, handler Handler) { r.addRoute(PATCH, path, handler) }

// Any registers a route that accepts any HTTP method with the given path and handler.
// This is useful for catch-all routes or when you want to handle all methods the same way.
//
// Parameters:
//   - path: URL path
//   - handler: HTTP handler
//
// Example:
//
//	r.Any("/health", healthCheckHandler) // Handles GET, POST, PUT, etc.
func (r *Router) Any(path string, handler Handler) { r.addRoute(ANY, path, handler) }

// Options registers an OPTIONS route with the given path and handler.
// OPTIONS requests are used to describe communication options for the target resource.
//
// Parameters:
//   - path: URL path
//   - handler: HTTP handler
//
// Example:
//
//	r.Options("/users", optionsHandler)
func (r *Router) Options(path string, handler Handler) { r.addRoute("OPTIONS", path, handler) }

// Head registers a HEAD route with the given path and handler.
// HEAD requests are identical to GET requests but without the response body.
//
// Parameters:
//   - path: URL path
//   - handler: HTTP handler
//
// Example:
//
//	r.Head("/users", headHandler)
func (r *Router) Head(path string, handler Handler) { r.addRoute("HEAD", path, handler) }

// Context-aware handler registration helpers
// These methods register routes with context-aware handlers that receive
// a request context and can access route parameters and other request-scoped data.

// GetCtx registers a GET route with a context-aware handler.
// Context-aware handlers provide direct access to request context and parameters.
//
// Parameters:
//   - path: URL path
//   - handler: Context-aware handler function
//
// Example:
//
//	r.GetCtx("/users/:id", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
//	    params := router.ParamsFromContext(ctx)
//	    userID := params["id"]
//	    // Handle request...
//	    return nil
//	})
func (r *Router) GetCtx(path string, handler contract.CtxHandlerFunc) {
	r.addCtxRoute(GET, path, handler)
}

// PostCtx registers a POST route with a context-aware handler.
//
// Parameters:
//   - path: URL path
//   - handler: Context-aware handler function
//
// Example:
//
//	r.PostCtx("/users", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
//	    // Handle request...
//	    return nil
//	})
func (r *Router) PostCtx(path string, handler contract.CtxHandlerFunc) {
	r.addCtxRoute(POST, path, handler)
}

// PutCtx registers a PUT route with a context-aware handler.
//
// Parameters:
//   - path: URL path
//   - handler: Context-aware handler function
//
// Example:
//
//	r.PutCtx("/users/:id", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
//	    params := router.ParamsFromContext(ctx)
//	    userID := params["id"]
//	    // Handle request...
//	    return nil
//	})
func (r *Router) PutCtx(path string, handler contract.CtxHandlerFunc) {
	r.addCtxRoute(PUT, path, handler)
}

// DeleteCtx registers a DELETE route with a context-aware handler.
//
// Parameters:
//   - path: URL path
//   - handler: Context-aware handler function
//
// Example:
//
//	r.DeleteCtx("/users/:id", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
//	    params := router.ParamsFromContext(ctx)
//	    userID := params["id"]
//	    // Handle request...
//	    return nil
//	})
func (r *Router) DeleteCtx(path string, handler contract.CtxHandlerFunc) {
	r.addCtxRoute(DELETE, path, handler)
}

// PatchCtx registers a PATCH route with a context-aware handler.
//
// Parameters:
//   - path: URL path
//   - handler: Context-aware handler function
//
// Example:
//
//	r.PatchCtx("/users/:id", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
//	    params := router.ParamsFromContext(ctx)
//	    userID := params["id"]
//	    // Handle request...
//	    return nil
//	})
func (r *Router) PatchCtx(path string, handler contract.CtxHandlerFunc) {
	r.addCtxRoute(PATCH, path, handler)
}

// AnyCtx registers a route that accepts any HTTP method with a context-aware handler.
//
// Parameters:
//   - path: URL path
//   - handler: Context-aware handler function
//
// Example:
//
//	r.AnyCtx("/webhook", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
//	    // Handle webhook from any method...
//	    return nil
//	})
func (r *Router) AnyCtx(path string, handler contract.CtxHandlerFunc) {
	r.addCtxRoute(ANY, path, handler)
}

// HandleFunc registers a standard http.HandlerFunc for the given path and method.
// This is a convenience method that wraps AddRoute for http.HandlerFunc.
//
// Parameters:
//   - method: HTTP method
//   - path: URL path
//   - h: Standard http.HandlerFunc
//
// Example:
//
//	r.HandleFunc("GET", "/users", func(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Hello"))
//	})
func (r *Router) HandleFunc(method, path string, h http.HandlerFunc) {
	r.AddRoute(method, path, h)
}

// Handle registers a standard http.Handler for the given path and method.
// This is a convenience method that wraps AddRoute for http.Handler.
//
// Parameters:
//   - method: HTTP method
//   - path: URL path
//   - h: Standard http.Handler
//
// Example:
//
//	r.Handle("GET", "/users", http.HandlerFunc(myHandler))
func (r *Router) Handle(method, path string, h http.Handler) {
	r.AddRoute(method, path, h)
}

// HandleWithOptions registers a standard http.Handler for the given path and method with metadata.
// This method combines route registration with metadata attachment.
//
// Parameters:
//   - method: HTTP method
//   - path: URL path
//   - h: Standard http.Handler
//   - opts: Route metadata options
//
// Example:
//
//	r.HandleWithOptions(
//	    "GET",
//	    "/users",
//	    http.HandlerFunc(myHandler),
//	    router.WithRouteName("list_users"),
//	)
func (r *Router) HandleWithOptions(method, path string, h http.Handler, opts ...RouteOption) {
	_ = r.AddRouteWithOptions(method, path, h, opts...)
}

// GetFunc registers a GET route with a standard http.HandlerFunc.
// This is a convenience method combining Get and HandleFunc.
//
// Parameters:
//   - path: URL path
//   - h: Standard http.HandlerFunc
//
// Example:
//
//	r.GetFunc("/users", func(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Hello"))
//	})
func (r *Router) GetFunc(path string, h http.HandlerFunc) {
	r.HandleFunc(GET, path, h)
}

// PostFunc registers a POST route with a standard http.HandlerFunc.
//
// Parameters:
//   - path: URL path
//   - h: Standard http.HandlerFunc
//
// Example:
//
//	r.PostFunc("/users", func(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Created"))
//	})
func (r *Router) PostFunc(path string, h http.HandlerFunc) {
	r.HandleFunc(POST, path, h)
}

// PutFunc registers a PUT route with a standard http.HandlerFunc.
//
// Parameters:
//   - path: URL path
//   - h: Standard http.HandlerFunc
//
// Example:
//
//	r.PutFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Updated"))
//	})
func (r *Router) PutFunc(path string, h http.HandlerFunc) {
	r.HandleFunc(PUT, path, h)
}

// DeleteFunc registers a DELETE route with a standard http.HandlerFunc.
//
// Parameters:
//   - path: URL path
//   - h: Standard http.HandlerFunc
//
// Example:
//
//	r.DeleteFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Deleted"))
//	})
func (r *Router) DeleteFunc(path string, h http.HandlerFunc) {
	r.HandleFunc(DELETE, path, h)
}

// PatchFunc registers a PATCH route with a standard http.HandlerFunc.
//
// Parameters:
//   - path: URL path
//   - h: Standard http.HandlerFunc
//
// Example:
//
//	r.PatchFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Patched"))
//	})
func (r *Router) PatchFunc(path string, h http.HandlerFunc) {
	r.HandleFunc(PATCH, path, h)
}

// AnyFunc registers a route for any HTTP method with a standard http.HandlerFunc.
//
// Parameters:
//   - path: URL path
//   - h: Standard http.HandlerFunc
//
// Example:
//
//	r.AnyFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Webhook received"))
//	})
func (r *Router) AnyFunc(path string, h http.HandlerFunc) {
	r.HandleFunc(ANY, path, h)
}

// Resource registers REST-style routes for a resource.
// This method automatically creates standard REST endpoints for a resource controller.
//
// Parameters:
//   - path: Base path for the resource (e.g., "/users")
//   - c: ResourceController implementation
//
// Example:
//
//	type UserController struct{}
//
//	func (uc *UserController) Index(w http.ResponseWriter, r *http.Request) {
//	    // List all users
//	}
//
//	func (uc *UserController) Show(w http.ResponseWriter, r *http.Request) {
//	    // Get specific user
//	}
//
//	// ... implement other methods
//
//	r := router.NewRouter()
//	r.Resource("/users", &UserController{})
//
// This creates the following routes:
//   - GET    /users          -> Index
//   - POST   /users          -> Create
//   - GET    /users/:id      -> Show
//   - PUT    /users/:id      -> Update
//   - DELETE /users/:id      -> Delete
//   - PATCH  /users/:id      -> Patch
//   - OPTIONS /users         -> Options
//   - OPTIONS /users/:id     -> Options
//   - HEAD   /users          -> Head
//   - HEAD   /users/:id      -> Head
func (r *Router) Resource(path string, c ResourceController) {
	path = strings.TrimSuffix(path, "/")

	r.Get(path, http.HandlerFunc(c.Index))
	r.Post(path, http.HandlerFunc(c.Create))
	r.Get(path+"/:id", http.HandlerFunc(c.Show))
	r.Put(path+"/:id", http.HandlerFunc(c.Update))
	r.Delete(path+"/:id", http.HandlerFunc(c.Delete))
	r.Patch(path+"/:id", http.HandlerFunc(c.Patch))
	r.Options(path, http.HandlerFunc(c.Options))
	r.Options(path+"/:id", http.HandlerFunc(c.Options))
	r.Head(path, http.HandlerFunc(c.Head))
	r.Head(path+"/:id", http.HandlerFunc(c.Head))
}

// Print prints all registered routes grouped by method.
// Wildcard routes are marked specially. This is useful for debugging and
// understanding the routing structure.
//
// Parameters:
//   - w: io.Writer to write the output to
//
// Example:
//
//	r := router.NewRouter()
//	r.Get("/users", handler)
//	r.Get("/files/*path", fileHandler)
//
//	// Print to stdout
//	r.Print(os.Stdout)
//
// Output:
//
//	Registered Routes:
//	GET    /users
//	GET    /files/*path   [wildcard]
func (r *Router) Print(w io.Writer) {
	fmt.Fprintln(w, "Registered Routes:")

	methods := make([]string, 0, len(r.routes))
	for method := range r.routes {
		methods = append(methods, method)
	}
	sort.Strings(methods)

	for _, method := range methods {
		routes := r.routes[method]
		sort.Slice(routes, func(i, j int) bool { return routes[i].Path < routes[j].Path })

		for _, route := range routes {
			label := route.Path
			if strings.Contains(route.Path, "/*") {
				label += "   [wildcard]"
			}
			fmt.Fprintf(w, "%-6s %s\n", route.Method, label)
		}
	}
}
