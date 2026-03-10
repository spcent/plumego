# Request Handling (`contract.Ctx`)

> Package: `github.com/spcent/plumego/contract`

This page documents canonical request handling with `contract.Ctx` in Plumego v1.

## Canonical Pattern

`core.App` route methods keep the standard handler shape.
Use `contract.AdaptCtxHandler` when you want `*contract.Ctx` helpers:

```go
app.Post("/users", contract.AdaptCtxHandler(func(ctx *contract.Ctx) {
    var req CreateUserRequest
    if err := ctx.BindAndValidateJSONWithOptions(&req, contract.BindOptions{
        DisallowUnknownFields: true,
        Logger:                ctx.Logger,
    }); err != nil {
        contract.WriteBindError(ctx.W, ctx.R, err)
        return
    }

    _ = ctx.Response(http.StatusCreated, map[string]any{"id": "u_123"}, nil)
}, app.Logger()).ServeHTTP)
```

## Reading Request Data

### Path params

```go
app.Get("/users/:id", contract.AdaptCtxHandler(func(ctx *contract.Ctx) {
    id, ok := ctx.Param("id")
    if !ok {
        contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
            Status(http.StatusBadRequest).
            Code(contract.CodeValidationError).
            Message("missing id").
            Build())
        return
    }
    _ = ctx.Response(http.StatusOK, map[string]any{"id": id}, nil)
}, app.Logger()).ServeHTTP)
```

### Query and headers

```go
q := ctx.Query.Get("q")
lang := ctx.Headers.Get("Accept-Language")
```

### Raw body

```go
body, err := ctx.ReadBody()
if err != nil {
    contract.WriteBindError(ctx.W, ctx.R, err)
    return
}
_ = body
```

## JSON Binding

### Basic bind

```go
var req CreateUserRequest
if err := ctx.BindJSON(&req); err != nil {
    contract.WriteBindError(ctx.W, ctx.R, err)
    return
}
```

### Bind + validate

```go
if err := ctx.BindAndValidateJSONWithOptions(&req, contract.BindOptions{
    DisallowUnknownFields: true,
    MaxBodyBytes:          1 << 20,
    Logger:                ctx.Logger,
}); err != nil {
    contract.WriteBindError(ctx.W, ctx.R, err)
    return
}
```

## Form and file upload

`contract.Ctx` exposes standard request objects (`ctx.R`) so multipart/form parsing follows standard-library behavior.
Use `ctx.R.ParseMultipartForm(...)`, then read `ctx.R.Form` / `ctx.R.MultipartForm` as needed.

## Error Handling

Use one structured error path:

```go
contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
    Status(http.StatusUnauthorized).
    Category(contract.CategoryAuthentication).
    Code(contract.CodeAuthUnauthenticated).
    Message("invalid token").
    Build())
```

For bind/validation errors, prefer `contract.WriteBindError(...)`.

## Response Helpers

- `ctx.Response(status, data, meta)`
- `ctx.JSON(status, data)`
- `ctx.Text(status, text)`
- `ctx.StreamNDJSON(items)` / `ctx.StreamSSE(...)` for streaming

## Testing

Use `httptest.NewRecorder` + `httptest.NewRequest`, then call adapted handler:

```go
h := contract.AdaptCtxHandler(func(ctx *contract.Ctx) {
    _ = ctx.JSON(http.StatusOK, map[string]any{"ok": true})
}, log.NewGLogger())

rr := httptest.NewRecorder()
req := httptest.NewRequest(http.MethodGet, "/ping", nil)
h.ServeHTTP(rr, req)
```
