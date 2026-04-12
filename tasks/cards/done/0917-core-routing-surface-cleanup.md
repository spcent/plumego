# Card 0917

Priority: P2
State: done
Primary Module: core
Owned Files:
- `core/routing.go`
Depends On: 0827 (router public surface pruning, not a hard dependency but related)

Goal:
- Remove the confusing `addNamedRoute` indirection whose only purpose is to swap
  `name` and `path` arguments.
- Align the convenience route methods (`Get`, `Post`, `Put`, `Delete`, `Patch`, `Any`)
  to accept `http.Handler` instead of `http.HandlerFunc`, matching `AddRoute`.

---

## Issue A — `addNamedRoute` parameter order swap creates silent confusion

**File:** `core/routing.go` lines 37–43, 60–63

```go
// Public API: (method, path, name, handler)
func (a *App) AddRouteWithName(method, path, name string, handler http.Handler) error {
    return a.addNamedRoute(method, name, path, handler)  // ← name and path are intentionally swapped
}

// Private helper: (method, name, path, handler)
func (a *App) addNamedRoute(method, name, path string, handler http.Handler) error {
    return a.registerRoute(method, path, name, handler)  // ← swapped back again
}
```

`addNamedRoute` exists solely to perform an argument swap. It holds no logic,
no validation, and no documentation. The double-swap is confusing to anyone tracing
through the call chain — the intermediate function looks like it's calling `registerRoute`
with `path` in the wrong position until you realize the parameters themselves were
re-ordered.

**Fix:** Call `registerRoute` directly from `AddRouteWithName` and delete `addNamedRoute`:

```go
func (a *App) AddRouteWithName(method, path, name string, handler http.Handler) error {
    return a.registerRoute(method, path, name, handler)
}
```

---

## Issue B — Convenience methods accept `http.HandlerFunc`, `AddRoute` accepts `http.Handler`

**File:** `core/routing.go` lines 74–103

```go
// Convenience methods — accept concrete type
func (a *App) Get(path string, handler http.HandlerFunc) error { ... }
func (a *App) Post(path string, handler http.HandlerFunc) error { ... }
// ...

// Explicit methods — accept interface
func (a *App) AddRoute(method, path string, handler http.Handler) error { ... }
func (a *App) AddRouteWithName(method, path, name string, handler http.Handler) error { ... }
```

`http.HandlerFunc` implements `http.Handler`, but the reverse is not true. A caller who
holds an `http.Handler` value (a struct implementing `ServeHTTP`, or the result of
`http.StripPrefix`, `http.TimeoutHandler`, etc.) cannot pass it directly to `Get`, `Post`,
etc. — they must use `AddRoute`. This forces callers to choose the registration method
based on the type of handler value they hold, not based on their intent (GET route vs
named route).

The canonical handler shape in the repo is `func(w http.ResponseWriter, r *http.Request)`
(from `CLAUDE.md`). Convenience methods should accept `http.Handler` so they are equally
flexible regardless of whether the caller uses a function literal or a struct.

**Fix:** Change all convenience method signatures to `http.Handler`:

```go
func (a *App) Get(path string, handler http.Handler) error { ... }
func (a *App) Post(path string, handler http.Handler) error { ... }
func (a *App) Put(path string, handler http.Handler) error { ... }
func (a *App) Delete(path string, handler http.Handler) error { ... }
func (a *App) Patch(path string, handler http.Handler) error { ... }
func (a *App) Any(path string, handler http.Handler) error { ... }
```

`http.HandlerFunc` already satisfies `http.Handler`, so all existing call sites continue
to compile without change.

---

Scope:
- Delete `addNamedRoute`; inline its call into `AddRouteWithName`.
- Change convenience method signatures from `http.HandlerFunc` to `http.Handler`.
- No behavior changes; callers with function literals continue to work unchanged.

Non-goals:
- Do not change `registerRoute` internals.
- Do not change the router package.
- Do not add new route-registration methods.

Files:
- `core/routing.go`
- `core/routing_test.go`

Tests:
- `grep -rn 'addNamedRoute' . --include='*.go'` returns empty after deletion.
- `go build ./...`
- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

Docs Sync: —

Done Definition:
- `addNamedRoute` is deleted; `AddRouteWithName` calls `registerRoute` directly.
- All six convenience methods accept `http.Handler`.
- `go build ./...` passes; no call sites need updating.
- All tests pass.

Outcome:
