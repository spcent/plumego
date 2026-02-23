# Path Parameters

> **Package**: `github.com/spcent/plumego/router`

Path parameters allow you to extract dynamic values from URL paths. This document covers named parameters, wildcards, and parameter extraction.

---

## Table of Contents

- [Overview](#overview)
- [Named Parameters](#named-parameters)
- [Wildcard Parameters](#wildcard-parameters)
- [Parameter Extraction](#parameter-extraction)
- [Multiple Parameters](#multiple-parameters)
- [Validation](#validation)
- [Best Practices](#best-practices)

---

## Overview

### Parameter Types

Plumego supports two types of path parameters:

1. **Named Parameters** (`:name`): Match a single path segment
2. **Wildcard Parameters** (`*name`): Match remaining path segments

### Syntax

```go
// Named parameter
app.Get("/users/:id", handler)        // Matches: /users/123

// Wildcard parameter
app.Get("/files/*path", handler)      // Matches: /files/docs/readme.md
```

---

## Named Parameters

Named parameters match exactly one path segment.

### Basic Usage

```go
// Define route with parameter
app.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
    id := plumego.Param(r, "id")
    fmt.Fprintf(w, "User ID: %s", id)
})

// Matches:
// /users/123       → id = "123"
// /users/abc       → id = "abc"
// /users/user-123  → id = "user-123"

// Does NOT match:
// /users           → No parameter
// /users/123/edit  → Extra segment
```

### Parameter Naming

Parameters can use any valid identifier:

```go
app.Get("/users/:userId", handler)
app.Get("/posts/:post_id", handler)
app.Get("/articles/:articleID", handler)
app.Get("/files/:fileName", handler)
```

### Position

Parameters can appear anywhere in the path:

```go
// At the end
app.Get("/users/:id", handler)

// In the middle
app.Get("/users/:id/posts", handler)

// Multiple positions
app.Get("/users/:userId/posts/:postId", handler)
```

---

## Wildcard Parameters

Wildcard parameters match all remaining path segments.

### Basic Usage

```go
// Define route with wildcard
app.Get("/files/*path", func(w http.ResponseWriter, r *http.Request) {
    path := plumego.Param(r, "path")
    fmt.Fprintf(w, "File path: %s", path)
})

// Matches:
// /files/docs/readme.md      → path = "docs/readme.md"
// /files/a/b/c/d.txt         → path = "a/b/c/d.txt"
// /files/                    → path = ""
// /files/image.png           → path = "image.png"
```

### Restrictions

- Wildcards must be at the **end** of the path
- Only **one** wildcard per route
- Wildcards match **everything** after their position

```go
// ✅ Valid
app.Get("/files/*path", handler)
app.Get("/api/v1/files/*path", handler)

// ❌ Invalid - wildcard not at end
app.Get("/*path/files", handler)

// ❌ Invalid - multiple wildcards
app.Get("/*path1/*path2", handler)
```

---

## Parameter Extraction

### Standard Library Handler

```go
import "github.com/spcent/plumego"

app.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
    // Extract parameter
    id := plumego.Param(r, "id")

    // Use parameter
    fmt.Fprintf(w, "User ID: %s", id)
})
```

### Context-Aware Handler

```go
app.GetCtx("/users/:id", func(ctx *plumego.Context) {
    // Extract from context
    id := ctx.Param("id")

    // Use parameter
    ctx.JSON(200, map[string]string{"id": id})
})
```

### All Parameters

```go
app.Get("/users/:id/posts/:postId", func(w http.ResponseWriter, r *http.Request) {
    // Get all parameters as map
    params := plumego.Params(r)

    id := params["id"]
    postId := params["postId"]
})
```

---

## Multiple Parameters

Routes can have multiple named parameters.

### Two Parameters

```go
app.Get("/users/:userId/posts/:postId", func(w http.ResponseWriter, r *http.Request) {
    userId := plumego.Param(r, "userId")
    postId := plumego.Param(r, "postId")

    fmt.Fprintf(w, "User: %s, Post: %s", userId, postId)
})

// Example:
// /users/123/posts/456
// → userId = "123", postId = "456"
```

### Three or More Parameters

```go
app.Get("/:org/:repo/issues/:number", func(w http.ResponseWriter, r *http.Request) {
    org := plumego.Param(r, "org")
    repo := plumego.Param(r, "repo")
    number := plumego.Param(r, "number")

    fmt.Fprintf(w, "Org: %s, Repo: %s, Issue: %s", org, repo, number)
})

// Example:
// /golang/go/issues/12345
// → org = "golang", repo = "go", number = "12345"
```

### Parameter + Wildcard

```go
app.Get("/users/:id/files/*path", func(w http.ResponseWriter, r *http.Request) {
    id := plumego.Param(r, "id")
    path := plumego.Param(r, "path")

    fmt.Fprintf(w, "User: %s, Path: %s", id, path)
})

// Example:
// /users/123/files/docs/readme.md
// → id = "123", path = "docs/readme.md"
```

---

## Validation

Always validate parameters before use.

### Type Conversion

```go
app.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
    idStr := plumego.Param(r, "id")

    // Convert to integer
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }

    // Use validated ID
    user := getUserByID(id)
    json.NewEncoder(w).Encode(user)
})
```

### Format Validation

```go
import "regexp"

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

app.Get("/resources/:uuid", func(w http.ResponseWriter, r *http.Request) {
    uuid := plumego.Param(r, "uuid")

    // Validate UUID format
    if !uuidRegex.MatchString(uuid) {
        http.Error(w, "Invalid UUID format", http.StatusBadRequest)
        return
    }

    // Use validated UUID
    resource := getResource(uuid)
    json.NewEncoder(w).Encode(resource)
})
```

### Range Validation

```go
app.Get("/pages/:page", func(w http.ResponseWriter, r *http.Request) {
    pageStr := plumego.Param(r, "page")

    page, err := strconv.Atoi(pageStr)
    if err != nil || page < 1 || page > 1000 {
        http.Error(w, "Invalid page number (1-1000)", http.StatusBadRequest)
        return
    }

    // Use validated page
    results := getPage(page)
    json.NewEncoder(w).Encode(results)
})
```

### Enum Validation

```go
var validStatuses = map[string]bool{
    "active":   true,
    "inactive": true,
    "pending":  true,
}

app.Get("/users/status/:status", func(w http.ResponseWriter, r *http.Request) {
    status := plumego.Param(r, "status")

    // Validate enum
    if !validStatuses[status] {
        http.Error(w, "Invalid status", http.StatusBadRequest)
        return
    }

    // Use validated status
    users := getUsersByStatus(status)
    json.NewEncoder(w).Encode(users)
})
```

---

## Complete Examples

### User Profile API

```go
package main

import (
    "encoding/json"
    "net/http"
    "strconv"
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
    app := core.New()

    // Get user by ID
    app.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
        idStr := plumego.Param(r, "id")

        id, err := strconv.Atoi(idStr)
        if err != nil {
            http.Error(w, "Invalid user ID", http.StatusBadRequest)
            return
        }

        user := User{ID: id, Name: "John Doe"}
        json.NewEncoder(w).Encode(user)
    })

    // Get user's posts
    app.Get("/users/:id/posts", func(w http.ResponseWriter, r *http.Request) {
        idStr := plumego.Param(r, "id")
        id, _ := strconv.Atoi(idStr)

        posts := []map[string]interface{}{
            {"id": 1, "userId": id, "title": "Post 1"},
            {"id": 2, "userId": id, "title": "Post 2"},
        }
        json.NewEncoder(w).Encode(posts)
    })

    // Get specific post
    app.Get("/users/:userId/posts/:postId", func(w http.ResponseWriter, r *http.Request) {
        userId := plumego.Param(r, "userId")
        postId := plumego.Param(r, "postId")

        post := map[string]string{
            "userId": userId,
            "postId": postId,
            "title":  "Sample Post",
        }
        json.NewEncoder(w).Encode(post)
    })

    app.Boot()
}
```

### File Server with Wildcards

```go
package main

import (
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego"
)

func main() {
    app := core.New()

    // Serve files from /files/*path
    app.Get("/files/*path", func(w http.ResponseWriter, r *http.Request) {
        path := plumego.Param(r, "path")

        // Security: prevent directory traversal
        path = filepath.Clean(path)
        if filepath.IsAbs(path) {
            http.Error(w, "Invalid path", http.StatusBadRequest)
            return
        }

        // Build full path
        fullPath := filepath.Join("./public", path)

        // Check if file exists
        info, err := os.Stat(fullPath)
        if err != nil {
            http.Error(w, "File not found", http.StatusNotFound)
            return
        }

        // Don't serve directories
        if info.IsDir() {
            http.Error(w, "Cannot serve directory", http.StatusForbidden)
            return
        }

        // Serve file
        http.ServeFile(w, r, fullPath)
    })

    // File metadata endpoint
    app.Get("/files/*/info", func(w http.ResponseWriter, r *http.Request) {
        path := plumego.Param(r, "path")
        fullPath := filepath.Join("./public", filepath.Clean(path))

        info, err := os.Stat(fullPath)
        if err != nil {
            http.Error(w, "File not found", http.StatusNotFound)
            return
        }

        fmt.Fprintf(w, "Name: %s\nSize: %d bytes\n", info.Name(), info.Size())
    })

    app.Boot()
}
```

### GitHub-style Routes

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego"
)

func main() {
    app := core.New()

    // User profile
    app.Get("/:username", func(w http.ResponseWriter, r *http.Request) {
        username := plumego.Param(r, "username")
        fmt.Fprintf(w, "Profile: %s", username)
    })

    // User's repositories
    app.Get("/:username/:repo", func(w http.ResponseWriter, r *http.Request) {
        username := plumego.Param(r, "username")
        repo := plumego.Param(r, "repo")
        fmt.Fprintf(w, "Repository: %s/%s", username, repo)
    })

    // Repository file browser
    app.Get("/:username/:repo/tree/:branch/*path", func(w http.ResponseWriter, r *http.Request) {
        username := plumego.Param(r, "username")
        repo := plumego.Param(r, "repo")
        branch := plumego.Param(r, "branch")
        path := plumego.Param(r, "path")

        fmt.Fprintf(w, "Repo: %s/%s\nBranch: %s\nPath: %s",
            username, repo, branch, path)
    })

    // Repository issues
    app.Get("/:username/:repo/issues/:number", func(w http.ResponseWriter, r *http.Request) {
        username := plumego.Param(r, "username")
        repo := plumego.Param(r, "repo")
        number := plumego.Param(r, "number")

        fmt.Fprintf(w, "Issue: %s/%s#%s", username, repo, number)
    })

    app.Boot()
}
```

---

## Best Practices

### ✅ Do

1. **Use Descriptive Parameter Names**
   ```go
   // ✅ Clear
   app.Get("/users/:userId/posts/:postId", handler)

   // ❌ Unclear
   app.Get("/users/:id1/posts/:id2", handler)
   ```

2. **Validate Parameters**
   ```go
   // ✅ Always validate
   id, err := strconv.Atoi(plumego.Param(r, "id"))
   if err != nil {
       http.Error(w, "Invalid ID", 400)
       return
   }
   ```

3. **Use Named Parameters for Single Segments**
   ```go
   // ✅ Use :id for single segment
   app.Get("/users/:id", handler)

   // ❌ Don't use wildcard
   app.Get("/users/*id", handler)
   ```

4. **Sanitize Wildcards for File Paths**
   ```go
   // ✅ Clean and validate
   path := filepath.Clean(plumego.Param(r, "path"))
   if filepath.IsAbs(path) {
       return // Reject absolute paths
   }
   ```

### ❌ Don't

1. **Don't Trust Parameters Blindly**
   ```go
   // ❌ No validation
   id := plumego.Param(r, "id")
   db.Query("SELECT * FROM users WHERE id = " + id) // SQL injection!

   // ✅ Validate and use prepared statements
   id, err := strconv.Atoi(plumego.Param(r, "id"))
   if err != nil { return }
   db.Query("SELECT * FROM users WHERE id = ?", id)
   ```

2. **Don't Use Wildcards for Everything**
   ```go
   // ❌ Too broad
   app.Get("/*path", handler)

   // ✅ Specific routes
   app.Get("/files/*path", fileHandler)
   ```

3. **Don't Create Ambiguous Routes**
   ```go
   // ❌ Conflict
   app.Get("/users/:id", handler1)
   app.Get("/users/:name", handler2)

   // ✅ Disambiguate
   app.Get("/users/id/:id", handler1)
   app.Get("/users/name/:name", handler2)
   ```

---

## Route Matching Priority

When multiple routes could match, Plumego uses this priority:

1. **Static** segments (exact match)
2. **Named parameters** (`:name`)
3. **Wildcards** (`*name`)

```go
app.Get("/users/new", handler1)      // Priority 1 (static)
app.Get("/users/:id", handler2)      // Priority 2 (parameter)
app.Get("/users/*path", handler3)    // Priority 3 (wildcard)

// Request: GET /users/new
// Matches: handler1 (static wins)

// Request: GET /users/123
// Matches: handler2 (parameter)

// Request: GET /users/123/profile
// Matches: handler3 (wildcard)
```

---

## Testing Parameters

```go
package main

import (
    "net/http/httptest"
    "testing"
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego"
)

func TestPathParameters(t *testing.T) {
    app := core.New()

    app.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
        id := plumego.Param(r, "id")
        w.Write([]byte(id))
    })

    tests := []struct {
        path string
        want string
    }{
        {"/users/123", "123"},
        {"/users/abc", "abc"},
        {"/users/user-456", "user-456"},
    }

    for _, tt := range tests {
        req := httptest.NewRequest("GET", tt.path, nil)
        w := httptest.NewRecorder()

        app.Router().ServeHTTP(w, req)

        got := w.Body.String()
        if got != tt.want {
            t.Errorf("GET %s: got %q, want %q", tt.path, got, tt.want)
        }
    }
}
```

---

## Next Steps

- **[Reverse Routing](reverse-routing.md)** - Generate URLs from parameters
- **[Route Groups](route-groups.md)** - Organize parameterized routes
- **[Basic Routing](basic-routing.md)** - Route registration fundamentals

---

**Related**:
- [Router Overview](README.md)
- [Contract Module](../contract/)
- [Validator Module](../validator/)
