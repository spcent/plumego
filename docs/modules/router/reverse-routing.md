# Reverse Routing

> **Package**: `github.com/spcent/plumego/router`

Reverse routing allows you to generate URLs from route names and parameters, eliminating hardcoded URLs in your application.

---

## Table of Contents

- [Overview](#overview)
- [Named Routes](#named-routes)
- [URL Generation](#url-generation)
- [Parameters](#parameters)
- [Query Strings](#query-strings)
- [Use Cases](#use-cases)
- [Best Practices](#best-practices)

---

## Overview

### What is Reverse Routing?

Reverse routing generates URLs from route definitions rather than hardcoding paths:

```go
// ❌ Hardcoded URL
http.Redirect(w, r, "/users/123", http.StatusFound)

// ✅ Generated URL
url := router.URL("user.show", map[string]string{"id": "123"})
http.Redirect(w, r, url, http.StatusFound)
```

### Benefits

- **Maintainability**: Change route paths in one place
- **Type Safety**: Compile-time checking of route names
- **Refactoring**: Easier to update URLs across application
- **Documentation**: Route names serve as documentation

---

## Named Routes

### Registering Named Routes

```go
app := core.New()

// Name routes during registration
app.GetNamed("home", "/", homeHandler)
app.GetNamed("user.show", "/users/:id", getUserHandler)
app.PostNamed("user.create", "/users", createUserHandler)
app.PutNamed("user.update", "/users/:id", updateUserHandler)
app.DeleteNamed("user.destroy", "/users/:id", deleteUserHandler)
```

### Naming Conventions

Use consistent naming patterns:

```go
// Resource-based (Laravel style)
app.GetNamed("users.index", "/users", listUsers)
app.GetNamed("users.show", "/users/:id", getUser)
app.PostNamed("users.store", "/users", createUser)
app.PutNamed("users.update", "/users/:id", updateUser)
app.DeleteNamed("users.destroy", "/users/:id", deleteUser)

// Action-based
app.GetNamed("user.list", "/users", listUsers)
app.GetNamed("user.get", "/users/:id", getUser)
app.PostNamed("user.create", "/users", createUser)
app.PutNamed("user.update", "/users/:id", updateUser)
app.DeleteNamed("user.delete", "/users/:id", deleteUser)

// Dot notation for hierarchy
app.GetNamed("api.v1.users.index", "/api/v1/users", listUsers)
app.GetNamed("api.v1.users.show", "/api/v1/users/:id", getUser)
```

### Group Named Routes

```go
api := app.Group("/api/v1")

// Names are independent of group prefix
api.GetNamed("users.index", "/users", listUsers)  // /api/v1/users
api.GetNamed("products.index", "/products", listProducts)  // /api/v1/products
```

---

## URL Generation

### Basic Generation

```go
router := app.Router()

// Simple route (no parameters)
url := router.URL("home")
// Result: /

// Route with parameter
url := router.URL("user.show", map[string]string{
    "id": "123",
})
// Result: /users/123
```

### In Handlers

```go
app.GetCtx("/users/:id", func(ctx *plumego.Context) {
    id := ctx.Param("id")

    // Generate URL for related resource
    postsURL := app.Router().URL("user.posts", map[string]string{
        "id": id,
    })

    ctx.JSON(200, map[string]string{
        "id":       id,
        "postsURL": postsURL,
    })
})
```

### Redirects

```go
app.Post("/users", func(w http.ResponseWriter, r *http.Request) {
    // Create user
    user := createUser(r)

    // Redirect to user profile
    url := app.Router().URL("user.show", map[string]string{
        "id": user.ID,
    })
    http.Redirect(w, r, url, http.StatusSeeOther)
})
```

---

## Parameters

### Single Parameter

```go
// Route definition
app.GetNamed("user.show", "/users/:id", getUserHandler)

// URL generation
url := router.URL("user.show", map[string]string{
    "id": "123",
})
// Result: /users/123
```

### Multiple Parameters

```go
// Route definition
app.GetNamed("post.comment", "/posts/:postId/comments/:commentId", handler)

// URL generation
url := router.URL("post.comment", map[string]string{
    "postId":    "456",
    "commentId": "789",
})
// Result: /posts/456/comments/789
```

### Wildcard Parameters

```go
// Route definition
app.GetNamed("files.show", "/files/*path", handler)

// URL generation
url := router.URL("files.show", map[string]string{
    "path": "docs/readme.md",
})
// Result: /files/docs/readme.md
```

### Missing Parameters

```go
// If parameter is missing, returns empty string
url := router.URL("user.show", nil)
// Result: ""

// Check for empty result
url := router.URL("user.show", map[string]string{"id": "123"})
if url == "" {
    // Handle error
    log.Println("Failed to generate URL")
}
```

---

## Query Strings

Query strings must be added manually:

```go
import "net/url"

// Base URL
baseURL := router.URL("users.index")

// Add query parameters
params := url.Values{}
params.Add("page", "2")
params.Add("limit", "20")
params.Add("sort", "name")

fullURL := baseURL + "?" + params.Encode()
// Result: /users?page=2&limit=20&sort=name
```

### Helper Function

```go
func URLWithQuery(router *router.Router, name string, params map[string]string, query url.Values) string {
    baseURL := router.URL(name, params)
    if baseURL == "" {
        return ""
    }

    if len(query) > 0 {
        return baseURL + "?" + query.Encode()
    }

    return baseURL
}

// Usage
query := url.Values{}
query.Add("page", "2")

url := URLWithQuery(app.Router(), "users.index", nil, query)
// Result: /users?page=2
```

---

## Use Cases

### API Responses (HATEOAS)

```go
type UserResponse struct {
    ID    string            `json:"id"`
    Name  string            `json:"name"`
    Links map[string]string `json:"_links"`
}

app.GetCtx("/users/:id", func(ctx *plumego.Context) {
    id := ctx.Param("id")
    user := getUser(id)

    response := UserResponse{
        ID:   user.ID,
        Name: user.Name,
        Links: map[string]string{
            "self":   app.Router().URL("user.show", map[string]string{"id": id}),
            "posts":  app.Router().URL("user.posts", map[string]string{"id": id}),
            "update": app.Router().URL("user.update", map[string]string{"id": id}),
            "delete": app.Router().URL("user.destroy", map[string]string{"id": id}),
        },
    }

    ctx.JSON(200, response)
})

// Response:
// {
//   "id": "123",
//   "name": "John Doe",
//   "_links": {
//     "self": "/users/123",
//     "posts": "/users/123/posts",
//     "update": "/users/123",
//     "delete": "/users/123"
//   }
// }
```

### Pagination Links

```go
type PaginatedResponse struct {
    Data  []User            `json:"data"`
    Links map[string]string `json:"links"`
}

app.GetCtx("/users", func(ctx *plumego.Context) {
    page := ctx.QueryParam("page", "1")
    pageNum, _ := strconv.Atoi(page)

    users := getUsersPage(pageNum)

    // Build pagination URLs
    query := url.Values{}

    var prevURL, nextURL string
    if pageNum > 1 {
        query.Set("page", strconv.Itoa(pageNum-1))
        prevURL = app.Router().URL("users.index") + "?" + query.Encode()
    }

    query.Set("page", strconv.Itoa(pageNum+1))
    nextURL = app.Router().URL("users.index") + "?" + query.Encode()

    response := PaginatedResponse{
        Data: users,
        Links: map[string]string{
            "self": app.Router().URL("users.index"),
            "prev": prevURL,
            "next": nextURL,
        },
    }

    ctx.JSON(200, response)
})
```

### Email Templates

```go
func sendWelcomeEmail(user User) {
    confirmURL := app.Router().URL("user.confirm", map[string]string{
        "token": user.ConfirmToken,
    })

    // Build full URL with domain
    fullURL := "https://example.com" + confirmURL

    body := fmt.Sprintf(`
        Welcome %s!

        Please confirm your email by clicking:
        %s
    `, user.Name, fullURL)

    sendEmail(user.Email, "Welcome", body)
}
```

### Location Headers

```go
app.PostCtx("/users", func(ctx *plumego.Context) {
    var user User
    if err := ctx.Bind(&user); err != nil {
        ctx.Error(400, err.Error())
        return
    }

    // Create user
    createdUser := createUser(user)

    // Generate Location header
    location := app.Router().URL("user.show", map[string]string{
        "id": createdUser.ID,
    })

    ctx.Header("Location", location)
    ctx.Status(201).JSON(createdUser)
})
```

---

## Complete Examples

### RESTful API with Links

```go
package main

import (
    "net/http"
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego"
)

type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type UserWithLinks struct {
    User
    Links map[string]string `json:"_links"`
}

func main() {
    app := core.New()

    // Register named routes
    app.GetNamed("users.index", "/users", listUsers)
    app.GetNamed("users.show", "/users/:id", getUser)
    app.PostNamed("users.store", "/users", createUser)
    app.PutNamed("users.update", "/users/:id", updateUser)
    app.DeleteNamed("users.destroy", "/users/:id", deleteUser)

    app.Boot()
}

func getUser(ctx *plumego.Context) {
    id := ctx.Param("id")
    user := User{ID: id, Name: "John Doe"}

    // Add hypermedia links
    response := UserWithLinks{
        User: user,
        Links: map[string]string{
            "self":   app.Router().URL("users.show", map[string]string{"id": id}),
            "update": app.Router().URL("users.update", map[string]string{"id": id}),
            "delete": app.Router().URL("users.destroy", map[string]string{"id": id}),
            "list":   app.Router().URL("users.index"),
        },
    }

    ctx.JSON(http.StatusOK, response)
}

func createUser(ctx *plumego.Context) {
    var user User
    if err := ctx.Bind(&user); err != nil {
        ctx.Error(http.StatusBadRequest, err.Error())
        return
    }

    // Create user (assume ID is generated)
    user.ID = "123"

    // Set Location header
    location := app.Router().URL("users.show", map[string]string{
        "id": user.ID,
    })
    ctx.Header("Location", location)

    ctx.Status(http.StatusCreated).JSON(user)
}
```

### Sitemap Generator

```go
package main

import (
    "encoding/xml"
    "net/http"
    "github.com/spcent/plumego/core"
)

type URLSet struct {
    XMLName xml.Name `xml:"urlset"`
    URLs    []URL    `xml:"url"`
}

type URL struct {
    Loc     string `xml:"loc"`
    Lastmod string `xml:"lastmod,omitempty"`
}

func main() {
    app := core.New()

    // Named routes
    app.GetNamed("home", "/", homeHandler)
    app.GetNamed("about", "/about", aboutHandler)
    app.GetNamed("blog.index", "/blog", blogIndexHandler)
    app.GetNamed("blog.show", "/blog/:slug", blogShowHandler)

    // Sitemap endpoint
    app.Get("/sitemap.xml", func(w http.ResponseWriter, r *http.Request) {
        router := app.Router()
        baseURL := "https://example.com"

        urlset := URLSet{
            URLs: []URL{
                {Loc: baseURL + router.URL("home")},
                {Loc: baseURL + router.URL("about")},
                {Loc: baseURL + router.URL("blog.index")},
            },
        }

        // Add blog posts
        posts := getBlogPosts()
        for _, post := range posts {
            url := router.URL("blog.show", map[string]string{
                "slug": post.Slug,
            })
            urlset.URLs = append(urlset.URLs, URL{
                Loc:     baseURL + url,
                Lastmod: post.UpdatedAt.Format("2006-01-02"),
            })
        }

        w.Header().Set("Content-Type", "application/xml")
        xml.NewEncoder(w).Encode(urlset)
    })

    app.Boot()
}
```

---

## Best Practices

### ✅ Do

1. **Use Named Routes for All Important URLs**
   ```go
   app.GetNamed("user.show", "/users/:id", handler)
   app.GetNamed("user.posts", "/users/:id/posts", handler)
   ```

2. **Follow Consistent Naming Conventions**
   ```go
   // ✅ Consistent
   app.GetNamed("users.index", "/users", handler)
   app.GetNamed("users.show", "/users/:id", handler)
   app.GetNamed("posts.index", "/posts", handler)
   app.GetNamed("posts.show", "/posts/:id", handler)
   ```

3. **Generate URLs Instead of Hardcoding**
   ```go
   // ✅ Generated
   url := router.URL("user.show", map[string]string{"id": id})

   // ❌ Hardcoded
   url := fmt.Sprintf("/users/%s", id)
   ```

4. **Check for Empty Results**
   ```go
   url := router.URL("route.name", params)
   if url == "" {
       log.Println("Failed to generate URL")
       return
   }
   ```

### ❌ Don't

1. **Don't Mix Naming Styles**
   ```go
   // ❌ Inconsistent
   app.GetNamed("userIndex", "/users", handler)
   app.GetNamed("user.show", "/users/:id", handler)
   app.GetNamed("PostsList", "/posts", handler)
   ```

2. **Don't Forget to Name Important Routes**
   ```go
   // ❌ Not named
   app.Get("/users/:id", handler)

   // ✅ Named
   app.GetNamed("user.show", "/users/:id", handler)
   ```

3. **Don't Hardcode URLs in Templates**
   ```go
   // ❌ Hardcoded
   <a href="/users/{{.ID}}">Profile</a>

   // ✅ Generated
   <a href="{{url "user.show" .ID}}">Profile</a>
   ```

---

## Testing

```go
package main

import (
    "testing"
    "github.com/spcent/plumego/core"
)

func TestReverseRouting(t *testing.T) {
    app := core.New()

    // Register routes
    app.GetNamed("home", "/", handler)
    app.GetNamed("user.show", "/users/:id", handler)
    app.GetNamed("post.comment", "/posts/:postId/comments/:commentId", handler)

    router := app.Router()

    tests := []struct {
        name   string
        params map[string]string
        want   string
    }{
        {
            name:   "home",
            params: nil,
            want:   "/",
        },
        {
            name:   "user.show",
            params: map[string]string{"id": "123"},
            want:   "/users/123",
        },
        {
            name: "post.comment",
            params: map[string]string{
                "postId":    "456",
                "commentId": "789",
            },
            want: "/posts/456/comments/789",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := router.URL(tt.name, tt.params)
            if got != tt.want {
                t.Errorf("URL(%q, %v) = %q, want %q", tt.name, tt.params, got, tt.want)
            }
        })
    }
}
```

---

## Next Steps

- **[Middleware Binding](middleware-binding.md)** - Per-route middleware
- **[Route Groups](route-groups.md)** - Hierarchical organization
- **[Path Parameters](path-parameters.md)** - Parameter extraction

---

**Related**:
- [Router Overview](README.md)
- [Basic Routing](basic-routing.md)
- [Contract Module](../contract/)
