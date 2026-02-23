# Request Context

> **Package**: `github.com/spcent/plumego`

The `Context` struct provides a convenient wrapper around `http.ResponseWriter` and `*http.Request` with helper methods for common operations.

---

## Table of Contents

- [Overview](#overview)
- [Creating Context](#creating-context)
- [Parameter Methods](#parameter-methods)
- [Query Methods](#query-methods)
- [Header Methods](#header-methods)
- [Response Methods](#response-methods)
- [Request Methods](#request-methods)
- [Best Practices](#best-practices)

---

## Overview

### What is Context?

`Context` wraps the standard library's `http.ResponseWriter` and `*http.Request` to provide:

- Simplified parameter extraction
- Type-safe query parameter parsing
- Convenient response helpers
- Request body binding
- Status code management

### Structure

```go
type Context struct {
    Writer     http.ResponseWriter
    Request    *http.Request
    StatusCode int
    // ... internal fields
}
```

---

## Creating Context

### In Handlers

Plumego automatically creates context for context-aware handlers:

```go
app.GetCtx("/users/:id", func(ctx *plumego.Context) {
    // Context is provided
    id := ctx.Param("id")
})
```

### Manual Creation

For testing or custom use:

```go
import "github.com/spcent/plumego"

w := httptest.NewRecorder()
r := httptest.NewRequest("GET", "/test", nil)

ctx := plumego.NewContext(w, r)
```

---

## Parameter Methods

### Param

Extract path parameters:

```go
// Route: /users/:id
app.GetCtx("/users/:id", func(ctx *plumego.Context) {
    id := ctx.Param("id")
    // id = "123" for request /users/123
})
```

### Params

Get all path parameters as a map:

```go
// Route: /posts/:category/:id
app.GetCtx("/posts/:category/:id", func(ctx *plumego.Context) {
    params := ctx.Params()
    // params = map[string]string{
    //   "category": "tech",
    //   "id": "456"
    // }

    category := params["category"]
    id := params["id"]
})
```

### ParamInt

Convert path parameter to integer:

```go
app.GetCtx("/pages/:page", func(ctx *plumego.Context) {
    page, err := ctx.ParamInt("page")
    if err != nil {
        ctx.Error(400, "Invalid page number")
        return
    }

    // Use page as int
})
```

---

## Query Methods

### Query

Get query parameter as string:

```go
// Request: /search?q=golang
app.GetCtx("/search", func(ctx *plumego.Context) {
    query := ctx.Query("q")
    // query = "golang"

    // With default value
    query := ctx.QueryDefault("q", "default")
})
```

### QueryInt

Get query parameter as integer:

```go
// Request: /users?page=2
app.GetCtx("/users", func(ctx *plumego.Context) {
    page := ctx.QueryInt("page", 1) // Default: 1
    // page = 2
})
```

### QueryBool

Get query parameter as boolean:

```go
// Request: /users?active=true
app.GetCtx("/users", func(ctx *plumego.Context) {
    active := ctx.QueryBool("active", false) // Default: false
    // active = true

    // Accepts: true, false, 1, 0, yes, no, on, off
})
```

### QueryFloat

Get query parameter as float:

```go
// Request: /products?price=19.99
app.GetCtx("/products", func(ctx *plumego.Context) {
    price := ctx.QueryFloat("price", 0.0) // Default: 0.0
    // price = 19.99
})
```

### QueryArray

Get query parameter as array:

```go
// Request: /filter?tags=go&tags=web&tags=api
app.GetCtx("/filter", func(ctx *plumego.Context) {
    tags := ctx.QueryArray("tags")
    // tags = []string{"go", "web", "api"}
})
```

### AllQueries

Get all query parameters:

```go
// Request: /search?q=golang&page=2&limit=10
app.GetCtx("/search", func(ctx *plumego.Context) {
    queries := ctx.AllQueries()
    // queries = url.Values{
    //   "q":     []string{"golang"},
    //   "page":  []string{"2"},
    //   "limit": []string{"10"},
    // }
})
```

---

## Header Methods

### GetHeader

Get request header:

```go
app.GetCtx("/api/users", func(ctx *plumego.Context) {
    token := ctx.GetHeader("Authorization")
    contentType := ctx.GetHeader("Content-Type")
    userAgent := ctx.GetHeader("User-Agent")
})
```

### SetHeader

Set response header:

```go
app.GetCtx("/data", func(ctx *plumego.Context) {
    ctx.SetHeader("X-Custom-Header", "value")
    ctx.SetHeader("Cache-Control", "no-cache")

    ctx.JSON(200, data)
})
```

### SetCookie

Set response cookie:

```go
app.GetCtx("/login", func(ctx *plumego.Context) {
    cookie := &http.Cookie{
        Name:     "session",
        Value:    "abc123",
        Path:     "/",
        MaxAge:   3600,
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
    }

    ctx.SetCookie(cookie)
})
```

### GetCookie

Get request cookie:

```go
app.GetCtx("/profile", func(ctx *plumego.Context) {
    cookie, err := ctx.GetCookie("session")
    if err != nil {
        ctx.Error(401, "Not authenticated")
        return
    }

    sessionID := cookie.Value
})
```

---

## Response Methods

### Status

Set response status code:

```go
app.PostCtx("/users", func(ctx *plumego.Context) {
    // Create user...
    ctx.Status(http.StatusCreated).JSON(user)
})
```

### JSON

Send JSON response:

```go
app.GetCtx("/users", func(ctx *plumego.Context) {
    users := []User{
        {ID: "1", Name: "Alice"},
        {ID: "2", Name: "Bob"},
    }

    ctx.JSON(http.StatusOK, users)
})

// With status chaining
ctx.Status(201).JSON(user)
```

### XML

Send XML response:

```go
type Response struct {
    XMLName xml.Name `xml:"response"`
    Status  string   `xml:"status"`
    Message string   `xml:"message"`
}

app.GetCtx("/data", func(ctx *plumego.Context) {
    resp := Response{
        Status:  "ok",
        Message: "Success",
    }

    ctx.XML(http.StatusOK, resp)
})
```

### String

Send plain text response:

```go
app.GetCtx("/ping", func(ctx *plumego.Context) {
    ctx.String(http.StatusOK, "pong")
})

// With formatting
ctx.String(200, "User ID: %s", userID)
```

### HTML

Send HTML response:

```go
app.GetCtx("/", func(ctx *plumego.Context) {
    html := "<h1>Welcome</h1><p>Hello World</p>"
    ctx.HTML(http.StatusOK, html)
})
```

### Stream

Send streaming response:

```go
app.GetCtx("/events", func(ctx *plumego.Context) {
    events := make(chan string)

    // Start event producer
    go produceEvents(events)

    // Stream events
    ctx.Stream(http.StatusOK, "text/event-stream", func(w io.Writer) error {
        for event := range events {
            if _, err := fmt.Fprintf(w, "data: %s\n\n", event); err != nil {
                return err
            }

            if f, ok := w.(http.Flusher); ok {
                f.Flush()
            }
        }
        return nil
    })
})
```

### File

Send file response:

```go
app.GetCtx("/download/:filename", func(ctx *plumego.Context) {
    filename := ctx.Param("filename")
    path := "./uploads/" + filename

    // Check file exists
    if _, err := os.Stat(path); os.IsNotExist(err) {
        ctx.Error(404, "File not found")
        return
    }

    ctx.File(path)
})
```

### Attachment

Send file as downloadable attachment:

```go
app.GetCtx("/reports/:id", func(ctx *plumego.Context) {
    id := ctx.Param("id")
    path := "./reports/" + id + ".pdf"

    ctx.Attachment(path, "report.pdf")
})
```

### Data

Send raw bytes:

```go
app.GetCtx("/image", func(ctx *plumego.Context) {
    imageData := loadImage()

    ctx.Data(http.StatusOK, "image/png", imageData)
})
```

### Redirect

Redirect to another URL:

```go
app.GetCtx("/old-path", func(ctx *plumego.Context) {
    ctx.Redirect(http.StatusMovedPermanently, "/new-path")
})

// Temporary redirect
ctx.Redirect(http.StatusFound, "/login")
```

### NoContent

Send 204 No Content:

```go
app.DeleteCtx("/users/:id", func(ctx *plumego.Context) {
    id := ctx.Param("id")
    deleteUser(id)

    ctx.NoContent(http.StatusNoContent)
})
```

### Error

Send error response:

```go
app.GetCtx("/users/:id", func(ctx *plumego.Context) {
    id := ctx.Param("id")

    user, err := getUser(id)
    if err != nil {
        ctx.Error(http.StatusNotFound, "User not found")
        return
    }

    ctx.JSON(http.StatusOK, user)
})
```

---

## Request Methods

### Bind

Parse request body into struct:

```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"gte=0,lte=120"`
}

app.PostCtx("/users", func(ctx *plumego.Context) {
    var req CreateUserRequest

    if err := ctx.Bind(&req); err != nil {
        ctx.Error(http.StatusBadRequest, err.Error())
        return
    }

    // Use req...
    user := createUser(req)
    ctx.Status(http.StatusCreated).JSON(user)
})
```

### BindJSON

Parse JSON body (explicit):

```go
app.PostCtx("/users", func(ctx *plumego.Context) {
    var req CreateUserRequest

    if err := ctx.BindJSON(&req); err != nil {
        ctx.Error(400, "Invalid JSON")
        return
    }

    // Use req...
})
```

### BindXML

Parse XML body:

```go
app.PostCtx("/data", func(ctx *plumego.Context) {
    var req XMLRequest

    if err := ctx.BindXML(&req); err != nil {
        ctx.Error(400, "Invalid XML")
        return
    }

    // Use req...
})
```

### BindForm

Parse form data:

```go
app.PostCtx("/submit", func(ctx *plumego.Context) {
    var form ContactForm

    if err := ctx.BindForm(&form); err != nil {
        ctx.Error(400, "Invalid form data")
        return
    }

    // Use form...
})
```

### Body

Access raw request body:

```go
app.PostCtx("/webhook", func(ctx *plumego.Context) {
    body, err := io.ReadAll(ctx.Request.Body)
    if err != nil {
        ctx.Error(400, "Failed to read body")
        return
    }

    // Process raw body...
})
```

---

## Complete Examples

### CRUD API

```go
package main

import (
    "net/http"
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego"
)

type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

var users = make(map[string]User)

func main() {
    app := core.New()

    // List users with pagination
    app.GetCtx("/users", func(ctx *plumego.Context) {
        page := ctx.QueryInt("page", 1)
        limit := ctx.QueryInt("limit", 10)

        // Apply pagination...
        ctx.JSON(http.StatusOK, users)
    })

    // Get user
    app.GetCtx("/users/:id", func(ctx *plumego.Context) {
        id := ctx.Param("id")

        user, exists := users[id]
        if !exists {
            ctx.Error(http.StatusNotFound, "User not found")
            return
        }

        ctx.JSON(http.StatusOK, user)
    })

    // Create user
    app.PostCtx("/users", func(ctx *plumego.Context) {
        var user User

        if err := ctx.Bind(&user); err != nil {
            ctx.Error(http.StatusBadRequest, err.Error())
            return
        }

        // Generate ID
        user.ID = generateID()
        users[user.ID] = user

        ctx.Status(http.StatusCreated).JSON(user)
    })

    // Update user
    app.PutCtx("/users/:id", func(ctx *plumego.Context) {
        id := ctx.Param("id")

        if _, exists := users[id]; !exists {
            ctx.Error(http.StatusNotFound, "User not found")
            return
        }

        var user User
        if err := ctx.Bind(&user); err != nil {
            ctx.Error(http.StatusBadRequest, err.Error())
            return
        }

        user.ID = id
        users[id] = user

        ctx.JSON(http.StatusOK, user)
    })

    // Delete user
    app.DeleteCtx("/users/:id", func(ctx *plumego.Context) {
        id := ctx.Param("id")

        if _, exists := users[id]; !exists {
            ctx.Error(http.StatusNotFound, "User not found")
            return
        }

        delete(users, id)
        ctx.NoContent(http.StatusNoContent)
    })

    app.Boot()
}
```

### File Upload and Download

```go
package main

import (
    "io"
    "os"
    "path/filepath"
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego"
)

func main() {
    app := core.New()

    // Upload file
    app.PostCtx("/upload", func(ctx *plumego.Context) {
        file, header, err := ctx.Request.FormFile("file")
        if err != nil {
            ctx.Error(400, "No file uploaded")
            return
        }
        defer file.Close()

        // Validate file
        if header.Size > 10<<20 { // 10 MB
            ctx.Error(400, "File too large")
            return
        }

        // Save file
        filename := filepath.Base(header.Filename)
        dst, err := os.Create("./uploads/" + filename)
        if err != nil {
            ctx.Error(500, "Failed to save file")
            return
        }
        defer dst.Close()

        if _, err := io.Copy(dst, file); err != nil {
            ctx.Error(500, "Failed to save file")
            return
        }

        ctx.JSON(200, map[string]string{
            "filename": filename,
            "size":     fmt.Sprintf("%d", header.Size),
        })
    })

    // Download file
    app.GetCtx("/download/:filename", func(ctx *plumego.Context) {
        filename := ctx.Param("filename")
        path := "./uploads/" + filename

        if _, err := os.Stat(path); os.IsNotExist(err) {
            ctx.Error(404, "File not found")
            return
        }

        ctx.Attachment(path, filename)
    })

    app.Boot()
}
```

---

## Best Practices

### ✅ Do

1. **Use Type-Safe Methods**
   ```go
   // ✅ Type-safe
   page := ctx.QueryInt("page", 1)

   // ❌ Manual conversion
   pageStr := ctx.Query("page")
   page, _ := strconv.Atoi(pageStr)
   ```

2. **Check Errors**
   ```go
   // ✅ Check errors
   if err := ctx.Bind(&req); err != nil {
       ctx.Error(400, err.Error())
       return
   }

   // ❌ Ignore errors
   ctx.Bind(&req)
   ```

3. **Use Response Helpers**
   ```go
   // ✅ Use helpers
   ctx.JSON(200, data)

   // ❌ Manual encoding
   w.Header().Set("Content-Type", "application/json")
   json.NewEncoder(w).Encode(data)
   ```

### ❌ Don't

1. **Don't Write Multiple Responses**
   ```go
   // ❌ Multiple writes
   ctx.JSON(200, data1)
   ctx.JSON(200, data2) // Error!

   // ✅ One response
   ctx.JSON(200, data)
   ```

2. **Don't Store Context**
   ```go
   // ❌ Storing context
   var globalCtx *plumego.Context
   app.GetCtx("/", func(ctx *plumego.Context) {
       globalCtx = ctx // Don't!
   })

   // ✅ Use within handler only
   app.GetCtx("/", func(ctx *plumego.Context) {
       // Use ctx here
   })
   ```

---

## Next Steps

- **[Errors](errors.md)** - Error handling
- **[Response](response.md)** - Response helpers details
- **[Request](request.md)** - Request parsing details

---

**Related**:
- [Contract Overview](README.md)
- [Router Module](../router/)
- [Validator Module](../validator/)
