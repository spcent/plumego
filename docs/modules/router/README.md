# Router Module

> **Package**: `github.com/spcent/plumego/router`
> **Stability**: High - Routing API is stable
> **Go Version**: 1.24+

The `router` package provides a fast, flexible HTTP router built on a radix tree (trie) data structure. It supports path parameters, route groups, middleware binding, and reverse routing.

---

## Overview

The router is the core request dispatch mechanism in Plumego, responsible for matching incoming HTTP requests to registered handlers based on method and path.

### Key Features

- **Fast Matching**: Radix tree algorithm with O(log n) lookup time
- **Path Parameters**: Named parameters (`:id`) and wildcards (`*path`)
- **Route Groups**: Hierarchical route organization with prefix and middleware
- **Reverse Routing**: Generate URLs from route names
- **Middleware Binding**: Per-route or per-group middleware
- **Method-Based**: Separate routes for GET, POST, PUT, DELETE, etc.
- **Zero Allocations**: Optimized for performance

---

## Quick Start

```go
package main

import (
    "net/http"
    "github.com/spcent/plumego/core"
)

func main() {
    app := core.New()

    // Basic routes
    app.Get("/", homeHandler)
    app.Get("/users/:id", getUserHandler)
    app.Post("/users", createUserHandler)

    // Route groups
    api := app.Group("/api/v1")
    api.Get("/products", listProductsHandler)
    api.Get("/products/:id", getProductHandler)

    app.Boot()
}
```

---

## Core Concepts

### 1. Router

The `Router` struct manages all routes and performs request matching.

```go
type Router struct {
    // Routes organized by HTTP method
    trees map[string]*node
    // Named routes for reverse routing
    routes map[string]*Route
}
```

### 2. Route

A `Route` represents a single endpoint:

```go
type Route struct {
    Method      string
    Path        string
    Handler     http.Handler
    Middleware  []func(http.Handler) http.Handler
    Name        string
}
```

### 3. Group

A `Group` provides hierarchical route organization:

```go
type Group struct {
    prefix     string
    router     *Router
    middleware []func(http.Handler) http.Handler
}
```

### 4. Path Parameters

Extract values from URL paths:

```go
// Route: /users/:id
// Request: /users/123
id := plumego.Param(r, "id") // "123"

// Route: /files/*path
// Request: /files/docs/readme.md
path := plumego.Param(r, "path") // "docs/readme.md"
```

---

## Documentation Structure

| Document | Description |
|----------|-------------|
| **[Basic Routing](basic-routing.md)** | HTTP methods, route registration, handlers |
| **[Route Groups](route-groups.md)** | Hierarchical organization, prefixes |
| **[Path Parameters](path-parameters.md)** | Named params, wildcards, extraction |
| **[Reverse Routing](reverse-routing.md)** | URL generation from route names |
| **[Middleware Binding](middleware-binding.md)** | Per-route and per-group middleware |
| **[Trie Router](trie-router.md)** | Internal algorithm and performance |
| **[Advanced Patterns](advanced-patterns.md)** | Custom matchers, constraints |

---

## Usage Examples

### Basic Routes

```go
app := core.New()

// HTTP methods
app.Get("/users", listUsers)
app.Post("/users", createUser)
app.Put("/users/:id", updateUser)
app.Delete("/users/:id", deleteUser)

// Any method
app.Any("/webhook", handleWebhook)
```

### Path Parameters

```go
// Named parameter
app.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
    id := plumego.Param(r, "id")
    fmt.Fprintf(w, "User ID: %s", id)
})

// Multiple parameters
app.Get("/posts/:category/:id", func(w http.ResponseWriter, r *http.Request) {
    category := plumego.Param(r, "category")
    id := plumego.Param(r, "id")
    fmt.Fprintf(w, "Category: %s, ID: %s", category, id)
})

// Wildcard
app.Get("/files/*path", func(w http.ResponseWriter, r *http.Request) {
    path := plumego.Param(r, "path")
    fmt.Fprintf(w, "File path: %s", path)
})
```

### Route Groups

```go
app := core.New()

// API v1
v1 := app.Group("/api/v1")
v1.Get("/users", listUsers)
v1.Get("/users/:id", getUser)

// API v2
v2 := app.Group("/api/v2")
v2.Get("/users", listUsersV2)
v2.Get("/users/:id", getUserV2)

// Admin with middleware
admin := app.Group("/admin", authMiddleware)
admin.Get("/dashboard", adminDashboard)
admin.Post("/settings", updateSettings)
```

### Nested Groups

```go
api := app.Group("/api")

v1 := api.Group("/v1")
v1.Get("/users", listUsers)

v2 := api.Group("/v2")
v2.Get("/users", listUsersV2)

// Results in:
// /api/v1/users
// /api/v2/users
```

### Middleware Binding

```go
// Global middleware
app.Use(loggingMiddleware)

// Group middleware
authenticated := app.Group("/api", authMiddleware)
authenticated.Get("/profile", getProfile)

// Per-route middleware
app.Get("/admin",
    requireAdmin(
        http.HandlerFunc(adminHandler),
    ),
)
```

### Named Routes

```go
// Register with name
app.GetNamed("user.show", "/users/:id", getUserHandler)

// Generate URL
url := app.Router().URL("user.show", map[string]string{
    "id": "123",
})
// Result: /users/123
```

