# HTTP Adapter

> **Package**: `github.com/spcent/plumego`

The HTTP adapter is the default protocol for Plumego and requires no explicit configuration.

---

## Overview

HTTP is the native protocol for Plumego. All features work out-of-the-box:

- ✅ Standard `http.Handler` interface
- ✅ Context-aware handlers
- ✅ All HTTP methods (GET, POST, PUT, DELETE, etc.)
- ✅ Path parameters and query strings
- ✅ Request/response helpers
- ✅ Middleware support

---

## Handler Types

### 1. Standard Library Handler

```go
app.Get("/users", func(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(users)
})
```

### 2. Context-Aware Handler

```go
app.GetCtx("/users", func(ctx *plumego.Context) {
    ctx.JSON(200, users)
})
```

### 3. http.Handler Interface

```go
type UserHandler struct {}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Handle request
}

app.Get("/users", &UserHandler{})
```

---

## Request Handling

### Path Parameters

```go
app.GetCtx("/users/:id", func(ctx *plumego.Context) {
    id := ctx.Param("id")
    ctx.JSON(200, getUser(id))
})
```

### Query Parameters

```go
app.GetCtx("/search", func(ctx *plumego.Context) {
    q := ctx.Query("q")
    page := ctx.QueryInt("page", 1)
    ctx.JSON(200, search(q, page))
})
```

### Headers

```go
app.GetCtx("/api/users", func(ctx *plumego.Context) {
    token := ctx.GetHeader("Authorization")
    ctx.JSON(200, users)
})
```

### Body

```go
app.PostCtx("/users", func(ctx *plumego.Context) {
    var user User
    if err := ctx.Bind(&user); err != nil {
        ctx.Error(400, err.Error())
        return
    }
    ctx.JSON(201, createUser(user))
})
```

---

## Response Handling

### JSON

```go
ctx.JSON(200, data)
```

### XML

```go
ctx.XML(200, data)
```

### Plain Text

```go
ctx.String(200, "Hello World")
```

### HTML

```go
ctx.HTML(200, "<h1>Hello</h1>")
```

### File

```go
ctx.File("path/to/file.pdf")
```

### Stream

```go
ctx.Stream(200, "text/event-stream", eventStream)
```

---

## Examples

See [Context](../context.md), [Request](../request.md), and [Response](../response.md) documentation for complete HTTP adapter usage.

---

**Next**: [gRPC/GraphQL Adapters](grpc-graphql.md)
