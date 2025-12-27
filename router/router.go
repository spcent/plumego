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

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
	PATCH  = "PATCH"
	ANY    = "ANY"
)

// Handler is an alias to the standard http.Handler for route handlers.
// Use HandlerFunc when registering inline functions.
type Handler = http.Handler

// HandlerFunc is an alias to the standard http.HandlerFunc for convenience.
type HandlerFunc = http.HandlerFunc

type RouteRegistrar interface {
	Register(r *Router)
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

type Router struct {
	prefix      string                  // Group prefix
	trees       map[string]*node        // Method -> root node
	registrars  []RouteRegistrar        // Route registrars
	routes      map[string][]route      // Registered routes
	frozen      bool                    // Whether router is frozen
	mu          sync.RWMutex            // Mutex for concurrent access
	parent      *Router                 // Parent router for groups
	middlewares []middleware.Middleware // Group-level middlewares
	logger      log.StructuredLogger    // Logger for contextual handlers
}

func NewRouter() *Router {
	return &Router{
		trees:       make(map[string]*node),
		routes:      make(map[string][]route),
		prefix:      "",
		parent:      nil,
		middlewares: []middleware.Middleware{},
		logger:      log.NewGLogger(),
	}
}

// SetLogger configures the logger used by context-aware handlers.
func (r *Router) SetLogger(logger log.StructuredLogger) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logger = logger
}

func (r *Router) Freeze() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.frozen = true
}

func (r *Router) Register(registrars ...RouteRegistrar) {
	// Get unique registrars without holding the lock
	// This prevents deadlock when reg.Register(r) calls back into AddRoute
	var newRegistrars []RouteRegistrar

	r.mu.RLock()
	seen := make(map[RouteRegistrar]bool)
	for _, reg := range r.registrars {
		seen[reg] = true
	}
	r.mu.RUnlock()

	// Find new registrars
	for _, reg := range registrars {
		if !seen[reg] {
			newRegistrars = append(newRegistrars, reg)
			seen[reg] = true
		}
	}

	// Register new registrars
	for _, reg := range newRegistrars {
		// Add to registrars list
		r.mu.Lock()
		r.registrars = append(r.registrars, reg)
		r.mu.Unlock()

		// Call Register without holding the lock
		// This allows reg.Register(r) to call AddRoute safely
		reg.Register(r)
	}
}

// Group creates a new route group with the given prefix and inherits parent middlewares
func (r *Router) Group(prefix string) *Router {
	// Create full prefix by combining with parent's prefix
	fullPrefix := r.prefix + prefix

	// Inherit middlewares from parent
	inheritedMiddlewares := make([]middleware.Middleware, len(r.middlewares))
	copy(inheritedMiddlewares, r.middlewares)

	return &Router{
		prefix:      fullPrefix,
		trees:       r.trees,
		routes:      r.routes,
		frozen:      r.frozen,
		parent:      r,
		middlewares: inheritedMiddlewares,
	}
}

// AddRoute adds a route to the router with the given method, path and handler
func (r *Router) AddRoute(method, path string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.frozen {
		panic("router is frozen, cannot add route after freeze")
	}

	// Combine with group prefix
	fullPath := r.prefix + strings.TrimRight(path, "/")
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
			panic("duplicate route registration: " + method + " /")
		}
		current.handler = handler
		current.fullPath = fullPath
		r.routes[method] = append(r.routes[method], route{Method: method, Path: fullPath})
		return
	}

	segments := compileTemplate(fullPath)
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
		panic("duplicate route registration: " + method + " " + fullPath)
	}
	current.handler = handler
	current.paramKeys = paramKeys
	current.fullPath = fullPath
	current.middlewares = r.routeMiddlewares()

	r.routes[method] = append(r.routes[method], route{Method: method, Path: fullPath})
}

func (r *Router) routeMiddlewares() []middleware.Middleware {
	if r.parent == nil || len(r.middlewares) == 0 {
		return nil
	}

	parentLen := len(r.parent.middlewares)
	if parentLen >= len(r.middlewares) {
		return nil
	}

	extra := make([]middleware.Middleware, len(r.middlewares)-parentLen)
	copy(extra, r.middlewares[parentLen:])
	return extra
}

