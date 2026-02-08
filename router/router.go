package router

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	version     uint64
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
	mm.version++
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
func (mm *MiddlewareManager) Version() uint64 {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return mm.version
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
	namedRoutes       map[string]*NamedRoute // Named routes for URL generation
	methodNotAllowed  bool
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
			r.logger = logger
		}
	}
}

// WithMethodNotAllowed enables returning 405 with Allow header when path matches another method.
func WithMethodNotAllowed(enabled bool) RouterOption {
	return func(r *Router) {
		r.methodNotAllowed = enabled
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
		namedRoutes:       make(map[string]*NamedRoute),
		routeCache:        NewRouteCache(DefaultCacheCapacity),
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

// SetMethodNotAllowed toggles 405 responses when another method matches the path.
func (r *Router) SetMethodNotAllowed(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.methodNotAllowed = enabled
}

// MethodNotAllowedEnabled reports whether 405 handling is enabled.
func (r *Router) MethodNotAllowedEnabled() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.methodNotAllowed
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
		namedRoutes:       r.namedRoutes, // Share named routes with parent
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
// If the metadata includes a Name, the route will be registered as a named route
// for reverse URL generation.
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

	// Register named route if name is provided
	if meta.Name != "" {
		fullPath := r.fullPath(path)
		r.registerNamedRoute(meta.Name, method, fullPath)
	}
}

// registerNamedRoute registers a named route for reverse URL generation
func (r *Router) registerNamedRoute(name, method, pattern string) {
	if r.namedRoutes == nil {
		r.namedRoutes = make(map[string]*NamedRoute)
	}

	// Parse parameter positions
	paramPos := make(map[string]int)
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	pos := 0
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			paramPos[part[1:]] = pos
			pos++
		} else if strings.HasPrefix(part, "*") {
			paramPos[part[1:]] = pos
			pos++
		}
	}

	r.namedRoutes[name] = &NamedRoute{
		Method:   method,
		Pattern:  pattern,
		ParamPos: paramPos,
	}
}

// URL generates a URL for a named route with the given parameters.
// Parameters are passed as alternating key-value pairs.
//
// Parameters:
//   - name: The name of the route
//   - params: Alternating parameter names and values (key1, val1, key2, val2, ...)
//
// Returns:
//   - string: The generated URL, or empty string if route not found
//
// Example:
//
//	r.Get("/users/:id", getUserHandler, router.WithRouteName("user.show"))
//	url := r.URL("user.show", "id", "123") // returns "/users/123"
//
//	r.Get("/files/*path", fileHandler, router.WithRouteName("files"))
//	url := r.URL("files", "path", "docs/readme.md") // returns "/files/docs/readme.md"
func (r *Router) URL(name string, params ...string) string {
	r.mu.RLock()
	namedRoute, exists := r.namedRoutes[name]
	r.mu.RUnlock()

	if !exists {
		return ""
	}

	// Build parameter map from variadic arguments
	paramMap := make(map[string]string)
	for i := 0; i < len(params)-1; i += 2 {
		paramMap[params[i]] = params[i+1]
	}

	// Replace parameters in pattern
	result := namedRoute.Pattern
	parts := strings.Split(strings.Trim(result, "/"), "/")
	resultParts := make([]string, 0, len(parts))

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			paramName := part[1:]
			if val, ok := paramMap[paramName]; ok {
				resultParts = append(resultParts, url.PathEscape(val))
			} else {
				resultParts = append(resultParts, part) // Keep original if not provided
			}
		} else if strings.HasPrefix(part, "*") {
			paramName := part[1:]
			if val, ok := paramMap[paramName]; ok {
				// Wildcard values may contain slashes (e.g., file paths), so don't escape slashes
				segments := strings.Split(val, "/")
				for i, seg := range segments {
					segments[i] = url.PathEscape(seg)
				}
				resultParts = append(resultParts, strings.Join(segments, "/"))
			} else {
				resultParts = append(resultParts, "") // Empty for wildcard if not provided
			}
		} else {
			resultParts = append(resultParts, part)
		}
	}

	return "/" + strings.Join(resultParts, "/")
}

