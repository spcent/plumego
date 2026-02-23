# Contract Module

> **Package**: `github.com/spcent/plumego/contract`
> **Stability**: High - Core API is stable
> **Go Version**: 1.24+

The `contract` package provides request/response abstractions, error handling, and protocol adapters. It defines the contract between HTTP handlers and the application.

---

## Overview

The contract module provides:

- **Context**: Enhanced request context with helpers
- **Error Handling**: Structured error types with categories
- **Response Helpers**: JSON, XML, Stream, File responses
- **Request Binding**: Parse and validate request data
- **Protocol Adapters**: HTTP, gRPC, GraphQL support

---

## Quick Start

### Using Context

```go
import "github.com/spcent/plumego"

app := core.New()

app.GetCtx("/users/:id", func(ctx *plumego.Context) {
    id := ctx.Param("id")

    user, err := getUser(id)
    if err != nil {
        ctx.Error(404, "User not found")
        return
    }

    ctx.JSON(200, user)
})
```

### Error Handling

```go
import "github.com/spcent/plumego/contract"

func handler(w http.ResponseWriter, r *http.Request) {
    err := contract.NewValidationError("email", "invalid format")
    contract.WriteError(w, r, err)
}
```

### Response Helpers

```go
app.GetCtx("/data", func(ctx *plumego.Context) {
    // JSON response
    ctx.JSON(200, map[string]string{"status": "ok"})

    // XML response
    ctx.XML(200, data)

    // Stream response
    ctx.Stream(200, "text/event-stream", eventStream)

    // File download
    ctx.File("report.pdf")
})
```

---

## Core Concepts

### 1. Context

The `Context` struct wraps `http.ResponseWriter` and `*http.Request` with convenience methods:

```go
type Context struct {
    Writer  http.ResponseWriter
    Request *http.Request
    // ... internal fields
}
```

**Features**:
- Path parameter extraction
- Query parameter parsing
- Request body binding
- Response helpers (JSON, XML, etc.)
- Status code management

### 2. Error Types

Structured errors with categories:

```go
type Error struct {
    Status   int
    Code     string
    Message  string
    Category ErrorCategory
    Details  map[string]interface{}
}
```

**Categories**:
- Client errors (400-499)
- Server errors (500-599)
- Business logic errors
- Validation errors
- Authentication errors
- Rate limit errors

### 3. Response Helpers

Type-safe response methods:

```go
// JSON
ctx.JSON(200, data)

// XML
ctx.XML(200, data)

// Plain text
ctx.String(200, "Hello")

// HTML
ctx.HTML(200, "<h1>Hello</h1>")

// Stream
ctx.Stream(200, "text/event-stream", reader)

// File
ctx.File("path/to/file.pdf")

// Redirect
ctx.Redirect(302, "/login")
```

### 4. Request Binding

Parse request data into structs:

```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

var req CreateUserRequest
if err := ctx.Bind(&req); err != nil {
    ctx.Error(400, err.Error())
    return
}
```

---

## Documentation Structure

| Document | Description |
|----------|-------------|
| **[Context](context.md)** | Request context and helpers |
| **[Errors](errors.md)** | Error types and handling |
| **[Response](response.md)** | Response helper methods |
| **[Request](request.md)** | Request parsing and binding |
| **[Protocol Adapters](protocol/)** | HTTP, gRPC, GraphQL adapters |

---

## Usage Examples

### Basic API Handler

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

func main() {
    app := core.New()

    // List users
    app.GetCtx("/users", func(ctx *plumego.Context) {
        users := []User{
            {ID: "1", Name: "Alice"},
            {ID: "2", Name: "Bob"},
        }
        ctx.JSON(http.StatusOK, users)
    })

    // Get user
    app.GetCtx("/users/:id", func(ctx *plumego.Context) {
        id := ctx.Param("id")
        user := User{ID: id, Name: "Alice"}
        ctx.JSON(http.StatusOK, user)
    })

    // Create user
    app.PostCtx("/users", func(ctx *plumego.Context) {
        var user User
        if err := ctx.Bind(&user); err != nil {
            ctx.Error(http.StatusBadRequest, err.Error())
            return
        }

        // Save user...
        ctx.Status(http.StatusCreated).JSON(user)
    })

    app.Boot()
}
```

### Error Handling

```go
import "github.com/spcent/plumego/contract"