// Use adds middlewares to the router group
func (r *Router) Use(middlewares ...middleware.Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Add middlewares to the group
	r.middlewares = append(r.middlewares, middlewares...)
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

func (r *Router) addCtxRoute(method, path string, h contract.CtxHandlerFunc) {
	if err := contract.ValidateCtxHandler(h); err != nil {
		panic(err.Error())
	}

	r.AddRoute(method, path, contract.AdaptCtxHandler(h, r.logger))
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	method := req.Method
	tree := r.trees[method]
	if tree == nil {
		tree = r.trees[ANY]
		if tree == nil {
			http.NotFound(w, req)
			return
		}
	}

	path := req.URL.Path
	if path != "/" {
		path = strings.Trim(path, "/")
	}

	// Handle root path specially
	if path == "/" || path == "" {
		if tree.handler != nil {
			// Apply middleware and serve request for root path
			r.applyMiddlewareAndServe(w, req, nil, tree.handler, tree.middlewares)
			return
		}
		http.NotFound(w, req)
		return
	}

	parts := strings.Split(path, "/")
	params := make(map[string]string)

	// Perform trie-based route matching and get handler, param values, and param keys
	handler, paramValues, paramKeys, routeMiddlewares := r.matchRoute(tree, parts)
	if handler == nil && method != ANY {
		if anyTree := r.trees[ANY]; anyTree != nil {
			handler, paramValues, paramKeys, routeMiddlewares = r.matchRoute(anyTree, parts)
		}
	}
	if handler == nil {
		http.NotFound(w, req)
		return
	}

	// Assign parameter values to map
	if len(paramValues) > 0 && len(paramKeys) > 0 {
		for i, key := range paramKeys {
			if i < len(paramValues) {
				params[key] = paramValues[i]
			}
		}
	}

	// Apply middleware and serve request
	r.applyMiddlewareAndServe(w, req, params, handler, routeMiddlewares)
}

// applyMiddlewareAndServe applies middleware chain to the handler and serves the request
func (r *Router) applyMiddlewareAndServe(w http.ResponseWriter, req *http.Request, params map[string]string, handler Handler, routeMiddlewares []middleware.Middleware) {
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

	combined := append([]middleware.Middleware{}, r.middlewares...)
	combined = append(combined, routeMiddlewares...)

	if len(combined) == 0 {
		handler.ServeHTTP(w, reqWithParams)
		return
	}

	chain := middleware.NewChain(combined...)
	wrappedHandler := chain.Apply(handler)
	wrappedHandler.ServeHTTP(w, reqWithParams)
}

// matchRoute performs efficient trie-based route matching
func (r *Router) matchRoute(root *node, parts []string) (Handler, []string, []string, []middleware.Middleware) {
	current := root
	paramValues := make([]string, 0, len(parts))

	for i, part := range parts {
		// Try to find exact match first
		child := r.findChildForPath(current, part)
		if child != nil {
			current = child
			continue
		}

		// Try param match
		paramChild := r.findParamChild(current)
		if paramChild != nil {
			paramValues = append(paramValues, part)
			current = paramChild
			continue
		}

		// Try wildcard match
		wildChild := r.findWildChild(current)
		if wildChild != nil {
			wildValue := strings.Join(parts[i:], "/")
			paramValues = append(paramValues, wildValue)
			current = wildChild
			break
		}

		// No match found
		return nil, nil, nil, nil
	}

	// Use the paramKeys stored in the node during route registration
	return current.handler, paramValues, current.paramKeys, current.middlewares
}

// findChildForPath finds a child node that matches the given path segment
func (r *Router) findChildForPath(parent *node, path string) *node {
	// Check exact match first
	for _, child := range parent.children {
		if child.path == path {
			return child
		}
	}
	return nil
}

// findParamChild finds a param child node if exists
func (r *Router) findParamChild(parent *node) *node {
	for _, child := range parent.children {
		if len(child.path) > 0 && child.path[0] == ':' {
			return child
		}
	}
	return nil
}

// findWildChild finds a wildcard child node if exists
func (r *Router) findWildChild(parent *node) *node {
	for _, child := range parent.children {
		if len(child.path) > 0 && child.path[0] == '*' {
			return child
		}
	}
	return nil
}

func compileTemplate(path string) []segment {
	if path == "/" {
		return nil
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	segments := make([]segment, 0, len(parts))
	for _, p := range parts {
		if strings.HasPrefix(p, ":") {
			segments = append(segments, segment{
				raw:       p,
				isParam:   true,
				paramName: p[1:],
			})
		} else if strings.HasPrefix(p, "*") {
			segments = append(segments, segment{
				raw:       p,
				isWild:    true,
				paramName: p[1:],
			})
		} else {
			segments = append(segments, segment{raw: p})
		}
	}
	return segments
}

// --- Helper API ---

func (r *Router) Get(path string, h Handler)    { r.AddRoute(GET, path, h) }
func (r *Router) Post(path string, h Handler)   { r.AddRoute(POST, path, h) }
func (r *Router) Put(path string, h Handler)    { r.AddRoute(PUT, path, h) }
func (r *Router) Delete(path string, h Handler) { r.AddRoute(DELETE, path, h) }
func (r *Router) Patch(path string, h Handler)  { r.AddRoute(PATCH, path, h) }
func (r *Router) Any(path string, h Handler)    { r.AddRoute(ANY, path, h) }

// Context-aware handler registration helpers
func (r *Router) GetCtx(path string, h contract.CtxHandlerFunc)    { r.addCtxRoute(GET, path, h) }
func (r *Router) PostCtx(path string, h contract.CtxHandlerFunc)   { r.addCtxRoute(POST, path, h) }
func (r *Router) PutCtx(path string, h contract.CtxHandlerFunc)    { r.addCtxRoute(PUT, path, h) }
func (r *Router) DeleteCtx(path string, h contract.CtxHandlerFunc) { r.addCtxRoute(DELETE, path, h) }
func (r *Router) PatchCtx(path string, h contract.CtxHandlerFunc)  { r.addCtxRoute(PATCH, path, h) }
func (r *Router) AnyCtx(path string, h contract.CtxHandlerFunc)    { r.addCtxRoute(ANY, path, h) }

// HandleFunc registers a standard http.HandlerFunc for the given path and method
func (r *Router) HandleFunc(method, path string, h http.HandlerFunc) {
	r.AddRoute(method, path, h)
}

// Handle registers a standard http.Handler for the given path and method
func (r *Router) Handle(method, path string, h http.Handler) {
	r.AddRoute(method, path, h)
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
}

// Print prints all registered routes grouped by method.
// Wildcard routes are marked specially.
func (r *Router) Print(w io.Writer) {
	fmt.Fprintln(w, "Registered Routes:")

	methods := make([]string, 0, len(r.routes))
	for m := range r.routes {
		methods = append(methods, m)
	}
	sort.Strings(methods)

	for _, m := range methods {
		rs := r.routes[m]
		sort.Slice(rs, func(i, j int) bool { return rs[i].Path < rs[j].Path })

		for _, rt := range rs {
			label := rt.Path
			if strings.Contains(rt.Path, "/*") {
				label += "   [wildcard]"
			}
			fmt.Fprintf(w, "%-6s %s\n", rt.Method, label)
		}
	}
}
