# Frontend Package

The `frontend` package provides static file serving for built frontend applications (React, Vue, Next.js, Vite, etc.) with advanced features for production deployments.

## Features

- ✅ **Dual Source Support**: Serve from disk directories or embedded filesystems
- ✅ **SPA Routing**: Automatic fallback to `index.html` for client-side routing
- ✅ **Pre-compressed Files**: Serve `.gz` and `.br` files automatically when available
- ✅ **Custom Error Pages**: Configure custom 404 and 5xx error pages
- ✅ **MIME Type Overrides**: Set custom content types for specific file extensions
- ✅ **Cache Control**: Separate caching strategies for index vs. assets
- ✅ **Security**: Path traversal protection and method restrictions
- ✅ **Custom Headers**: Apply security headers or custom metadata
- ✅ **Flexible Mounting**: Mount at any URL prefix (`/`, `/app`, etc.)

## Quick Start

### Basic Usage

```go
import (
    "github.com/spcent/plumego/frontend"
    "github.com/spcent/plumego/router"
)

r := router.NewRouter()

// Serve a built frontend directory (e.g., Next.js `out/`, Vite `dist/`)
err := frontend.RegisterFromDir(r, "./dist",
    frontend.WithCacheControl("public, max-age=31536000"),
    frontend.WithIndexCacheControl("no-cache"),
)
```

### Embedded Filesystem

```go
import (
    "embed"
    "io/fs"
    "net/http"
)

//go:embed dist/*
var distFS embed.FS

func setupFrontend(r *router.Router) {
    subFS, _ := fs.Sub(distFS, "dist")
    frontend.RegisterFS(r, http.FS(subFS),
        frontend.WithPrefix("/"),
        frontend.WithFallback(true),
    )
}
```

## Configuration Options

### WithPrefix(prefix string)

Mount the frontend at a specific URL prefix.

```go
// Mount at /app
frontend.WithPrefix("/app")

// Access via: /app, /app/assets/*, /app/page1, etc.
```

**Default**: `/`

### WithIndex(filename string)

Override the default index file name.

```go
frontend.WithIndex("main.html")
```

**Default**: `index.html`

### WithCacheControl(header string)

Set `Cache-Control` header for asset files (JS, CSS, images, etc.).

```go
frontend.WithCacheControl("public, max-age=31536000, immutable")
```

**Default**: No caching (empty)

### WithIndexCacheControl(header string)

Set `Cache-Control` header for index file responses.

```go
frontend.WithIndexCacheControl("no-cache, must-revalidate")
```

**Default**: No caching (empty)

**Why separate?** Index files should not be cached (to ensure users get updates), while assets with hashed names can be cached forever.

### WithFallback(enabled bool)

Enable SPA fallback mode. When enabled, requests for missing files return `index.html` instead of 404.

```go
frontend.WithFallback(true)  // Enable for SPAs
frontend.WithFallback(false) // Disable for static sites
```

**Default**: `true`

### WithPrecompressed(enabled bool) ⭐ NEW

Enable automatic serving of pre-compressed files (`.gz`, `.br`).

When enabled:
- Request for `app.js` → serves `app.js.br` if client supports Brotli
- Request for `app.js` → serves `app.js.gz` if client supports Gzip
- Request for `app.js` → serves `app.js` if no pre-compressed version exists

```go
frontend.WithPrecompressed(true)
```

**Benefits**:
- Faster responses (no runtime compression)
- Lower CPU usage
- Better compression ratios (pre-compressed at build time)

**Build Setup**:
```bash
# Using Brotli (best compression)
find dist -type f \( -name '*.js' -o -name '*.css' -o -name '*.html' \) \
    -exec brotli -q 11 -o {}.br {} \;

# Using Gzip (widely supported)
find dist -type f \( -name '*.js' -o -name '*.css' -o -name '*.html' \) \
    -exec gzip -9 -k {} \;
```

**Default**: `false`

### WithNotFoundPage(path string) ⭐ NEW

Set a custom 404 error page. Path is relative to the filesystem root.

```go
frontend.WithNotFoundPage("404.html")
```

**Example** `404.html`:
```html
<!DOCTYPE html>
<html>
<head><title>Page Not Found</title></head>
<body>
    <h1>404 - Page Not Found</h1>
    <p>The page you're looking for doesn't exist.</p>
    <a href="/">Go Home</a>
</body>
</html>
```

**Default**: Standard `http.NotFound` response

### WithErrorPage(path string) ⭐ NEW

Set a custom 5xx error page for server errors.

```go
frontend.WithErrorPage("500.html")
```

**Default**: Standard `http.Error` response

### WithMIMETypes(types map[string]string) ⭐ NEW

Override MIME types for specific file extensions.

```go
frontend.WithMIMETypes(map[string]string{
    ".wasm": "application/wasm",
    ".json": "application/json; charset=utf-8",
    ".webp": "image/webp",
})
```

**Use Cases**:
- WebAssembly files (`.wasm`)
- Custom file formats
- Charset specification
- Modern image formats (WebP, AVIF)

**Default**: Uses `http.ServeContent` auto-detection

### WithHeaders(headers map[string]string)

Apply custom headers to all successful responses.

```go
frontend.WithHeaders(map[string]string{
    "X-Frame-Options":        "DENY",
    "X-Content-Type-Options": "nosniff",
    "Referrer-Policy":        "strict-origin-when-cross-origin",
})
```

## Production Configuration

Recommended production setup with all features:

```go
err := frontend.RegisterFromDir(r, "./dist",
    // Mounting
    frontend.WithPrefix("/"),
    frontend.WithIndex("index.html"),

    // Caching Strategy
    frontend.WithCacheControl("public, max-age=31536000, immutable"),
    frontend.WithIndexCacheControl("no-cache, must-revalidate"),

    // Performance
    frontend.WithPrecompressed(true), // Serve .gz/.br files

    // Routing
    frontend.WithFallback(true), // SPA mode

    // Custom Error Pages
    frontend.WithNotFoundPage("404.html"),
    frontend.WithErrorPage("500.html"),

    // MIME Types
    frontend.WithMIMETypes(map[string]string{
        ".wasm": "application/wasm",
        ".json": "application/json; charset=utf-8",
    }),

    // Security Headers
    frontend.WithHeaders(map[string]string{
        "X-Frame-Options":        "DENY",
        "X-Content-Type-Options": "nosniff",
        "Referrer-Policy":        "strict-origin-when-cross-origin",
        "Permissions-Policy":     "geolocation=(), microphone=()",
    }),
)
```

## Build Optimization Guide

### Pre-compression Workflow

1. **Build your frontend**:
   ```bash
   npm run build
   # Output: dist/
   ```

2. **Pre-compress assets**:
   ```bash
   # Brotli (best compression, ~20% better than gzip)
   find dist -type f \( -name '*.js' -o -name '*.css' -o -name '*.html' -o -name '*.svg' \) \
       -exec brotli -q 11 -o {}.br {} \;

   # Gzip (universal compatibility)
   find dist -type f \( -name '*.js' -o -name '*.css' -o -name '*.html' -o -name '*.svg' \) \
       -exec gzip -9 -k {} \;
   ```

3. **Embed if desired**:
   ```go
   //go:embed dist/*
   var distFS embed.FS
   ```

### Build Tools Integration

**Vite** (`vite.config.js`):
```js
import { defineConfig } from 'vite'
import compress from 'vite-plugin-compression'

export default defineConfig({
  plugins: [
    compress({ algorithm: 'gzip' }),
    compress({ algorithm: 'brotliCompress', ext: '.br' }),
  ],
})
```

**Next.js** (`next.config.js`):
```js
module.exports = {
  compress: false, // Let plumego handle compression
  output: 'export', // Static export
}
```

Then use a build script to pre-compress.

## Caching Strategy

### Recommended Approach

```go
frontend.WithCacheControl("public, max-age=31536000, immutable")
frontend.WithIndexCacheControl("no-cache, must-revalidate")
```

**Why?**
- **Assets** (`app.abc123.js`): Hash-based names → cache forever
- **Index** (`index.html`): No hash → don't cache, always fresh

### Build Tool Setup

Ensure your build tool generates hashed filenames:

**Vite**: Enabled by default
**Next.js**: Enabled by default
**Create React App**: Enabled by default

## Security Considerations

### Path Traversal Protection

Built-in protection against:
- `../` sequences
- Absolute paths
- Directory escapes

### Method Restrictions

Only `GET` and `HEAD` requests are allowed. Other methods return `405 Method Not Allowed`.

### Recommended Headers

```go
frontend.WithHeaders(map[string]string{
    "X-Frame-Options":         "DENY",
    "X-Content-Type-Options":  "nosniff",
    "Referrer-Policy":         "strict-origin-when-cross-origin",
    "Content-Security-Policy": "default-src 'self'",
    "Permissions-Policy":      "geolocation=(), microphone=(), camera=()",
})
```

## Advanced Examples

### Multiple Frontends

Serve different apps at different paths:

```go
// Main app at /
frontend.RegisterFromDir(r, "./main-app", frontend.WithPrefix("/"))

// Admin panel at /admin
frontend.RegisterFromDir(r, "./admin-app", frontend.WithPrefix("/admin"))

// Docs at /docs
frontend.RegisterFromDir(r, "./docs", frontend.WithPrefix("/docs"))
```

### API + Frontend

```go
// API routes
r.Get("/api/users", getUsersHandler)
r.Post("/api/users", createUserHandler)

// Frontend (must be registered last to avoid catching API routes)
frontend.RegisterFromDir(r, "./dist",
    frontend.WithPrefix("/"),
    frontend.WithFallback(true),
)
```

### Embedded with Fallback to Disk (Development)

```go
//go:embed dist/*
var embeddedFS embed.FS

func setupFrontend(r *router.Router) {
    // Try embedded first
    if frontend.HasEmbedded() {
        subFS, _ := fs.Sub(embeddedFS, "dist")
        frontend.RegisterFS(r, http.FS(subFS))
        return
    }

    // Fallback to disk for development
    frontend.RegisterFromDir(r, "./dist")
}
```

## Performance Tips

1. **Use pre-compressed files** (`WithPrecompressed(true)`)
2. **Set long cache times for assets** (`max-age=31536000`)
3. **Don't cache index files** (`no-cache`)
4. **Use hashed filenames** (enabled by default in modern build tools)
5. **Embed assets in production** (single binary deployment)
6. **Combine with middleware.Gzip()** for dynamic responses

## Testing

The package includes comprehensive tests:

```bash
go test ./frontend/...
```

See `frontend_test.go` for test examples covering:
- Pre-compressed file serving
- Custom error pages
- MIME type overrides
- Security (path traversal, method restrictions)
- Cache control behavior
- SPA fallback routing

## Compatibility

- **Go**: 1.24+ (as per plumego requirements)
- **Build Tools**: Any (Vite, Next.js, Create React App, etc.)
- **Browsers**: All modern browsers (gzip/brotli support)

## Related Packages

- `middleware.Gzip()`: Dynamic compression for API responses
- `core.NewFrontendComponentFromDir()`: Component-based registration
- `core.NewFrontendComponentFromFS()`: Component-based embedded serving

## License

See the main plumego repository for license information.