app.GetCtx("/users/:id", func(ctx *plumego.Context) {
    id := ctx.Param("id")

    user, err := getUserFromDB(id)
    if err != nil {
        // Structured error
        cerr := contract.NewError(
            contract.WithStatus(http.StatusNotFound),
            contract.WithCode("USER_NOT_FOUND"),
            contract.WithMessage("User not found"),
            contract.WithCategory(contract.CategoryClient),
            contract.WithDetails(map[string]interface{}{
                "userId": id,
            }),
        )

        contract.WriteError(ctx.Writer, ctx.Request, cerr)
        return
    }

    ctx.JSON(http.StatusOK, user)
})
```

### File Upload

```go
app.PostCtx("/upload", func(ctx *plumego.Context) {
    file, header, err := ctx.Request.FormFile("file")
    if err != nil {
        ctx.Error(http.StatusBadRequest, "No file uploaded")
        return
    }
    defer file.Close()

    // Save file...
    filename := header.Filename
    dst, err := os.Create("./uploads/" + filename)
    if err != nil {
        ctx.Error(http.StatusInternalServerError, "Failed to save file")
        return
    }
    defer dst.Close()

    io.Copy(dst, file)

    ctx.JSON(http.StatusOK, map[string]string{
        "filename": filename,
        "size":     fmt.Sprintf("%d", header.Size),
    })
})
```

### SSE Streaming

```go
app.GetCtx("/events", func(ctx *plumego.Context) {
    ctx.Writer.Header().Set("Content-Type", "text/event-stream")
    ctx.Writer.Header().Set("Cache-Control", "no-cache")
    ctx.Writer.Header().Set("Connection", "keep-alive")

    flusher, ok := ctx.Writer.(http.Flusher)
    if !ok {
        ctx.Error(http.StatusInternalServerError, "Streaming not supported")
        return
    }

    // Send events
    for i := 0; i < 10; i++ {
        fmt.Fprintf(ctx.Writer, "data: Event %d\n\n", i)
        flusher.Flush()
        time.Sleep(1 * time.Second)
    }
})
```

---

## Context Methods

### Parameter Methods

```go
// Path parameters
id := ctx.Param("id")
params := ctx.Params() // map[string]string

// Query parameters
page := ctx.Query("page")              // Returns string
pageNum := ctx.QueryInt("page", 1)     // Returns int with default
enabled := ctx.QueryBool("enabled", false) // Returns bool with default

// Headers
token := ctx.GetHeader("Authorization")
ctx.SetHeader("X-Custom", "value")
```

### Request Methods

```go
// Bind request body
var req CreateUserRequest
if err := ctx.Bind(&req); err != nil {
    // Handle error
}

// Access raw request
method := ctx.Request.Method
path := ctx.Request.URL.Path
body := ctx.Request.Body
```

### Response Methods

```go
// Set status
ctx.Status(http.StatusCreated)

// JSON
ctx.JSON(200, data)

// XML
ctx.XML(200, data)

// String
ctx.String(200, "Hello")

// HTML
ctx.HTML(200, "<h1>Hello</h1>")

// Stream
ctx.Stream(200, contentType, reader)

// File
ctx.File("path/to/file")

// Redirect
ctx.Redirect(302, "/login")

// No content
ctx.NoContent(204)
```

---

## Error Categories

```go
const (
    CategoryClient         ErrorCategory = "client"         // 4xx errors
    CategoryServer         ErrorCategory = "server"         // 5xx errors
    CategoryBusiness       ErrorCategory = "business"       // Business logic
    CategoryValidation     ErrorCategory = "validation"     // Input validation
    CategoryAuthentication ErrorCategory = "authentication" // Auth failures
    CategoryAuthorization  ErrorCategory = "authorization"  // Permission denied
    CategoryRateLimit      ErrorCategory = "rate_limit"     // Rate limiting
    CategoryTimeout        ErrorCategory = "timeout"        // Timeouts
)
```

### Error Creation

```go
// Validation error
err := contract.NewValidationError("email", "invalid format")