// URLMust generates a URL for a named route and panics if the route doesn't exist.
// This is useful during application initialization where missing routes should be fatal.
//
// Parameters:
//   - name: The name of the route
//   - params: Alternating parameter names and values
//
// Returns:
//   - string: The generated URL
//
// Panics:
//   - If the named route doesn't exist
func (r *Router) URLMust(name string, params ...string) string {
	url := r.URL(name, params...)
	if url == "" {
		panic(fmt.Sprintf("named route %q not found", name))
	}
	return url
}

// HasRoute checks if a named route exists.
//
// Parameters:
//   - name: The name of the route to check
//
// Returns:
//   - bool: True if the route exists
func (r *Router) HasRoute(name string) bool {
	r.mu.RLock()
	_, exists := r.namedRoutes[name]
	r.mu.RUnlock()
	return exists
}

// NamedRoutes returns a copy of all registered named routes.
// This is useful for debugging and documentation generation.
//
// Returns:
//   - map[string]*NamedRoute: Copy of all named routes
func (r *Router) NamedRoutes() map[string]*NamedRoute {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*NamedRoute, len(r.namedRoutes))
	for k, v := range r.namedRoutes {
		// Create a copy
		paramPos := make(map[string]int, len(v.ParamPos))
		for pk, pv := range v.ParamPos {
			paramPos[pk] = pv
		}
		result[k] = &NamedRoute{
			Method:   v.Method,
			Pattern:  v.Pattern,
			ParamPos: paramPos,
		}
	}
	return result
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
	if r.routeCache == nil {
		return CacheStats{}
	}
	return r.routeCache.Stats()
}

// ClearCache clears all cached route matching results.
// This is useful when routes are modified after initial registration
// or for testing purposes.
//
// Example:
//
//	r.ClearCache()
func (r *Router) ClearCache() {
	if r.routeCache != nil {
		r.routeCache.Clear()
	}
}

// CacheSize returns the current number of cached entries.
//
// Returns:
//   - int: Total number of cached entries (exact + pattern)
func (r *Router) CacheSize() int {
	if r.routeCache == nil {
		return 0
	}
	return r.routeCache.Size()
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
	// Parse and normalize request path first
	path := r.normalizePath(req.URL.Path)
	cachePath := path
	if path != "/" {
		cachePath = "/" + path
	}

	// Check exact route cache first for better performance (no lock needed)
	cacheKey := req.Method + ":" + cachePath
	if cachedResult, exists := r.routeCache.Get(cacheKey); exists {
		r.serveCachedMatch(w, req, cachedResult, nil)
		return
	}

	// Check pattern cache for parameterized routes (no lock needed)
	if cachedResult, paramValues, exists := r.routeCache.GetByPattern(req.Method, cachePath); exists {
		r.serveCachedMatch(w, req, cachedResult, paramValues)
		return
	}

	// Perform route matching with minimal lock time
	result, matchedAny := r.matchRoute(req.Method, path)

	if result == nil {
		// Need lock to check methodNotAllowed setting and trees
		r.mu.RLock()
		methodNotAllowed := r.methodNotAllowed
		r.mu.RUnlock()

		if methodNotAllowed {
			r.mu.RLock()
			allowed := r.allowedMethods(path)
			r.mu.RUnlock()
			if len(allowed) > 0 {
				w.Header().Set("Allow", strings.Join(allowed, ", "))
				http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
				return
			}
		}
		http.NotFound(w, req)
		return
	}

	// Set route metadata
	if result.RouteMethod == "" {
		if matchedAny {
			result.RouteMethod = ANY
		} else {
			result.RouteMethod = req.Method
		}
	}
	if result.RoutePattern == "" {
		if path == "/" {
			result.RoutePattern = "/"
		} else {
			result.RoutePattern = "/" + path
		}
	}

	// Build parameter map
	params := r.buildParamMap(result.ParamValues, result.ParamKeys)

	// Validate parameters if validations are registered
	if params != nil {
		if r.writeValidationError(w, req, params) {
			return
		}
	}

	// Cache the matching result using appropriate strategy
	if result.RoutePattern != "" && IsParameterized(result.RoutePattern) {
		r.routeCache.SetPattern(req.Method, result.RoutePattern, result)
	} else {
		r.routeCache.Set(cacheKey, result)
	}

	// No lock needed for middleware execution
	r.applyMiddlewareAndServe(w, req, params, result)
}

