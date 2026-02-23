# Basic Routing

> **Package**: `github.com/spcent/plumego/router`

This document covers the fundamentals of route registration and HTTP method handling in Plumego.

---

## Table of Contents

- [HTTP Methods](#http-methods)
- [Handler Types](#handler-types)
- [Route Registration](#route-registration)
- [Static Routes](#static-routes)
- [Any Method](#any-method)
- [Route Naming](#route-naming)
- [Best Practices](#best-practices)

---

## HTTP Methods

Plumego supports all standard HTTP methods:

```go
app.Get(path, handler)      // GET
app.Post(path, handler)     // POST
app.Put(path, handler)      // PUT
app.Patch(path, handler)    // PATCH
app.Delete(path, handler)   // DELETE
app.Head(path, handler)     // HEAD
app.Options(path, handler)  // OPTIONS
```

### Method Examples

```go
package main

import (
    "net/http"
    "github.com/spcent/plumego/core"
)

func main() {
    app := core.New()

    // GET - Retrieve resources
    app.Get("/users", listUsers)
    app.Get("/users/:id", getUser)

    // POST - Create resources
    app.Post("/users", createUser)

    // PUT - Update entire resource
    app.Put("/users/:id", updateUser)

    // PATCH - Partial update
    app.Patch("/users/:id", patchUser)

    // DELETE - Remove resources
    app.Delete("/users/:id", deleteUser)

    // HEAD - Headers only (like GET but no body)
    app.Head("/users/:id", headUser)

    // OPTIONS - Supported methods
    app.Options("/users", optionsUsers)

    app.Boot()
}
```

---

## Handler Types

Plumego supports multiple handler signatures for flexibility.

### 1. Standard Library Handler

```go
func(w http.ResponseWriter, r *http.Request)
```

**Example**:
```go
app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("pong"))
})
```

**When to use**: Simple handlers, compatibility with standard library

### 2. Context-Aware Handler

```go
func(ctx *plumego.Context)
```

**Example**:
```go
import "github.com/spcent/plumego"

app.GetCtx("/users/:id", func(ctx *plumego.Context) {
    id := ctx.Param("id")
    ctx.JSON(http.StatusOK, map[string]string{
        "id": id,
    })
})
```

**When to use**: JSON APIs, need for context helpers

### 3. http.Handler Interface

```go
type Handler interface {
    ServeHTTP(w http.ResponseWriter, r *http.Request)
}
```

**Example**:
```go
type UserHandler struct {
    db *sql.DB
}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Handle request
}

app.Get("/users", &UserHandler{db: db})
```

**When to use**: Stateful handlers, dependency injection

---

## Route Registration

### Direct Registration

```go
app := core.New()

// GET route
app.Get("/", homeHandler)

// POST route
app.Post("/users", createUserHandler)

// Multiple routes for same path
app.Get("/users", listUsers)
app.Post("/users", createUser)
```

### Method Chaining

```go
// Not supported, use separate calls
app.Get("/users", listUsers)
app.Post("/users", createUser)
```

### Via Router

```go
router := app.Router()

router.Get("/users", listUsers)
router.Post("/users", createUser)
```

---

## Static Routes

Static routes have fixed paths without parameters.

### Simple Static Routes

```go
app.Get("/", homeHandler)
app.Get("/about", aboutHandler)
app.Get("/contact", contactHandler)
```

### Nested Static Routes

```go
app.Get("/api/users", listUsers)
app.Get("/api/products", listProducts)
app.Get("/api/orders", listOrders)
```

### Static Files

```go
// Serve directory
app.Static("/assets", "./public")

// Serve embedded files
//go:embed static/*
var staticFS embed.FS

app.StaticFS("/static", http.FS(staticFS))
```

---

## Any Method

Handle any HTTP method with a single route.

### Basic Usage

```go
app.Any("/webhook", func(w http.ResponseWriter, r *http.Request) {
    method := r.Method
    fmt.Fprintf(w, "Received %s request", method)
})
```

### Method Switching

```go
app.Any("/users/:id", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        getUser(w, r)
    case http.MethodPut:
        updateUser(w, r)
    case http.MethodDelete:
        deleteUser(w, r)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
})
```

### When to Use

Use `Any` for:
- Webhook endpoints that accept multiple methods
- Dynamic method handling
- Simplified route definitions

**Warning**: Prefer specific methods (GET, POST, etc.) for clarity and RESTful design.

---

## Route Naming

Named routes enable reverse routing (URL generation).

### Basic Naming

```go
// Register with name
app.GetNamed("home", "/", homeHandler)
app.GetNamed("user.show", "/users/:id", getUserHandler)
app.PostNamed("user.create", "/users", createUserHandler)
```

### Generating URLs

```go
// Simple route
url := app.Router().URL("home")
// Result: /

// Route with parameters
url := app.Router().URL("user.show", map[string]string{
    "id": "123",
})
// Result: /users/123
```

### Naming Conventions

```go
// Resource-based naming
app.GetNamed("users.index", "/users", listUsers)
app.GetNamed("users.show", "/users/:id", getUser)
app.PostNamed("users.store", "/users", createUser)
app.PutNamed("users.update", "/users/:id", updateUser)
app.DeleteNamed("users.destroy", "/users/:id", deleteUser)

// Dot notation for hierarchy
app.GetNamed("api.v1.users.index", "/api/v1/users", listUsers)
```

---

## Complete Examples

### REST API

```go
package main

import (
    "encoding/json"
    "net/http"
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego"
)

type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

var users = []User{
    {ID: "1", Name: "Alice"},
    {ID: "2", Name: "Bob"},
}

func main() {
    app := core.New()

    // List all users
    app.Get("/users", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(users)
    })

    // Get single user
    app.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
        id := plumego.Param(r, "id")

        for _, user := range users {
            if user.ID == id {
                w.Header().Set("Content-Type", "application/json")
                json.NewEncoder(w).Encode(user)
                return
            }
        }

        http.Error(w, "User not found", http.StatusNotFound)
    })

    // Create user
    app.Post("/users", func(w http.ResponseWriter, r *http.Request) {
        var user User
        if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        users = append(users, user)

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(user)
    })

    // Update user
    app.Put("/users/:id", func(w http.ResponseWriter, r *http.Request) {
        id := plumego.Param(r, "id")

        var updatedUser User
        if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        for i, user := range users {
            if user.ID == id {
                users[i] = updatedUser
                w.Header().Set("Content-Type", "application/json")
                json.NewEncoder(w).Encode(updatedUser)
                return
            }
        }

        http.Error(w, "User not found", http.StatusNotFound)
    })

    // Delete user
    app.Delete("/users/:id", func(w http.ResponseWriter, r *http.Request) {
        id := plumego.Param(r, "id")

        for i, user := range users {
            if user.ID == id {
                users = append(users[:i], users[i+1:]...)
                w.WriteHeader(http.StatusNoContent)
                return
            }
        }

        http.Error(w, "User not found", http.StatusNotFound)
    })

    app.Boot()
}
```

### With Context Handlers

```go
package main

import (
    "net/http"
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego"
)

func main() {
    app := core.New()

    // Using context-aware handlers
    app.GetCtx("/users", func(ctx *plumego.Context) {
        ctx.JSON(http.StatusOK, users)
    })

    app.GetCtx("/users/:id", func(ctx *plumego.Context) {
        id := ctx.Param("id")

        for _, user := range users {
            if user.ID == id {
                ctx.JSON(http.StatusOK, user)
                return
            }
        }

        ctx.Error(http.StatusNotFound, "User not found")
    })

    app.PostCtx("/users", func(ctx *plumego.Context) {
        var user User
        if err := ctx.Bind(&user); err != nil {
            ctx.Error(http.StatusBadRequest, err.Error())
            return
        }

        users = append(users, user)
        ctx.Status(http.StatusCreated).JSON(user)
    })

    app.Boot()
}
```

### Stateful Handlers

```go
package main

import (
    "database/sql"
    "net/http"
    "github.com/spcent/plumego/core"
)

type UserHandler struct {
    db *sql.DB
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
    rows, err := h.db.Query("SELECT id, name FROM users")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    // Process rows...
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    // Insert into database...
}

func main() {
    db, _ := sql.Open("postgres", "...")
    handler := &UserHandler{db: db}

    app := core.New()

    app.Get("/users", handler.List)
    app.Post("/users", handler.Create)

    app.Boot()
}
```

---

## Best Practices

### ✅ Do

1. **Use Specific Methods**
   ```go
   // ✅ Clear and RESTful
   app.Get("/users", listUsers)
   app.Post("/users", createUser)
   ```

2. **Follow REST Conventions**
   ```go
   app.Get("/resources", list)        // List all
   app.Get("/resources/:id", get)     // Get one
   app.Post("/resources", create)     // Create
   app.Put("/resources/:id", update)  // Update
   app.Delete("/resources/:id", del)  // Delete
   ```

3. **Use Context Handlers for APIs**
   ```go
   app.GetCtx("/api/users", func(ctx *plumego.Context) {
       ctx.JSON(200, users)
   })
   ```

4. **Name Important Routes**
   ```go
   app.GetNamed("user.profile", "/profile", profileHandler)
   ```

### ❌ Don't

1. **Don't Overuse `Any`**
   ```go
   // ❌ Unclear intent
   app.Any("/users", handler)

   // ✅ Explicit methods
   app.Get("/users", listUsers)
   app.Post("/users", createUsers)
   ```

2. **Don't Mix Handler Types**
   ```go
   // ❌ Inconsistent
   app.Get("/users", standardHandler)
   app.GetCtx("/products", contextHandler)

   // ✅ Consistent
   app.GetCtx("/users", handler1)
   app.GetCtx("/products", handler2)
   ```

3. **Don't Ignore Errors**
   ```go
   // ❌ No error handling
   json.NewDecoder(r.Body).Decode(&user)

   // ✅ Handle errors
   if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
       http.Error(w, err.Error(), http.StatusBadRequest)
       return
   }
   ```

---

## Common Patterns

### Health Check Endpoints

```go
app.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok"}`))
})

app.Get("/health/ready", readinessCheck)
app.Get("/health/live", livenessCheck)
```

### Version Info

```go
app.Get("/version", func(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]string{
        "version": "1.0.0",
        "build":   "abc123",
    })
})
```

### Redirects

```go
app.Get("/old-path", func(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/new-path", http.StatusMovedPermanently)
})
```

### Favicon

```go
app.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "./static/favicon.ico")
})
```

---

## Testing Routes

```go
package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/spcent/plumego/core"
)

func TestRoutes(t *testing.T) {
    app := core.New()
    app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("pong"))
    })

    tests := []struct {
        method string
        path   string
        want   int
    }{
        {"GET", "/ping", http.StatusOK},
        {"GET", "/notfound", http.StatusNotFound},
        {"POST", "/ping", http.StatusMethodNotAllowed},
    }

    for _, tt := range tests {
        req := httptest.NewRequest(tt.method, tt.path, nil)
        w := httptest.NewRecorder()

        app.Router().ServeHTTP(w, req)

        if w.Code != tt.want {
            t.Errorf("%s %s: got %d, want %d", tt.method, tt.path, w.Code, tt.want)
        }
    }
}
```

---

## Next Steps

- **[Route Groups](route-groups.md)** - Organize routes with prefixes
- **[Path Parameters](path-parameters.md)** - Extract URL parameters
- **[Middleware Binding](middleware-binding.md)** - Add per-route middleware
- **[Reverse Routing](reverse-routing.md)** - Generate URLs from routes

---

**Related**:
- [Router Overview](README.md)
- [Core Module](../core/)
- [Contract Module](../contract/)
