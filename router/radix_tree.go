package router

import (
	"strings"
	"sync"

	"github.com/spcent/plumego/middleware"
)

// RadixNode represents a node in the radix tree
type RadixNode struct {
	path        string                  // Path segment
	fullPath    string                  // Full path for this node
	indices     string                  // Child indices (first char of each child path)
	children    []*RadixNode            // Child nodes
	handler     Handler                 // Handler for this node
	paramKeys   []string                // Parameter keys
	priority    int                     // Priority for node ordering
	middlewares []middleware.Middleware // Route-specific middlewares
	isParam     bool                    // Whether this is a parameter node
	isWild      bool                    // Whether this is a wildcard node
}

// RadixTree implements an efficient radix tree router
type RadixTree struct {
	root map[string]*RadixNode // Method -> root node
	mu   sync.RWMutex
}

// NewRadixTree creates a new radix tree
func NewRadixTree() *RadixTree {
	return &RadixTree{
		root: make(map[string]*RadixNode),
	}
}

// Insert adds a route to the radix tree
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

// Find matches a path against the radix tree
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

// matchRecursive performs recursive matching
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

	// 1. Try static children first
	for _, child := range node.children {
		if !child.isParam && !child.isWild && child.path == currentPart {
			result := rt.matchRecursive(child, parts, idx+1, paramValues, paramKeys)
			if result != nil {
				return result
			}
		}
	}

	// 2. Try parameter children
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

	// 3. Try wildcard children
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

// findOrCreateStaticChild finds or creates a static child node
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

// findOrCreateParamChild finds or creates a parameter child node
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

// findOrCreateWildChild finds or creates a wildcard child node
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

// insertChildSorted inserts a child node in sorted order
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

// GetRoot returns the root node for a method (for testing)
func (rt *RadixTree) GetRoot(method string) *RadixNode {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.root[method]
}
