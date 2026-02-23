# Frontend Module

> **Package Path**: `github.com/spcent/plumego/frontend` | **Stability**: Medium | **Priority**: P2

## Overview

The `frontend/` package serves static files for Single Page Applications (SPAs). It supports serving from the local filesystem (development) or embedded assets compiled into the binary (production).

**Key Features**:
- **SPA Routing**: Fallback to `index.html` for client-side routing
- **Go Embed Support**: Bundle frontend assets into the binary
- **Pre-compressed Files**: Serve `.gz`/`.br` files when supported
- **Custom Headers**: Cache-Control, security headers per route
- **Custom MIME Types**: Override default content-type detection
- **Custom Error Pages**: 404 and 5xx error page support

## Quick Start

### Serve from Disk (Development)

```go
import (
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/frontend"
)

app := core.New(core.WithAddr(":8080"))

// Mount SPA at root
err := frontend.Register(app.Router(), "./dist",
    frontend.WithPrefix("/"),
    frontend.WithFallback(true),  // SPA mode: unknown paths → index.html
    frontend.WithCacheControl("public, max-age=3600"),
)

app.Boot()
```

**File structure**:
```
dist/
├── index.html
├── assets/
│   ├── main.js
│   └── main.css
└── images/
    └── logo.svg
```

### Embedded Assets (Production)

```go
import (
    "embed"
    "github.com/spcent/plumego/frontend"
)

//go:embed dist/*
var distFS embed.FS

err := frontend.RegisterFS(app.Router(), http.FS(distFS),
    frontend.WithPrefix("/"),
    frontend.WithFallback(true),
    frontend.WithCacheControl("public, max-age=86400, immutable"),
)
```

## Configuration

```go
frontend.Register(router, "./dist",
    // URL prefix where frontend is mounted
    frontend.WithPrefix("/app"),

    // SPA mode: route unknown paths to index.html
    frontend.WithFallback(true),

    // Cache-Control for assets
    frontend.WithCacheControl("public, max-age=31536000, immutable"),

    // Separate cache policy for index.html
    frontend.WithIndexCacheControl("no-cache, no-store, must-revalidate"),

    // Serve pre-compressed files (.gz, .br)
    frontend.WithPrecompressed(true),

    // Custom 404 page
    frontend.WithNotFoundPage("dist/404.html"),

    // Additional response headers
    frontend.WithHeaders(map[string]string{
        "X-Content-Type-Options": "nosniff",
        "X-Frame-Options":        "DENY",
    }),

    // Override MIME types
    frontend.WithMIMETypes(map[string]string{
        ".wasm": "application/wasm",
        ".avif": "image/avif",
    }),
)
```

## Mounting Strategies

### SPA at Root

```go
// All routes: serve from dist/, unknown → index.html
frontend.Register(router, "./dist",
    frontend.WithPrefix("/"),
    frontend.WithFallback(true),
)

// Requests:
// GET /          → dist/index.html
// GET /about     → dist/index.html (SPA handles routing)
// GET /assets/main.js → dist/assets/main.js (static file)
```

### Sub-path Mount

```go
// Mount dashboard SPA at /dashboard
frontend.Register(router, "./dashboard/dist",
    frontend.WithPrefix("/dashboard"),
    frontend.WithFallback(true),
)

// Mount marketing site at /
frontend.Register(router, "./marketing/dist",
    frontend.WithPrefix("/"),
    frontend.WithFallback(false), // No SPA fallback
)
```

### API + Frontend on Same Port

```go
// API routes registered first
app.Get("/api/users", handleListUsers)
app.Post("/api/users", handleCreateUser)

// Frontend catches everything else
frontend.Register(app.Router(), "./dist",
    frontend.WithPrefix("/"),
    frontend.WithFallback(true),
)
```

## Module Documentation

- **[Static Files](static-files.md)** — Serving from filesystem
- **[Embedded Assets](embedded-assets.md)** — Bundling into binary

## Related Documentation

- [Core Module](../core/README.md) — Application setup
- [Router Module](../router/README.md) — Route registration
- [Security: Headers](../security/headers.md) — Security headers for frontend