---

## Performance

### Benchmarks

```
BenchmarkRouter_StaticRoutes-8       50000000    25.3 ns/op    0 B/op    0 allocs/op
BenchmarkRouter_ParamRoutes-8        20000000    65.8 ns/op    0 B/op    0 allocs/op
BenchmarkRouter_WildcardRoutes-8     10000000   125.0 ns/op    0 B/op    0 allocs/op
```

### Characteristics

- **O(log n)** lookup time using radix tree
- **Zero allocations** for static routes
- **Minimal allocations** for parameterized routes
- **Efficient memory usage** with tree compression

---

## Router Options

```go
import "github.com/spcent/plumego/router"

// Create custom router
r := router.New(
    router.WithCaseSensitive(false),     // Case-insensitive matching
    router.WithStrictSlash(false),       // Ignore trailing slashes
    router.WithRedirectTrailingSlash(true), // Auto-redirect /foo/ to /foo
    router.WithNotFoundHandler(custom404), // Custom 404 handler
    router.WithMethodNotAllowedHandler(custom405), // Custom 405 handler
)

app := core.New(
    core.WithRouter(r),
)
```

---

## Route Matching

### Priority Order

1. **Static routes**: `/users/profile`
2. **Named parameters**: `/users/:id`
3. **Wildcards**: `/files/*path`

### Examples

```go
// Registered routes
app.Get("/users/new", handleNew)      // 1. Static - highest priority
app.Get("/users/:id", handleUser)     // 2. Parameter
app.Get("/users/*action", handleAny)  // 3. Wildcard - lowest priority

// Request: GET /users/new
// Matches: handleNew (static)

// Request: GET /users/123
// Matches: handleUser (parameter)

// Request: GET /users/profile/edit
// Matches: handleAny (wildcard)
```

---

## Common Patterns

### REST API

```go
// Users resource
app.Get("/users", listUsers)           // List all
app.Post("/users", createUser)         // Create
app.Get("/users/:id", getUser)         // Get one
app.Put("/users/:id", updateUser)      // Update
app.Delete("/users/:id", deleteUser)   // Delete
```

### Nested Resources

```go
// Posts with comments
app.Get("/posts/:post_id/comments", listComments)
app.Post("/posts/:post_id/comments", createComment)
app.Get("/posts/:post_id/comments/:id", getComment)
app.Delete("/posts/:post_id/comments/:id", deleteComment)
```

### API Versioning

```go
v1 := app.Group("/api/v1")
v1.Get("/users", listUsersV1)

v2 := app.Group("/api/v2")
v2.Get("/users", listUsersV2)
```

### Health Checks

```go
app.Get("/health", healthCheck)
app.Get("/health/ready", readinessCheck)
app.Get("/health/live", livenessCheck)
```

---

## Error Handling

### 404 Not Found

```go
// Default behavior
// Returns: 404 Not Found

// Custom handler
app := core.New(
    core.WithRouter(router.New(
        router.WithNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "Route not found",
                "path":  r.URL.Path,
            })
        })),
    )),
)
```

### 405 Method Not Allowed

```go
// When route exists but method doesn't match
// Default: Returns 405 with Allow header

// Custom handler
router.WithMethodNotAllowedHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusMethodNotAllowed)
    json.NewEncoder(w).Encode(map[string]string{
        "error":  "Method not allowed",
        "method": r.Method,
        "path":   r.URL.Path,
    })
}))
```

---

## Best Practices

### ✅ Do

- Use route groups for organization
- Name routes for reverse routing
- Follow REST conventions
- Use path parameters for IDs
- Use wildcards sparingly

### ❌ Don't

- Don't use wildcards for simple parameters
- Don't create conflicting routes
- Don't hardcode URLs in handlers
- Don't ignore route conflicts warnings

---

## Troubleshooting

### Route Not Matching

```go
// Problem: Route not matching despite correct path
app.Get("/users/:id", handler)
// Request: GET /users/

// Solution: Trailing slash matters by default
app.Get("/users/:id/", handler)  // With trailing slash
// Or configure router to ignore trailing slashes
router.WithStrictSlash(false)
```

### Parameter Not Found

```go
// Problem: Param returns empty string
id := plumego.Param(r, "id")

// Solution: Check route definition
app.Get("/users/:userId", handler)  // Parameter named "userId"
id := plumego.Param(r, "userId")    // Must match parameter name
```

### Route Conflicts

```go
// Problem: Routes conflict
app.Get("/users/:id", handler1)
app.Get("/users/:name", handler2)  // Conflict!

// Solution: Use distinct static parts
app.Get("/users/id/:id", handler1)
app.Get("/users/name/:name", handler2)
```

---

## Next Steps

- **[Basic Routing](basic-routing.md)** - Learn route registration
- **[Route Groups](route-groups.md)** - Organize routes hierarchically
- **[Path Parameters](path-parameters.md)** - Extract URL parameters
- **[Reverse Routing](reverse-routing.md)** - Generate URLs from routes

---

## Related Modules

- **[Core](../core/)** - Application and lifecycle
- **[Middleware](../middleware/)** - Request processing
- **[Contract](../contract/)** - Context and error handling

---

**Stability**: High - Breaking changes require major version bump
**Maintainers**: Plumego Core Team
**Last Updated**: 2026-02-11