// matchRoute performs route matching with minimal lock time
// Returns the match result and whether it matched the ANY method
func (r *Router) matchRoute(method, path string) (*MatchResult, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Handle root path specially
	if path == "/" {
		tree := r.trees[method]
		if tree != nil && tree.handler != nil {
			return &MatchResult{
				Handler:          tree.handler,
				RouteMiddlewares: tree.middlewares,
				RoutePattern:     "/",
				RouteMethod:      method,
			}, false
		}
		// Try ANY method
		if method != ANY {
			if anyTree := r.trees[ANY]; anyTree != nil && anyTree.handler != nil {
				return &MatchResult{
					Handler:          anyTree.handler,
					RouteMiddlewares: anyTree.middlewares,
					RoutePattern:     "/",
					RouteMethod:      ANY,
				}, true
			}
		}
		return nil, false
	}

	// Find route tree
	tree := r.trees[method]
	if tree == nil {
		tree = r.trees[ANY]
		if tree == nil {
			return nil, false
		}
	}

	// Use pooled path parts
	partsPtr := SplitPathToParts(path)
	parts := *partsPtr
	defer PutPathParts(partsPtr)

	// Try to match
	matcher := GetRouteMatcher(tree)
	result := matcher.Match(parts)
	PutRouteMatcher(matcher)

	matchedAny := false

	// Try ANY method if specific method didn't match
	if result == nil && method != ANY {
		if anyTree := r.trees[ANY]; anyTree != nil {
			anyMatcher := GetRouteMatcher(anyTree)
			result = anyMatcher.Match(parts)
			PutRouteMatcher(anyMatcher)
			if result != nil {
				matchedAny = true
			}
		}
	}

	return result, matchedAny
}

// normalizePath normalizes the request path by trimming trailing slashes.
// Uses byte-level operations to avoid strings.Trim allocation.
func (r *Router) normalizePath(path string) string {
	return fastNormalizePath(path)
}

// serveCachedMatch handles requests using cached matching results.
// It validates parameters and applies middleware before serving.
// The paramValues argument overrides result.ParamValues when non-nil
// (used by pattern cache where values are extracted at lookup time).
func (r *Router) serveCachedMatch(w http.ResponseWriter, req *http.Request, result *MatchResult, paramValues []string) {
	if paramValues == nil {
		paramValues = result.ParamValues
	}

	params := r.buildParamMap(paramValues, result.ParamKeys)

	if params != nil {
		if r.writeValidationError(w, req, params) {
			return
		}
	}

	r.applyMiddlewareAndServe(w, req, params, result)
}

// writeValidationError validates route parameters and writes a 400 error if validation fails.
// Returns true if an error was written (caller should return), false otherwise.
func (r *Router) writeValidationError(w http.ResponseWriter, req *http.Request, params map[string]string) bool {
	path := r.normalizePath(req.URL.Path)
	fullPath := r.prefix + path
	if err := r.validateRouteParams(req.Method, fullPath, params); err != nil {
		contract.WriteError(w, req, contract.APIError{
			Status:   http.StatusBadRequest,
			Code:     "VALIDATION_ERROR",
			Message:  err.Error(),
			Category: contract.CategoryValidation,
		})
		return true
	}
	return false
}

// buildParamMap creates a parameter map from values and keys.
// Uses a pooled map to reduce allocations on the hot path.
func (r *Router) buildParamMap(paramValues []string, paramKeys []string) map[string]string {
	return buildParamMapPooled(paramValues, paramKeys)
}

