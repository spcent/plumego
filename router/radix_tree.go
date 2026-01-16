package router

import (
	"strings"
	"sync"

	"github.com/spcent/plumego/middleware"
)

// RadixNode represents a node in the radix tree.
// Each node corresponds to a path segment in the URL and stores routing information.
// The radix tree structure enables efficient route matching with O(log n) complexity.
//
// Node types:
//   - Static: Regular path segment (e.g., "/users")
//   - Parameter: Dynamic segment with colon prefix (e.g., ":id")
//   - Wildcard: Catch-all segment with asterisk prefix (e.g., "*path")
//
// Example radix tree for routes:
//   - GET /users
//   - GET /users/:id
//   - GET /files/*path
//
// Tree structure:
//
//	(root)
//	├── users (static)
//	│   ├── (handler for GET /users)
//	│   └── :id (parameter)
//	│       └── (handler for GET /users/:id)
//	└── files (static)
//	    └── *path (wildcard)
//	        └── (handler for GET /files/*path)
type RadixNode struct {
	path        string                  // Path segment (e.g., "users", ":id", "*path")
	fullPath    string                  // Full path for this node (e.g., "/users/:id")
	indices     string                  // Child indices (first char of each child path) for fast lookup
	children    []*RadixNode            // Child nodes
	handler     Handler                 // Handler for this node (nil if not a terminal node)
	paramKeys   []string                // Parameter keys for this route (e.g., ["id"])
	priority    int                     // Priority for node ordering (higher = more specific)
	middlewares []middleware.Middleware // Route-specific middlewares
	isParam     bool                    // Whether this is a parameter node (path starts with ":")
	isWild      bool                    // Whether this is a wildcard node (path starts with "*")
}

// RadixTree implements an efficient radix tree router.
// This data structure provides fast route matching with O(log n) complexity for static routes
// and O(n) for parameterized routes. It's the core routing engine used by the Router.
//
// Features:
//   - Efficient route matching using prefix tree structure
//   - Support for static, parameter, and wildcard routes
//   - Thread-safe concurrent access
//   - Route caching for improved performance
//
// Example:
//
//	tree := NewRadixTree()
//	tree.Insert("GET", "/users", handler, []string{}, nil)
//	tree.Insert("GET", "/users/:id", handler, []string{"id"}, nil)
//
//	result := tree.Find("GET", "/users/123")
//	// result.Handler contains the matched handler
//	// result.ParamValues contains ["123"]
//	// result.ParamKeys contains ["id"]
type RadixTree struct {
	root map[string]*RadixNode // Method -> root node
	mu   sync.RWMutex
}

// NewRadixTree creates a new radix tree.
// The tree is ready to accept route insertions immediately after creation.
//
// Returns:
//   - *RadixTree: A new radix tree instance
//
// Example:
//
//	tree := NewRadixTree()
//	tree.Insert("GET", "/users", handler, []string{}, nil)
func NewRadixTree() *RadixTree {
	return &RadixTree{
		root: make(map[string]*RadixNode),
	}
}

// Insert adds a route to the radix tree.
// This method parses the path into segments and builds the tree structure.
// It panics if a duplicate route is registered.
//
// Parameters:
//   - method: HTTP method (GET, POST, etc.)
//   - path: URL path (e.g., "/users/:id")
//   - handler: HTTP handler for the route
//   - paramKeys: Parameter keys extracted from the path (e.g., ["id"])
//   - middlewares: Route-specific middlewares
//
// Example:
//
//	tree := NewRadixTree()
//	tree.Insert("GET", "/users/:id", handler, []string{"id"}, nil)
func (rt *RadixTree) Insert(method, path string, handler Handler, paramKeys []string, middlewares []middleware.Middleware) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.root[method] == nil {
		rt.root[method] = &RadixNode{}
	}

	root := rt.root[method]
	root.priority++

	// Handle root path
	if path == "/" {
		if root.handler != nil {
			panic("duplicate route: " + method + " /")
		}
		root.handler = handler
		root.fullPath = path
		return
	}

	segments := compilePathSegments(path)
	current := root

	for _, seg := range segments {
		var child *RadixNode

		if seg.isParam {
			child = rt.findOrCreateParamChild(current, seg.paramName)
		} else if seg.isWild {
			child = rt.findOrCreateWildChild(current, seg.paramName)
		} else {
			child = rt.findOrCreateStaticChild(current, seg.raw)
		}

		current = child
		current.priority++
	}

	// Set handler on final node
	if current.handler != nil {
		panic("duplicate route: " + method + " " + path)
	}
	current.handler = handler
	current.paramKeys = paramKeys
	current.fullPath = path
	current.middlewares = middlewares
}

