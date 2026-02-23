# Embedded Static Assets

> **Package**: `github.com/spcent/plumego/frontend` | **Feature**: Binary bundling with go:embed

## Overview

Embed frontend assets directly into the Go binary using `go:embed`. This produces a single deployable binary with no external file dependencies — ideal for production deployments, Docker containers, and CLI tools.

## Quick Start

```go
package main

import (
    "embed"
    "net/http"

    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/frontend"
)

// Embed entire dist/ directory
//go:embed dist
var distFS embed.FS

func main() {
    app := core.New(core.WithAddr(":8080"))

    // Mount embedded SPA
    err := frontend.RegisterFS(app.Router(), http.FS(distFS),
        frontend.WithPrefix("/"),
        frontend.WithFallback(true),
        frontend.WithCacheControl("public, max-age=31536000, immutable"),
        frontend.WithIndexCacheControl("no-cache, no-store, must-revalidate"),
    )
    if err != nil {
        panic(err)
    }

    app.Boot()
}
```

## Using the Embedded Directory

Plumego provides a pre-configured embedded directory at `frontend/embedded/`:

```go
// Place your built frontend in frontend/embedded/
// Then use RegisterEmbedded:
err := frontend.RegisterEmbedded(app.Router(),
    frontend.WithPrefix("/"),
    frontend.WithFallback(true),
)
```

**Directory structure**:
```
frontend/
├── embedded/        # Place built frontend here
│   ├── index.html
│   ├── assets/
│   │   ├── main.abc123.js
│   │   └── main.abc123.css
│   └── .keep        # Placeholder (keep this file)
```

## Build Integration

### Makefile

```makefile
.PHONY: build frontend

frontend:
    cd web && npm run build
    rm -rf frontend/embedded/*
    cp -r web/dist/* frontend/embedded/

build: frontend
    go build -o myapp .

# Or build without frontend (for development)
build-dev:
    go build -o myapp .
```

### GitHub Actions

```yaml
jobs:
  build:
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Build frontend
        run: |
          cd web
          npm ci
          npm run build
          cp -r dist/* ../frontend/embedded/

      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Build binary
        run: go build -o myapp .

      - name: Upload binary
        uses: actions/upload-artifact@v4
        with:
          name: myapp
          path: myapp
```

## Multiple Embedded Filesystems

```go
//go:embed dist
var mainAppFS embed.FS

//go:embed admin/dist
var adminAppFS embed.FS

// Mount both
frontend.RegisterFS(router, http.FS(mainAppFS),
    frontend.WithPrefix("/"),
    frontend.WithFallback(true),
)

frontend.RegisterFS(router, http.FS(adminAppFS),
    frontend.WithPrefix("/admin"),
    frontend.WithFallback(true),
)
```

## Sub-FS for Correct Root

When using `go:embed`, the embedded path includes the directory name. Use `fs.Sub` to adjust the root:

```go
//go:embed dist
var distFS embed.FS

// Without Sub: requests must be like /dist/index.html
// With Sub: requests can be /index.html

import "io/fs"

subFS, err := fs.Sub(distFS, "dist")
if err != nil {
    log.Fatal(err)
}

frontend.RegisterFS(router, http.FS(subFS),
    frontend.WithPrefix("/"),
    frontend.WithFallback(true),
)
```

## Development/Production Switch

```go
//go:embed dist
var embeddedDist embed.FS

func setupFrontend(router *router.Router) error {
    if os.Getenv("SERVE_FROM_DISK") == "true" {
        // Development: hot-reload from disk
        return frontend.Register(router, "./web/dist",
            frontend.WithPrefix("/"),
            frontend.WithFallback(true),
            frontend.WithIndexCacheControl("no-cache"),
        )
    }

    // Production: serve embedded (zero external dependencies)
    subFS, _ := fs.Sub(embeddedDist, "dist")
    return frontend.RegisterFS(router, http.FS(subFS),
        frontend.WithPrefix("/"),
        frontend.WithFallback(true),
        frontend.WithCacheControl("public, max-age=31536000, immutable"),
    )
}
```

## Vite Configuration

Configure Vite to output files compatible with embedding:

```javascript
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
    plugins: [react()],
    build: {
        outDir: '../frontend/embedded',  // Output to Go embed directory
        emptyOutDir: true,               // Clean before build

        rollupOptions: {
            output: {
                // Content-hashed filenames for long-term caching
                entryFileNames: 'assets/[name].[hash].js',
                chunkFileNames: 'assets/[name].[hash].js',
                assetFileNames: 'assets/[name].[hash].[ext]',
            },
        },
    },
});
```

## Binary Size

Frontend assets increase binary size. Strategies to minimize:

```bash
# Check binary size
ls -lh myapp

# Without frontend
go build -o myapp-api ./cmd/api

# With minification (Vite default)
# HTML/CSS/JS are automatically minified
cd web && npm run build

# With compression
gzip -c myapp > myapp.gz
```

Typical sizes:
- Simple React app: +500KB–2MB to binary
- Large dashboard: +2MB–10MB to binary

## Checking if Assets are Embedded

```go
if frontend.HasEmbedded() {
    log.Info("Serving embedded frontend assets")
    frontend.RegisterEmbedded(app.Router())
} else {
    log.Warn("No embedded assets, serving from disk")
    frontend.Register(app.Router(), "./dist")
}
```

## Testing with Embedded Assets

```go
func TestFrontendHandler(t *testing.T) {
    router := router.New()
    err := frontend.RegisterFS(router, http.FS(distFS),
        frontend.WithPrefix("/"),
        frontend.WithFallback(true),
    )
    require.NoError(t, err)

    srv := httptest.NewServer(router)
    defer srv.Close()

    // Static file
    resp, _ := http.Get(srv.URL + "/assets/main.js")
    assert.Equal(t, 200, resp.StatusCode)
    assert.Contains(t, resp.Header.Get("Content-Type"), "javascript")

    // SPA fallback
    resp, _ = http.Get(srv.URL + "/some/spa/route")
    assert.Equal(t, 200, resp.StatusCode)
    body, _ := io.ReadAll(resp.Body)
    assert.Contains(t, string(body), "<!DOCTYPE html>")
}
```

## Related Documentation

- [Static Files](static-files.md) — Serving from filesystem
- [Frontend Overview](README.md) — Module overview