// applyMiddlewareAndServe applies middleware chain to the handler and serves the request
func (r *Router) applyMiddlewareAndServe(w http.ResponseWriter, req *http.Request, params map[string]string, result *MatchResult) {
	reqWithParams := req
	ctx := req.Context()

	// Build RequestContext — always install so downstream code has a predictable
	// place to read/write request-scoped data.
	existingRC, ok := ctx.Value(contract.RequestContextKey{}).(contract.RequestContext)
	if !ok {
		existingRC = contract.RequestContext{}
	}

	if len(params) > 0 {
		existingRC.Params = params
	}
	if result != nil {
		if result.RoutePattern != "" {
			existingRC.RoutePattern = result.RoutePattern
		}
		meta := r.routeMeta[r.routeKey(result.RouteMethod, result.RoutePattern)]
		if meta.Name != "" {
			existingRC.RouteName = meta.Name
		}
	}

	// Use a single context.WithValue layer that carries both params and RequestContext.
	// This replaces up to 2 separate WithValue calls, reducing context chain depth.
	if len(params) > 0 {
		ctx = context.WithValue(ctx, contract.ParamsContextKey{}, params)
	}
	ctx = context.WithValue(ctx, contract.RequestContextKey{}, existingRC)
	if ctx != req.Context() {
		reqWithParams = req.WithContext(ctx)
	}

	if result == nil {
		return
	}

	version := r.middlewareManager.Version()
	if cached, ok := result.loadCached(version); ok {
		cached.ServeHTTP(w, reqWithParams)
		return
	}

	// Combine middleware slices directly
	allMiddlewares := r.middlewareManager.GetMiddlewares()
	combined := make([]middleware.Middleware, 0, len(allMiddlewares)+len(result.RouteMiddlewares))
	combined = append(combined, allMiddlewares...)
	combined = append(combined, result.RouteMiddlewares...)

	if len(combined) == 0 {
		result.Handler.ServeHTTP(w, reqWithParams)
		return
	}

	chain := middleware.NewChain(combined...)
	wrappedHandler := chain.Apply(result.Handler)
	result.storeCached(version, wrappedHandler)
	wrappedHandler.ServeHTTP(w, reqWithParams)
}