// Not found error
err := contract.NewNotFoundError("User")

// Custom error
err := contract.NewError(
    contract.WithStatus(http.StatusBadRequest),
    contract.WithCode("INVALID_REQUEST"),
    contract.WithMessage("Request validation failed"),
    contract.WithCategory(contract.CategoryValidation),
    contract.WithDetails(map[string]interface{}{
        "field": "email",
        "issue": "invalid format",
    }),
)
```

---

## Protocol Adapters

### HTTP Adapter (Default)

```go
// Standard HTTP handler
app.Get("/users", func(w http.ResponseWriter, r *http.Request) {
    // Handle HTTP request
})

// Context-aware handler
app.GetCtx("/users", func(ctx *plumego.Context) {
    // Use context helpers
})
```

### gRPC Adapter

```go
import "github.com/spcent/plumego/contract/protocol"

// Convert gRPC request to Plumego context
grpcAdapter := protocol.NewGRPCAdapter()
ctx := grpcAdapter.ToContext(grpcRequest)

// Use standard handlers
handler(ctx)
```

### GraphQL Adapter

```go
import "github.com/spcent/plumego/contract/protocol"

// Convert GraphQL request to Plumego context
gqlAdapter := protocol.NewGraphQLAdapter()
ctx := gqlAdapter.ToContext(gqlRequest)

// Use standard handlers
handler(ctx)
```

---

## Best Practices

### ✅ Do

1. **Use Context Helpers**
   ```go
   // ✅ Use helpers
   ctx.JSON(200, data)

   // ❌ Manual encoding
   json.NewEncoder(ctx.Writer).Encode(data)
   ```

2. **Use Structured Errors**
   ```go
   // ✅ Structured
   err := contract.NewValidationError("email", "invalid")
   contract.WriteError(ctx.Writer, ctx.Request, err)

   // ❌ Plain text
   http.Error(ctx.Writer, "Invalid email", 400)
   ```

3. **Validate Input**
   ```go
   // ✅ Bind and validate
   var req CreateUserRequest
   if err := ctx.Bind(&req); err != nil {
       ctx.Error(400, err.Error())
       return
   }
   ```

### ❌ Don't

1. **Don't Write After Response Sent**
   ```go
   // ❌ Response already sent
   ctx.JSON(200, data)
   ctx.String(200, "extra") // Error!

   // ✅ One response per request
   ctx.JSON(200, data)
   ```

2. **Don't Ignore Errors**
   ```go
   // ❌ Ignoring error
   ctx.Bind(&req)

   // ✅ Handle error
   if err := ctx.Bind(&req); err != nil {
       ctx.Error(400, err.Error())
       return
   }
   ```

---

## Performance Considerations

1. **JSON Encoding**: Use `ctx.JSON()` for automatic content-type and encoding
2. **Streaming**: Use `ctx.Stream()` for large responses
3. **Error Allocation**: Reuse error objects when possible
4. **Context Reuse**: Don't store context beyond request lifetime

---

## Next Steps

- **[Context](context.md)** - Request context details
- **[Errors](errors.md)** - Error handling patterns
- **[Response](response.md)** - Response helpers
- **[Request](request.md)** - Request parsing
- **[Protocol Adapters](protocol/)** - Multi-protocol support

---

## Related Modules

- **[Core](../core/)** - Application and lifecycle
- **[Router](../router/)** - HTTP routing
- **[Middleware](../middleware/)** - Request processing
- **[Validator](../validator/)** - Input validation

---

**Stability**: High - Breaking changes require major version bump
**Maintainers**: Plumego Core Team
**Last Updated**: 2026-02-11
