# Frontend Module

> **Package Path**: `github.com/spcent/plumego/x/frontend` | **Stability**: Medium | **Priority**: P2

## Overview

`frontend/` serves static assets and SPA bundles through the Plumego router.

Current registration APIs are explicit:

- `frontend.NewMountFromDir(dir, opts...)`
- `frontend.NewMountFS(fsys, opts...)`
- `frontend.NewMountEmbedded(opts...)`
- `frontend.RegisterFromDir(router, dir, opts...)`
- `frontend.RegisterFS(router, fsys, opts...)`
- `frontend.RegisterEmbedded(router, opts...)`

There is no implicit mount through `core`; you register the frontend bundle directly on the router or wrap it as a `core` component with `core.NewFrontendComponentFromDir(...)` / `core.NewFrontendComponentFromFS(...)`.

## Quick Start

### Serve a bundle from disk

```go
app := core.New(core.WithAddr(":8080"))

mount, err := frontend.NewMountFromDir("./dist",
    frontend.WithPrefix("/"),
    frontend.WithFallback(true),
    frontend.WithCacheControl("public, max-age=3600"),
)
if err != nil {
    log.Fatal(err)
}
if err := mount.Register(app.Router()); err != nil {
    log.Fatal(err)
}

if err := app.Run(context.Background()); err != nil {
    log.Fatal(err)
}
```

### Serve embedded assets

```go
mount, err := frontend.NewMountEmbedded(
    frontend.WithPrefix("/"),
    frontend.WithFallback(true),
    frontend.WithIndexCacheControl("no-cache, no-store, must-revalidate"),
)
if err != nil {
    log.Fatal(err)
}
if err := mount.Register(app.Router()); err != nil {
    log.Fatal(err)
}
```

## Options

- `WithPrefix("/app")` mounts under a sub-path.
- `WithFallback(true)` enables SPA fallback to the index file.
- `WithCacheControl(...)` sets asset cache headers.
- `WithIndexCacheControl(...)` sets index response cache headers.
- `WithPrecompressed(true)` serves `.gz` / `.br` files when available.
- `WithHeaders(...)` applies extra response headers.
- `WithNotFoundPage(...)` sets a custom 404 page.
- `WithErrorPage(...)` sets a custom 5xx page.
- `WithMIMETypes(...)` overrides MIME mappings.

## Related Documentation

- [Static Files](static-files.md)
- [Embedded Assets](embedded-assets.md)
- [Core Module](../core/README.md)