func (r *Router) allowedMethods(path string) []string {
	normalized := path
	if normalized == "" {
		normalized = "/"
	}

	allowed := make([]string, 0, len(r.trees))

	if normalized == "/" {
		for method, tree := range r.trees {
			if method == ANY || tree == nil {
				continue
			}
			if tree.handler != nil {
				allowed = append(allowed, method)
			}
		}
	} else {
		// Use pooled path parts
		partsPtr := SplitPathToParts(normalized)
		parts := *partsPtr

		for method, tree := range r.trees {
			if method == ANY || tree == nil {
				continue
			}
			matcher := GetRouteMatcher(tree)
			if matcher.Match(parts) != nil {
				allowed = append(allowed, method)
			}
			PutRouteMatcher(matcher)
		}

		PutPathParts(partsPtr)
	}

	if len(allowed) > 1 {
		sort.Strings(allowed)
	}
	return allowed
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

// HTTP method-specific route registration
// These convenience methods return errors for proper error handling.

// Get registers a GET route with the given path and handler.
// GET requests are used to retrieve resources.
//
// Parameters:
//   - path: URL path (can include parameters like "/users/:id")
//   - handler: HTTP handler
//
// Returns:
//   - error: Error if route registration fails
//
// Example:
//
//	if err := r.Get("/users", listUsersHandler); err != nil {
//	    log.Fatal(err)
//	}
func (r *Router) Get(path string, handler Handler) error { return r.AddRoute(GET, path, handler) }

// Post registers a POST route with the given path and handler.
// POST requests are used to create new resources.
//
// Parameters:
//   - path: URL path
//   - handler: HTTP handler
//
// Returns:
//   - error: Error if route registration fails
//
// Example:
//
//	if err := r.Post("/users", createUserHandler); err != nil {
//	    log.Fatal(err)
//	}
func (r *Router) Post(path string, handler Handler) error { return r.AddRoute(POST, path, handler) }

// Put registers a PUT route with the given path and handler.
// PUT requests are used to replace existing resources.
//
// Parameters:
//   - path: URL path (typically includes resource ID like "/users/:id")
//   - handler: HTTP handler
//
// Returns:
//   - error: Error if route registration fails
//
// Example:
//
//	if err := r.Put("/users/:id", updateUserHandler); err != nil {
//	    log.Fatal(err)
//	}
func (r *Router) Put(path string, handler Handler) error { return r.AddRoute(PUT, path, handler) }

// Delete registers a DELETE route with the given path and handler.
// DELETE requests are used to remove resources.
//
// Parameters:
//   - path: URL path (typically includes resource ID like "/users/:id")
//   - handler: HTTP handler
//
// Returns:
//   - error: Error if route registration fails
//
// Example:
//
//	if err := r.Delete("/users/:id", deleteUserHandler); err != nil {
//	    log.Fatal(err)
//	}
func (r *Router) Delete(path string, handler Handler) error {
	return r.AddRoute(DELETE, path, handler)
}

// Patch registers a PATCH route with the given path and handler.
// PATCH requests are used for partial updates to resources.
//
// Parameters:
//   - path: URL path (typically includes resource ID like "/users/:id")
//   - handler: HTTP handler
//
// Returns:
//   - error: Error if route registration fails
//
// Example:
//
//	if err := r.Patch("/users/:id", patchUserHandler); err != nil {
//	    log.Fatal(err)
//	}
func (r *Router) Patch(path string, handler Handler) error { return r.AddRoute(PATCH, path, handler) }

// Any registers a route that accepts any HTTP method with the given path and handler.
// This is useful for catch-all routes or when you want to handle all methods the same way.
//
// Parameters:
//   - path: URL path
//   - handler: HTTP handler
//
// Returns:
//   - error: Error if route registration fails
//
// Example:
//
//	if err := r.Any("/health", healthCheckHandler); err != nil {
//	    log.Fatal(err)
//	}
func (r *Router) Any(path string, handler Handler) error { return r.AddRoute(ANY, path, handler) }

// Options registers an OPTIONS route with the given path and handler.
// OPTIONS requests are used to describe communication options for the target resource.
//
// Parameters:
//   - path: URL path
//   - handler: HTTP handler
//
// Returns:
//   - error: Error if route registration fails
//
// Example:
//
//	if err := r.Options("/users", optionsHandler); err != nil {
//	    log.Fatal(err)
//	}
func (r *Router) Options(path string, handler Handler) error {
	return r.AddRoute("OPTIONS", path, handler)
}

// Head registers a HEAD route with the given path and handler.
// HEAD requests are identical to GET requests but without the response body.
//
// Parameters:
//   - path: URL path
//   - handler: HTTP handler
//
// Returns:
//   - error: Error if route registration fails
//
// Example:
//
//	if err := r.Head("/users", headHandler); err != nil {
//	    log.Fatal(err)
//	}
func (r *Router) Head(path string, handler Handler) error { return r.AddRoute("HEAD", path, handler) }

// Context-aware handler registration helpers
// These methods register routes with context-aware handlers that receive
// a request context and can access route parameters and other request-scoped data.

// GetCtx registers a GET route with a context-aware handler.
func (r *Router) GetCtx(path string, handler contract.CtxHandlerFunc) error {
	return r.addCtxRoute(GET, path, handler)
}

// PostCtx registers a POST route with a context-aware handler.
func (r *Router) PostCtx(path string, handler contract.CtxHandlerFunc) error {
	return r.addCtxRoute(POST, path, handler)
}

// PutCtx registers a PUT route with a context-aware handler.
func (r *Router) PutCtx(path string, handler contract.CtxHandlerFunc) error {
	return r.addCtxRoute(PUT, path, handler)
}

// DeleteCtx registers a DELETE route with a context-aware handler.
func (r *Router) DeleteCtx(path string, handler contract.CtxHandlerFunc) error {
	return r.addCtxRoute(DELETE, path, handler)
}

// PatchCtx registers a PATCH route with a context-aware handler.
func (r *Router) PatchCtx(path string, handler contract.CtxHandlerFunc) error {
	return r.addCtxRoute(PATCH, path, handler)
}

// AnyCtx registers a route that accepts any HTTP method with a context-aware handler.
func (r *Router) AnyCtx(path string, handler contract.CtxHandlerFunc) error {
	return r.addCtxRoute(ANY, path, handler)
}

// HandleFunc registers a standard http.HandlerFunc for the given path and method.
func (r *Router) HandleFunc(method, path string, h http.HandlerFunc) error {
	return r.AddRoute(method, path, h)
}

// Handle registers a standard http.Handler for the given path and method.
func (r *Router) Handle(method, path string, h http.Handler) error {
	return r.AddRoute(method, path, h)
}

// HandleWithOptions registers a standard http.Handler for the given path and method with metadata.
func (r *Router) HandleWithOptions(method, path string, h http.Handler, opts ...RouteOption) error {
	return r.AddRouteWithOptions(method, path, h, opts...)
}

// GetFunc registers a GET route with a standard http.HandlerFunc.
func (r *Router) GetFunc(path string, h http.HandlerFunc) error {
	return r.HandleFunc(GET, path, h)
}

// PostFunc registers a POST route with a standard http.HandlerFunc.
func (r *Router) PostFunc(path string, h http.HandlerFunc) error {
	return r.HandleFunc(POST, path, h)
}

// PutFunc registers a PUT route with a standard http.HandlerFunc.
//
// Parameters:
//   - path: URL path
//   - h: Standard http.HandlerFunc
//
// Returns:
//   - error: Error if route registration fails
//
// Example:
//
//	if err := r.PutFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Updated"))
//	}); err != nil {
//	    log.Fatal(err)
//	}
func (r *Router) PutFunc(path string, h http.HandlerFunc) error {
	return r.HandleFunc(PUT, path, h)
}

// DeleteFunc registers a DELETE route with a standard http.HandlerFunc.
//
// Parameters:
//   - path: URL path
//   - h: Standard http.HandlerFunc
//
// Returns:
//   - error: Error if route registration fails
//
// Example:
//
//	if err := r.DeleteFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Deleted"))
//	}); err != nil {
//	    log.Fatal(err)
//	}
func (r *Router) DeleteFunc(path string, h http.HandlerFunc) error {
	return r.HandleFunc(DELETE, path, h)
}

// PatchFunc registers a PATCH route with a standard http.HandlerFunc.
//
// Parameters:
//   - path: URL path
//   - h: Standard http.HandlerFunc
//
// Returns:
//   - error: Error if route registration fails
//
// Example:
//
//	if err := r.PatchFunc("/users/:id", func(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Patched"))
//	}); err != nil {
//	    log.Fatal(err)
//	}
func (r *Router) PatchFunc(path string, h http.HandlerFunc) error {
	return r.HandleFunc(PATCH, path, h)
}

// AnyFunc registers a route for any HTTP method with a standard http.HandlerFunc.
//
// Parameters:
//   - path: URL path
//   - h: Standard http.HandlerFunc
//
// Returns:
//   - error: Error if route registration fails
//
// Example:
//
//	if err := r.AnyFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Webhook received"))
//	}); err != nil {
//	    log.Fatal(err)
//	}
func (r *Router) AnyFunc(path string, h http.HandlerFunc) error {
	return r.HandleFunc(ANY, path, h)
}

// Resource registers REST-style routes for a resource.
// This method automatically creates standard REST endpoints for a resource controller.
//
// Parameters:
//   - path: Base path for the resource (e.g., "/users")
//   - c: ResourceController implementation
//
// Returns:
//   - error: Error if any route registration fails
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
//	if err := r.Resource("/users", &UserController{}); err != nil {
//	    log.Fatal(err)
//	}
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
func (r *Router) Resource(path string, c ResourceController) error {
	path = strings.TrimSuffix(path, "/")

	if err := r.Get(path, http.HandlerFunc(c.Index)); err != nil {
		return err
	}
	if err := r.Post(path, http.HandlerFunc(c.Create)); err != nil {
		return err
	}
	if err := r.Get(path+"/:id", http.HandlerFunc(c.Show)); err != nil {
		return err
	}
	if err := r.Put(path+"/:id", http.HandlerFunc(c.Update)); err != nil {
		return err
	}
	if err := r.Delete(path+"/:id", http.HandlerFunc(c.Delete)); err != nil {
		return err
	}
	if err := r.Patch(path+"/:id", http.HandlerFunc(c.Patch)); err != nil {
		return err
	}
	if err := r.Options(path, http.HandlerFunc(c.Options)); err != nil {
		return err
	}
	if err := r.Options(path+"/:id", http.HandlerFunc(c.Options)); err != nil {
		return err
	}
	if err := r.Head(path, http.HandlerFunc(c.Head)); err != nil {
		return err
	}
	if err := r.Head(path+"/:id", http.HandlerFunc(c.Head)); err != nil {
		return err
	}
	return nil
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
