# Reverse Routing

> **Package**: `github.com/spcent/plumego/router`

Reverse routing maps a route name to a URL path so handlers, redirects, and links do not hard-code path strings.

## Canonical APIs

- Register named route:
  - `r.AddRouteWithOptions(..., router.WithRouteName("name"))`
  - `r.GetNamed/ PostNamed / PutNamed / DeleteNamed / PatchNamed / AnyNamed`
  - `app.GetNamed / PostNamed / PutNamed / DeleteNamed / PatchNamed / AnyNamed`
- Generate URL:
  - `r.URL(name, "param", "value", ...) string`
  - `r.URLMust(name, "param", "value", ...) string`

`URL` returns empty string when name is unknown. `URLMust` panics for unknown names.

## Quick Start (core.App)

```go
app := core.New(core.WithAddr(":8080"))

app.GetNamed("users.show", "/users/:id", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
})

url := app.Router().URL("users.show", "id", "42")
// /users/42
_ = url
```

Compile-oriented complete example:

```go
package main

import (
    "log"
    "net/http"

    "github.com/spcent/plumego/core"
)

func main() {
    app := core.New(core.WithAddr(":8080"))

    app.GetNamed("users.show", "/users/:id", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    app.GetNamed("users.posts", "/users/:id/posts/:postId", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    _ = app.Router().URL("users.show", "id", "42")
    _ = app.Router().URL("users.posts", "id", "42", "postId", "7")

    if err := app.Boot(); err != nil {
        log.Fatal(err)
    }
}
```

## Quick Start (standalone router)

```go
r := router.NewRouter()

r.GetNamed("home", "/", http.HandlerFunc(homeHandler))
r.GetNamed("users.show", "/users/:id", http.HandlerFunc(showUser))

homeURL := r.URL("home")
showURL := r.URL("users.show", "id", "42")
```

## Named Routes in Groups

```go
r := app.Router()
api := r.Group("/api/v1")

api.GetNamed("users.index", "/users", http.HandlerFunc(listUsers))
api.GetNamed("users.show", "/users/:id", http.HandlerFunc(showUser))

u := r.URL("users.show", "id", "99")
// /api/v1/users/99
_ = u
```

Group prefixes are included automatically in generated URLs.

## Parameter Rules

`URL` consumes parameters as alternating key/value pairs:

```go
u := r.URL("post.comment", "postId", "10", "commentId", "5")
// /posts/10/comments/5
```

Behavior details:

- Unknown route name: returns `""`.
- Odd number of args: trailing key is ignored.
- Missing `:param`: placeholder remains in output (for diagnostics).
- Missing `*wildcard`: wildcard segment resolves to empty suffix.
- Values are path-escaped (`url.PathEscape`).

## Register via Route Options

```go
if err := r.AddRouteWithOptions(
    router.GET,
    "/files/:tenant/*path",
    http.HandlerFunc(downloadFile),
    router.WithRouteName("files.show"),
); err != nil {
    log.Fatal(err)
}

u := r.URL("files.show", "tenant", "acme corp", "path", "reports/2026 q1.pdf")
// /files/acme%20corp/reports/2026%20q1.pdf
_ = u
```

## Introspection

```go
if r.HasRoute("users.show") {
    meta := r.NamedRoutes()["users.show"]
    fmt.Println(meta.Method, meta.Pattern)
}
```

Use this for debugging route maps and build-time assertions.

## Testing Guidance

For reverse-routing changes, keep tests for:

- named route registration (including group prefixes)
- URL generation with param + wildcard escaping
- missing param behavior
- name collisions/override rules
- `URLMust` panic behavior

Repository references:

- `router/reverse_routing_group_test.go`
- `core/routing_test.go`
