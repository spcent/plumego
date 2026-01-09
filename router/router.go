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
	"github.com/spcent/plumego/middleware"
)

// HTTP method constants
const (
	GET    = "GET"    // HTTP GET method
	POST   = "POST"   // HTTP POST method
	PUT    = "PUT"    // HTTP PUT method
	DELETE = "DELETE" // HTTP DELETE method
	PATCH  = "PATCH"  // HTTP PATCH method
	ANY    = "ANY"    // Any HTTP method
)

// Handler is an alias to the standard http.Handler for route handlers.
// Use HandlerFunc when registering inline functions.
type Handler = http.Handler

// HandlerFunc is an alias to the standard http.HandlerFunc for convenience.
type HandlerFunc = http.HandlerFunc

// RouteRegistrar defines an interface for objects that can register routes with a router
type RouteRegistrar interface {
	Register(r *Router)
}

// MiddlewareManager manages middleware chain for routes and groups
type MiddlewareManager struct {
	middlewares []middleware.Middleware
	mu          sync.RWMutex
}

// NewMiddlewareManager creates a new middleware manager
func NewMiddlewareManager() *MiddlewareManager {
	return &MiddlewareManager{
		middlewares: make([]middleware.Middleware, 0),
	}
}

// AddMiddleware adds a middleware to the manager
func (mm *MiddlewareManager) AddMiddleware(m middleware.Middleware) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.middlewares = append(mm.middlewares, m)
}

// GetMiddlewares returns a copy of all middlewares
func (mm *MiddlewareManager) GetMiddlewares() []middleware.Middleware {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	result := make([]middleware.Middleware, len(mm.middlewares))
	copy(result, mm.middlewares)
	return result
}

// MergeMiddlewares merges another middleware manager's middlewares
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

// Router represents an HTTP router with path-based routing and middleware support
type Router struct {
	prefix            string                      // Group prefix
	trees             map[string]*node            // Method -> root node
	registrars        []RouteRegistrar            // Route registrars
	routes            map[string][]route          // Registered routes
	frozen            bool                        // Whether router is frozen
	mu                sync.RWMutex                // Mutex for concurrent access
	parent            *Router                     // Parent router for groups
	middlewareManager *MiddlewareManager          // Middleware management
	logger            log.StructuredLogger        // Logger for contextual handlers
	routeCache        *RouteCache                 // Route matching cache
	routeValidations  map[string]*RouteValidation // Route parameter validations
	validationIndex   map[string][]validationEntry
	routeMeta         map[string]RouteMeta
}

// RouterOption defines a function type for router configuration options
type RouterOption func(*Router)

// RouteOption defines an option for route metadata.
type RouteOption func(*RouteMeta)

// WithRouteName sets a route name.
func WithRouteName(name string) RouteOption {
	return func(meta *RouteMeta) {
		meta.Name = name
	}
}

// WithRouteTags sets route tags.
func WithRouteTags(tags ...string) RouteOption {
	return func(meta *RouteMeta) {
		meta.Tags = append([]string(nil), tags...)
	}
}

// WithLogger sets a custom logger for the router
func WithLogger(logger log.StructuredLogger) RouterOption {
	return func(r *Router) {
		if logger != nil {
			r.logger = logger
		}
	}
}

// NewRouter creates a new Router instance with default configuration
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
func (r *Router) SetLogger(logger log.StructuredLogger) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logger = logger
}

// Freeze prevents the router from accepting new route registrations
func (r *Router) Freeze() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.frozen = true
}

// Register adds route registrars to the router
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

// Group creates a new router group with the given prefix
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

