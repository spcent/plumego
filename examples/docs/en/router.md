# Router module

The **router** package provides trie-based matching with groups, parameters, and middleware inheritance while staying compatible with standard `http.Handler` interfaces.

## Handler shapes
- Standard handlers: `Get`, `Post`, `Put`, `Patch`, `Delete`, `Options`, `Head`, `Any` accept `http.Handler` values.
- Function helpers: `GetFunc`/`PostFunc`/etc accept `http.HandlerFunc`.
- Context-aware handlers: `GetCtx`/`PostCtx`/etc accept `contract.CtxHandlerFunc` where path params and request metadata are already parsed.

## Path patterns
- Static: `/docs/index.html`
- Parameters: `/users/:id`, `/teams/:teamID/members/:id`
- Wildcards: `/*filepath` (capture the remainder for static assets or docs)

## Grouping and middleware order
Middleware runs **global → group → route-specific**.

```go
api := app.Router().Group("/api", middleware.SimpleAuth("token"))
api.Use(middleware.Timeout(2 * time.Second))

api.GetCtx("/users/:id", func(ctx *contract.Ctx) {
    id := ctx.Param("id")
    ctx.Text(http.StatusOK, "user="+id)
})

// Nested groups inherit middleware and prefixes.
v1 := api.Group("/v1")
v1.Get("/ping", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("pong")) })
```

Register everything before `app.Boot()`; the router is frozen during boot to prevent missing registrations.

## Static frontends and catch-alls
Mount static assets under their own group so cache headers or auth can be isolated.

```go
assets := app.Router().Group("/docs")
assets.Get("/*filepath", func(w http.ResponseWriter, r *http.Request) {
    http.ServeFileFS(w, r, os.DirFS("./examples/docs"), path.Clean(r.URL.Path))
})
```

For embedded bundles, use the helpers in `frontend`:

```go
// Mount an embedded SPA or docs site at "/".
_ = frontend.RegisterFS(app.Router(), http.FS(staticFS), frontend.WithPrefix("/"))
```

## Debugging tools
- Enable `core.WithDebug()` to log every method/path pair during boot.
- Dump the routing table from `router.Routes()` in a development-only endpoint when diagnosing conflicts.
- Late registrations that panic during boot are expected; fix initialization order instead of deferring registration.

## Where to look in the repo
- `router/router.go`: trie matching, groups, and handler helpers.
- `frontend/register.go`: helpers for mounting static directories or embedded frontend bundles.
- `examples/reference/main.go`: real wiring of API, metrics, health, docs, and frontend routes.
