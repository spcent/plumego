# Serving Static Files

> **Package**: `github.com/spcent/plumego/frontend` | **Feature**: Filesystem serving

## Basic Setup

```go
package main

import (
    "context"
    "log"

    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/frontend"
)

func main() {
    app := core.New(core.WithAddr(":8080"))

    if err := frontend.RegisterFromDir(app.Router(), "./dist",
        frontend.WithPrefix("/"),
        frontend.WithFallback(true),
        frontend.WithIndexCacheControl("no-cache"),
    ); err != nil {
        log.Fatal(err)
    }

    if err := app.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

## SPA Routing

```go
if err := frontend.RegisterFromDir(router, "./dist",
    frontend.WithFallback(true),
); err != nil {
    return err
}
```

Request behavior:

- `/` -> `index.html`
- `/assets/app.js` -> static asset
- `/dashboard` -> `index.html` when fallback is enabled

## Recommended cache split

```go
frontend.RegisterFromDir(router, "./dist",
    frontend.WithCacheControl("public, max-age=31536000, immutable"),
    frontend.WithIndexCacheControl("no-cache, no-store, must-revalidate"),
)
```

## Development vs production

- Development: `RegisterFromDir(...)`
- Production: `RegisterFS(...)` or `RegisterEmbedded(...)`
