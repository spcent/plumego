# Serving Static Files

> **Package**: `github.com/spcent/plumego/frontend` | **Feature**: Filesystem serving

## Overview

Serve static files from the local filesystem — ideal for development and when files are deployed alongside the binary.

## Basic Setup

```go
import (
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/frontend"
)

app := core.New(core.WithAddr(":8080"))

// Serve SPA from ./dist directory
err := frontend.Register(app.Router(), "./dist",
    frontend.WithPrefix("/"),
    frontend.WithFallback(true),
)
if err != nil {
    log.Fatal(err)
}

app.Boot()
```

## SPA Routing (Fallback Mode)

Single Page Applications handle routing client-side. The server must return `index.html` for all unknown paths:

```go
frontend.Register(router, "./dist",
    frontend.WithFallback(true), // Enable SPA fallback
)

// Request flow:
// GET /             → dist/index.html    (found)
// GET /login        → dist/index.html    (fallback, SPA handles routing)
// GET /dashboard    → dist/index.html    (fallback)
// GET /assets/app.js → dist/assets/app.js (static file, served directly)
// GET /favicon.ico  → dist/favicon.ico  (static file, served directly)
```

## Caching Strategy

```go
frontend.Register(router, "./dist",
    // Assets with content hash in filename: cache forever
    frontend.WithCacheControl("public, max-age=31536000, immutable"),

    // index.html: never cache (always get latest)
    frontend.WithIndexCacheControl("no-cache, no-store, must-revalidate"),
)
```

**Why different policies?**
- Asset files (`app.abc123.js`) have content hashes → safe to cache forever
- `index.html` references assets by hash → must always be fresh

## Pre-compressed Files

Serve `.gz` or `.br` files when the client supports compression:

```go
frontend.Register(router, "./dist",
    frontend.WithPrecompressed(true),
)
```

**File structure**:
```
dist/
├── assets/
│   ├── main.js          # Served when no compression
│   ├── main.js.gz       # Served for Accept-Encoding: gzip
│   └── main.js.br       # Served for Accept-Encoding: br (Brotli)
```

**Build step** (Vite/webpack):
```bash
# Generate compressed versions
find dist -name "*.js" -o -name "*.css" | xargs gzip -k
find dist -name "*.js" -o -name "*.css" | xargs brotli -k
```

## Custom Error Pages

```go
frontend.Register(router, "./dist",
    frontend.WithNotFoundPage("dist/404.html"),  // Custom 404
    frontend.WithErrorPage("dist/500.html"),      // Custom 5xx
)
```

```html
<!-- dist/404.html -->
<!DOCTYPE html>
<html>
<head><title>Page Not Found</title></head>
<body>
  <h1>404 - Page Not Found</h1>
  <a href="/">Return Home</a>
</body>
</html>
```

## Security Headers

```go
frontend.Register(router, "./dist",
    frontend.WithHeaders(map[string]string{
        "X-Content-Type-Options":  "nosniff",
        "X-Frame-Options":         "SAMEORIGIN",
        "Referrer-Policy":         "strict-origin-when-cross-origin",
        "Permissions-Policy":      "geolocation=(), camera=()",
    }),
)
```

## Multiple Applications

```go
// Mount marketing website at /
frontend.Register(router, "./marketing/dist",
    frontend.WithPrefix("/"),
    frontend.WithFallback(false), // Traditional MPA, no fallback
)

// Mount admin SPA at /admin
frontend.Register(router, "./admin/dist",
    frontend.WithPrefix("/admin"),
    frontend.WithFallback(true), // SPA, needs fallback
)

// API routes
router.Get("/api/v1/users", handleUsers)
```

## Development vs Production

```go
func setupFrontend(app *core.App) error {
    debug := os.Getenv("APP_DEBUG") == "true"

    if debug {
        // Development: serve from disk (hot-reload friendly)
        return frontend.Register(app.Router(), "./frontend/dist",
            frontend.WithPrefix("/"),
            frontend.WithFallback(true),
            frontend.WithIndexCacheControl("no-cache"),
        )
    }

    // Production: serve embedded assets
    return frontend.RegisterEmbedded(app.Router(),
        frontend.WithPrefix("/"),
        frontend.WithFallback(true),
        frontend.WithCacheControl("public, max-age=31536000, immutable"),
    )
}
```

## Directory Layout

Recommended project structure:

```
myapp/
├── main.go
├── frontend/
│   ├── src/          # Source files (React, Vue, etc.)
│   ├── dist/         # Build output (git-ignored)
│   └── package.json
└── api/
    └── handlers.go
```

```bash
# Build frontend
cd frontend && npm run build

# Run Go with built frontend
go run . # Serves from ./frontend/dist
```

## MIME Types

The package uses Go's standard MIME detection. Override for custom types:

```go
frontend.Register(router, "./dist",
    frontend.WithMIMETypes(map[string]string{
        ".wasm":    "application/wasm",
        ".avif":    "image/avif",
        ".webmanifest": "application/manifest+json",
    }),
)
```

## Related Documentation

- [Embedded Assets](embedded-assets.md) — Bundle assets into binary
- [Frontend Overview](README.md) — Module overview
- [Security: Headers](../../security/headers.md) — Security headers
