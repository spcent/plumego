# Serving Static Files

> **Package**: `github.com/spcent/plumego/x/frontend` | **Feature**: Filesystem serving

## Basic Setup

```go
package main

import (
    "context"
    "log"

    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/x/frontend"
)

func main() {
    app := core.New(core.WithAddr(":8080"))

    mount, err := frontend.NewMountFromDir("./dist",
        frontend.WithPrefix("/"),
        frontend.WithFallback(true),
        frontend.WithIndexCacheControl("no-cache"),
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
}
```

## SPA Routing

```go
mount, err := frontend.NewMountFromDir("./dist", frontend.WithFallback(true))
if err != nil {
    return err
}
if err := mount.Register(router); err != nil {
    return err
}
```

Request behavior:

- `/` -> `index.html`
- `/assets/app.js` -> static asset
- `/dashboard` -> `index.html` when fallback is enabled

## Recommended cache split

```go
mount, err := frontend.NewMountFromDir("./dist",
    frontend.WithCacheControl("public, max-age=31536000, immutable"),
    frontend.WithIndexCacheControl("no-cache, no-store, must-revalidate"),
)
if err != nil {
    return err
}
return mount.Register(router)
```

## Development vs production

- Development: `NewMountFromDir(...)`
- Production: `NewMountFS(...)` or `NewMountEmbedded(...)`