// Find matches a path against the radix tree.
// This method performs efficient route matching using the radix tree structure.
// It returns a MatchResult containing the matched handler and extracted parameters.
//
// Parameters:
//   - method: HTTP method
//   - path: URL path to match
//
// Returns:
//   - *MatchResult: Match result with handler and parameters, or nil if no match
//
// Example:
//
//	tree := NewRadixTree()
//	tree.Insert("GET", "/users/:id", handler, []string{"id"}, nil)
//
//	result := tree.Find("GET", "/users/123")
//	// result.Handler == handler
//	// result.ParamValues == ["123"]
//	// result.ParamKeys == ["id"]
func (rt *RadixTree) Find(method, path string) *MatchResult {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	root := rt.root[method]
	if root == nil {
		root = rt.root[ANY]
		if root == nil {
			return nil
		}
	}

	if path == "/" {
		if root.handler != nil {
			return &MatchResult{
				Handler:          root.handler,
				ParamValues:      nil,
				ParamKeys:        root.paramKeys,
				RouteMiddlewares: root.middlewares,
			}
		}
		return nil
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	return rt.matchRecursive(root, parts, 0, nil, nil)
}

// matchRecursive performs recursive matching in the radix tree.
// This is the core matching algorithm that traverses the tree structure.
// It tries static children first, then parameter children, then wildcard children.
//
// Parameters:
//   - node: Current node in the tree
//   - parts: URL path segments
//   - idx: Current segment index
//   - paramValues: Accumulated parameter values
//   - paramKeys: Accumulated parameter keys
//
// Returns:
//   - *MatchResult: Match result or nil if no match
func (rt *RadixTree) matchRecursive(node *RadixNode, parts []string, idx int, paramValues []string, paramKeys []string) *MatchResult {
	if idx >= len(parts) {
		if node.handler != nil {
			return &MatchResult{
				Handler:          node.handler,
				ParamValues:      paramValues,
				ParamKeys:        paramKeys,
				RouteMiddlewares: node.middlewares,
			}
		}
		return nil
	}

	currentPart := parts[idx]

	// 1. Try static children first (most common case)
	for _, child := range node.children {
		if !child.isParam && !child.isWild && child.path == currentPart {
			result := rt.matchRecursive(child, parts, idx+1, paramValues, paramKeys)
			if result != nil {
				return result
			}
		}
	}

	// 2. Try parameter children (dynamic segments)
	for _, child := range node.children {
		if child.isParam {
			newParamValues := append(paramValues, currentPart)
			newParamKeys := append(paramKeys, child.paramKeys...)
			result := rt.matchRecursive(child, parts, idx+1, newParamValues, newParamKeys)
			if result != nil {
				return result
			}
		}
	}

	// 3. Try wildcard children (catch-all segments)
	for _, child := range node.children {
		if child.isWild {
			// Join remaining parts
			remaining := strings.Join(parts[idx:], "/")
			newParamValues := append(paramValues, remaining)
			newParamKeys := append(paramKeys, child.paramKeys...)
			if child.handler != nil {
				return &MatchResult{
					Handler:          child.handler,
					ParamValues:      newParamValues,
					ParamKeys:        newParamKeys,
					RouteMiddlewares: child.middlewares,
				}
			}
		}
	}

	return nil
}

// findOrCreateStaticChild finds or creates a static child node.
// Static nodes represent regular path segments without parameters.
//
// Parameters:
//   - parent: Parent node
//   - path: Static path segment
//
// Returns:
//   - *RadixNode: Existing or newly created child node
func (rt *RadixTree) findOrCreateStaticChild(parent *RadixNode, path string) *RadixNode {
	// Check if child already exists
	for _, child := range parent.children {
		if !child.isParam && !child.isWild && child.path == path {
			return child
		}
	}

	// Create new child
	child := &RadixNode{
		path:    path,
		isParam: false,
		isWild:  false,
	}

	// Insert in sorted order by path
	rt.insertChildSorted(parent, child)
	return child
}

// findOrCreateParamChild finds or creates a parameter child node.
// Parameter nodes represent dynamic path segments (e.g., ":id").
// Only one parameter node can exist per parent (first match wins).
//
// Parameters:
//   - parent: Parent node
//   - paramName: Parameter name (e.g., "id")
//
// Returns:
//   - *RadixNode: Existing or newly created parameter node
func (rt *RadixTree) findOrCreateParamChild(parent *RadixNode, paramName string) *RadixNode {
	// Check if param child already exists
	for _, child := range parent.children {
		if child.isParam {
			return child
		}
	}

	// Create new param child
	child := &RadixNode{
		path:      ":" + paramName,
		isParam:   true,
		paramKeys: []string{paramName},
	}

	parent.children = append(parent.children, child)
	return child
}

// findOrCreateWildChild finds or creates a wildcard child node.
// Wildcard nodes represent catch-all path segments (e.g., "*path").
// Only one wildcard node can exist per parent (first match wins).
//
// Parameters:
//   - parent: Parent node
//   - paramName: Parameter name (e.g., "path")
//
// Returns:
//   - *RadixNode: Existing or newly created wildcard node
func (rt *RadixTree) findOrCreateWildChild(parent *RadixNode, paramName string) *RadixNode {
	// Check if wild child already exists
	for _, child := range parent.children {
		if child.isWild {
			return child
		}
	}

	// Create new wild child
	child := &RadixNode{
		path:      "*" + paramName,
		isWild:    true,
		paramKeys: []string{paramName},
	}

	parent.children = append(parent.children, child)
	return child
}

// insertChildSorted inserts a child node in sorted order.
// This maintains the radix tree's ordering where static nodes come before
// parameter and wildcard nodes for deterministic matching.
//
// Parameters:
//   - parent: Parent node
//   - child: Child node to insert
func (rt *RadixTree) insertChildSorted(parent *RadixNode, child *RadixNode) {
	// Find insertion point
	i := 0
	for ; i < len(parent.children); i++ {
		if !parent.children[i].isParam && !parent.children[i].isWild {
			if parent.children[i].path > child.path {
				break
			}
		} else if parent.children[i].isParam || parent.children[i].isWild {
			// Static nodes come before param/wild nodes
			break
		}
	}

	// Insert at position
	parent.children = append(parent.children[:i], append([]*RadixNode{child}, parent.children[i:]...)...)
}

// GetRoot returns the root node for a method.
// This method is primarily used for testing and debugging purposes.
//
// Parameters:
//   - method: HTTP method
//
// Returns:
//   - *RadixNode: Root node for the method, or nil if not found
//
// Example:
//
//	tree := NewRadixTree()
//	tree.Insert("GET", "/users", handler, []string{}, nil)
//	root := tree.GetRoot("GET")
func (rt *RadixTree) GetRoot(method string) *RadixNode {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.root[method]
}
