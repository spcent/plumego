# Embedded Static Assets

> **Package**: `github.com/spcent/plumego/x/frontend` | **Feature**: Binary bundling with `go:embed`

## Using Plumego's built-in embedded directory

`frontend.NewMountEmbedded(...)` constructs a mount from assets under `frontend/embedded/` in this module.
It returns an error when that directory has no real assets.

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

    mount, err := frontend.NewMountEmbedded(
        frontend.WithPrefix("/"),
        frontend.WithFallback(true),
        frontend.WithCacheControl("public, max-age=31536000, immutable"),
        frontend.WithIndexCacheControl("no-cache, no-store, must-revalidate"),
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

## Using your own embedded filesystem

```go
//go:embed dist/*
var distFS embed.FS

subFS, err := fs.Sub(distFS, "dist")
if err != nil {
    log.Fatal(err)
}

mount, err := frontend.NewMountFS(http.FS(subFS),
    frontend.WithPrefix("/"),
    frontend.WithFallback(true),
)
if err != nil {
    log.Fatal(err)
}
if err := mount.Register(app.Router()); err != nil {
    log.Fatal(err)
}
```

## Operational notes

- Prefer content-hashed asset filenames for long-lived asset caching.
- Keep `index.html` non-cacheable so clients can discover new asset hashes.
- Use `fs.Sub(...)` when your embedded path includes a top-level `dist/` directory.
