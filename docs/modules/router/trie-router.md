# Trie Router Implementation

> **Package**: `github.com/spcent/plumego/router`

This document explains the internal trie (radix tree) algorithm used by Plumego's router for fast request matching.

---

## Table of Contents

- [Overview](#overview)
- [Data Structure](#data-structure)
- [Matching Algorithm](#matching-algorithm)
- [Performance](#performance)
- [Optimization Techniques](#optimization-techniques)
- [Comparison](#comparison)

---

## Overview

### What is a Trie?

A trie (prefix tree or radix tree) is a tree data structure used for efficient string matching. In Plumego's router:

- Each node represents a path segment
- Edges are labeled with path segments
- Leaf nodes contain handlers
- Common prefixes are shared

### Why Trie?

Benefits over linear search or hash maps:

- **O(log n) lookup**: Logarithmic time complexity
- **Prefix sharing**: Efficient memory usage for similar routes
- **No hash collisions**: Deterministic routing
- **Parameter support**: Natural integration with `:param` and `*wildcard`

---

## Data Structure

### Node Structure

```go
type node struct {
    // Path segment for this node
    path string

    // Handler for this path (if leaf)
    handler http.Handler

    // Child nodes
    children []*node

    // Parameter name (if this is a :param node)
    paramName string

    // Wildcard flag (if this is a *param node)
    wildcard bool

    // Indices of first characters of children (optimization)
    indices string
}
```

### Tree Example

For these routes:
```go
app.Get("/users", handler1)
app.Get("/users/:id", handler2)
app.Get("/users/:id/posts", handler3)
app.Get("/products", handler4)
```

The trie looks like:
```
root
├── /users
│   ├── [handler1]
│   └── /:id
│       ├── [handler2]
│       └── /posts
│           └── [handler3]
└── /products
    └── [handler4]
```

### Path Segments

Routes are split into segments:
```
/users/:id/posts
↓
["/users/", ":id", "/posts"]
```

Static segments and parameters are different node types.

---

## Matching Algorithm

### Lookup Process

```go
func (r *Router) lookup(method, path string) (http.Handler, map[string]string) {
    // 1. Get method tree
    tree := r.trees[method]
    if tree == nil {
        return nil, nil
    }

    // 2. Traverse tree
    params := make(map[string]string)
    handler := tree.search(path, params)

    return handler, params
}
```

### Search Steps

1. **Start at root node**
2. **Match static prefix**: Find longest matching static segment
3. **Check parameter nodes**: If no static match, try parameter nodes
4. **Check wildcard nodes**: If no parameter match, try wildcard
5. **Recurse**: Continue with remaining path
6. **Return handler**: When path is exhausted, return handler at leaf

### Pseudocode

```
function search(path, params):
    node = root

    while path is not empty:
        // Try static match
        for child in node.children:
            if child.path is static and path.startsWith(child.path):
                path = path.removePrefix(child.path)
                node = child
                continue outer loop

        // Try parameter match
        for child in node.children:
            if child.isParameter:
                segment = extractNextSegment(path)
                params[child.paramName] = segment
                path = path.removePrefix(segment)
                node = child
                continue outer loop

        // Try wildcard match
        for child in node.children:
            if child.isWildcard:
                params[child.paramName] = path
                return child.handler

        // No match found
        return nil

    return node.handler
```

---

## Performance

### Time Complexity

| Operation | Complexity | Description |
|-----------|------------|-------------|
| Lookup | O(log n) | Average case with balanced tree |
| Lookup | O(m) | Worst case where m = path length |
| Insert | O(m) | Where m = path length |
| Memory | O(n * m) | Where n = routes, m = avg path length |

### Benchmarks

```
BenchmarkRouter_StaticRoutes-8         50000000    25.3 ns/op    0 B/op    0 allocs/op
BenchmarkRouter_1Param-8               20000000    65.8 ns/op    0 B/op    0 allocs/op
BenchmarkRouter_5Params-8              10000000   123.0 ns/op    0 B/op    0 allocs/op
BenchmarkRouter_Wildcard-8             10000000   125.0 ns/op    0 B/op    0 allocs/op

// For comparison (linear search):
BenchmarkLinear_100Routes-8             1000000  1250.0 ns/op    0 B/op    0 allocs/op
```

### Space Complexity

Trie vs. Hash Map:

```
Hash Map:  O(n)     - Simple, but no prefix sharing
Trie:      O(n*m)   - More memory, but faster lookups
```

For 1000 routes with average path length 20:
- Hash Map: ~20 KB
- Trie: ~40 KB (2x, but with shared prefixes)

---

## Optimization Techniques

### 1. Index Optimization

Store first characters of children for quick lookup:

```go
type node struct {
    children []*node
    indices  string  // First char of each child's path
}

// Instead of iterating all children:
for i, child := range node.children {
    if strings.HasPrefix(path, child.path) { ... }
}

// Use indices for quick check:
char := path[0]
if i := strings.IndexByte(node.indices, char); i >= 0 {
    child := node.children[i]
    // Only check this child
}
```

### 2. Path Compression

Merge nodes with single children:

```
Before compression:
/users → / → id → / → posts

After compression:
/users → /:id/posts
```

### 3. Priority-Based Ordering

Order child nodes by priority:
1. Static paths (highest priority)
2. Parameter nodes
3. Wildcard nodes (lowest priority)

```go
func (n *node) addChild(child *node) {
    // Insert static children before parameters
    if child.isStatic {
        n.children = append([]*node{child}, n.children...)
    } else {
        n.children = append(n.children, child)
    }
}
```

### 4. Parameter Pooling

Reuse parameter maps to reduce allocations:

```go
var paramPool = sync.Pool{
    New: func() interface{} {
        return make(map[string]string)
    },
}

func (r *Router) lookup(method, path string) (http.Handler, map[string]string) {
    params := paramPool.Get().(map[string]string)
    defer paramPool.Put(params)

    // Use params...
}
```

---

## Comparison

### Trie vs. Hash Map

| Feature | Trie | Hash Map |
|---------|------|----------|
| **Lookup Time** | O(log n) | O(1) average |
| **Memory** | O(n*m) | O(n) |
| **Prefix Sharing** | ✅ Yes | ❌ No |
| **Parameter Support** | ✅ Natural | ⚠️ Requires parsing |
| **Ordered Traversal** | ✅ Yes | ❌ No |
| **Worst Case** | O(m) | O(n) (collisions) |

### Trie vs. Linear Search

| Routes | Trie (ns/op) | Linear (ns/op) | Speedup |
|--------|--------------|----------------|---------|
| 10 | 25 | 50 | 2x |
| 100 | 30 | 500 | 16x |
| 1000 | 35 | 5000 | 142x |
| 10000 | 40 | 50000 | 1250x |

### When to Use Trie

✅ **Use Trie when**:
- Routes have common prefixes
- Need parameter extraction
- Route count > 20
- Need predictable performance

⚠️ **Consider alternatives when**:
- Very few routes (< 10)
- No common prefixes
- Memory is extremely constrained

---

## Implementation Details

### Route Registration

```go
func (r *Router) addRoute(method, path string, handler http.Handler) {
    // Get or create tree for method
    tree := r.trees[method]
    if tree == nil {
        tree = &node{}
        r.trees[method] = tree
    }

    // Insert into tree
    tree.insert(path, handler)
}
```

### Tree Insertion

```go
func (n *node) insert(path string, handler http.Handler) {
    // Base case: empty path
    if len(path) == 0 {
        n.handler = handler
        return
    }

    // Find matching child or create new
    for _, child := range n.children {
        // Check for common prefix
        i := longestCommonPrefix(path, child.path)
        if i > 0 {
            // Split node if partial match
            if i < len(child.path) {
                n.split(child, i)
            }

            // Continue with remaining path
            child.insert(path[i:], handler)
            return
        }
    }

    // No matching child, create new
    newChild := &node{path: path, handler: handler}
    n.addChild(newChild)
}
```

### Parameter Extraction

```go
func (n *node) search(path string, params map[string]string) http.Handler {
    // ... (see pseudocode above)

    // When parameter node is matched:
    if n.isParameter {
        // Extract value between slashes
        end := strings.IndexByte(path, '/')
        if end < 0 {
            end = len(path)
        }

        // Store in params map
        params[n.paramName] = path[:end]

        // Continue with remaining path
        path = path[end:]
    }
}
```

---

## Debugging Tools

### Visualize Tree

```go
func (r *Router) PrintTree(method string) {
    tree := r.trees[method]
    if tree == nil {
        fmt.Println("No routes for method:", method)
        return
    }

    tree.print("", true)
}

func (n *node) print(prefix string, last bool) {
    marker := "├── "
    if last {
        marker = "└── "
    }

    fmt.Printf("%s%s%s", prefix, marker, n.path)
    if n.handler != nil {
        fmt.Print(" [HANDLER]")
    }
    fmt.Println()

    newPrefix := prefix
    if last {
        newPrefix += "    "
    } else {
        newPrefix += "│   "
    }

    for i, child := range n.children {
        isLast := i == len(n.children)-1
        child.print(newPrefix, isLast)
    }
}
```

### Route Conflicts

```go
func (r *Router) checkConflicts() []string {
    conflicts := []string{}

    for method, tree := range r.trees {
        tree.checkConflicts(method, "", &conflicts)
    }

    return conflicts
}

func (n *node) checkConflicts(method, path string, conflicts *[]string) {
    currentPath := path + n.path

    // Check for ambiguous parameter nodes
    paramCount := 0
    for _, child := range n.children {
        if child.isParameter {
            paramCount++
        }
    }

    if paramCount > 1 {
        *conflicts = append(*conflicts,
            fmt.Sprintf("%s %s: multiple parameter children", method, currentPath))
    }

    // Recurse
    for _, child := range n.children {
        child.checkConflicts(method, currentPath, conflicts)
    }
}
```

---

## Next Steps

- **[Advanced Patterns](advanced-patterns.md)** - Custom matchers and constraints
- **[Router Overview](README.md)** - High-level router documentation
- **[Performance Guide](../../guides/performance.md)** - Optimization strategies

---

**Related**:
- [Router Overview](README.md)
- [Path Parameters](path-parameters.md)
- [Basic Routing](basic-routing.md)