// AddRoute adds a route to the router with the given method, path and handler
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
		// Get segment value
		segValue := seg.raw
		if seg.isParam {
			segValue = ":"
			paramKeys = append(paramKeys, seg.paramName)
		} else if seg.isWild {
			segValue = "*"
			paramKeys = append(paramKeys, seg.paramName)
		}

		// Find or create child node
		child := r.findChild(current, segValue)
		if child == nil {
			child = &node{
				path: segValue,
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

// Use adds middlewares to the router group
func (r *Router) Use(middlewares ...middleware.Middleware) {
	for _, middleware := range middlewares {
		r.middlewareManager.AddMiddleware(middleware)
	}
}

// findChild finds a child node with the given path segment
func (r *Router) findChild(parent *node, path string) *node {
	// Check if it's a wildcard segment
	if len(path) == 1 {
		switch path {
		case ":":
			// Check for param child
			for _, child := range parent.children {
				if len(child.path) > 0 && child.path[0] == ':' {
					return child
				}
			}
		case "*":
			// Check for wild child
			for _, child := range parent.children {
				if len(child.path) > 0 && child.path[0] == '*' {
					return child
				}
			}
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

// ServeHTTP implements http.Handler and handles incoming HTTP requests
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

// Get registers a GET route with the given path and handler
func (r *Router) Get(path string, handler Handler) { r.addRoute(GET, path, handler) }

// Post registers a POST route with the given path and handler
func (r *Router) Post(path string, handler Handler) { r.addRoute(POST, path, handler) }

// Put registers a PUT route with the given path and handler
func (r *Router) Put(path string, handler Handler) { r.addRoute(PUT, path, handler) }

// Delete registers a DELETE route with the given path and handler
func (r *Router) Delete(path string, handler Handler) { r.addRoute(DELETE, path, handler) }

// Patch registers a PATCH route with the given path and handler
func (r *Router) Patch(path string, handler Handler) { r.addRoute(PATCH, path, handler) }

// Any registers a route that accepts any HTTP method with the given path and handler
func (r *Router) Any(path string, handler Handler) { r.addRoute(ANY, path, handler) }

// Options registers an OPTIONS route with the given path and handler
func (r *Router) Options(path string, handler Handler) { r.addRoute("OPTIONS", path, handler) }

// Head registers a HEAD route with the given path and handler
func (r *Router) Head(path string, handler Handler) { r.addRoute("HEAD", path, handler) }

// Context-aware handler registration helpers

// GetCtx registers a GET route with a context-aware handler
func (r *Router) GetCtx(path string, handler contract.CtxHandlerFunc) {
	r.addCtxRoute(GET, path, handler)
}

// PostCtx registers a POST route with a context-aware handler
func (r *Router) PostCtx(path string, handler contract.CtxHandlerFunc) {
	r.addCtxRoute(POST, path, handler)
}

// PutCtx registers a PUT route with a context-aware handler
func (r *Router) PutCtx(path string, handler contract.CtxHandlerFunc) {
	r.addCtxRoute(PUT, path, handler)
}

// DeleteCtx registers a DELETE route with a context-aware handler
func (r *Router) DeleteCtx(path string, handler contract.CtxHandlerFunc) {
	r.addCtxRoute(DELETE, path, handler)
}

// PatchCtx registers a PATCH route with a context-aware handler
func (r *Router) PatchCtx(path string, handler contract.CtxHandlerFunc) {
	r.addCtxRoute(PATCH, path, handler)
}

// AnyCtx registers a route that accepts any HTTP method with a context-aware handler
func (r *Router) AnyCtx(path string, handler contract.CtxHandlerFunc) {
	r.addCtxRoute(ANY, path, handler)
}

// HandleFunc registers a standard http.HandlerFunc for the given path and method
func (r *Router) HandleFunc(method, path string, h http.HandlerFunc) {
	r.AddRoute(method, path, h)
}

// Handle registers a standard http.Handler for the given path and method
func (r *Router) Handle(method, path string, h http.Handler) {
	r.AddRoute(method, path, h)
}

// HandleWithOptions registers a standard http.Handler for the given path and method with metadata.
func (r *Router) HandleWithOptions(method, path string, h http.Handler, opts ...RouteOption) {
	_ = r.AddRouteWithOptions(method, path, h, opts...)
}

// GetFunc registers a GET route with a standard http.HandlerFunc
func (r *Router) GetFunc(path string, h http.HandlerFunc) {
	r.HandleFunc(GET, path, h)
}

// PostFunc registers a POST route with a standard http.HandlerFunc
func (r *Router) PostFunc(path string, h http.HandlerFunc) {
	r.HandleFunc(POST, path, h)
}

// PutFunc registers a PUT route with a standard http.HandlerFunc
func (r *Router) PutFunc(path string, h http.HandlerFunc) {
	r.HandleFunc(PUT, path, h)
}

// DeleteFunc registers a DELETE route with a standard http.HandlerFunc
func (r *Router) DeleteFunc(path string, h http.HandlerFunc) {
	r.HandleFunc(DELETE, path, h)
}

// PatchFunc registers a PATCH route with a standard http.HandlerFunc
func (r *Router) PatchFunc(path string, h http.HandlerFunc) {
	r.HandleFunc(PATCH, path, h)
}

// AnyFunc registers a route for any HTTP method with a standard http.HandlerFunc
func (r *Router) AnyFunc(path string, h http.HandlerFunc) {
	r.HandleFunc(ANY, path, h)
}

// Resource REST-style routes
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
// Wildcard routes are marked specially.
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
