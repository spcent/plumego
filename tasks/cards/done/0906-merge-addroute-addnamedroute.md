# Card 0906

Priority: P3
State: done
Primary Module: core
Owned Files:
- `core/routing.go`
Depends On: —

Goal:
- Merge `addRoute()` and `addNamedRoute()` into a single private function to eliminate code duplication.

Problem:
`core/routing.go` has two near-identical private functions:

```go
func (a *App) addRoute(method, path string, handler http.Handler) error {
    if handler == nil {
        return contract.WrapError(contract.ErrHandlerNil, "add_route", "core",
            map[string]any{"method": method, "path": path})
    }
    if err := a.ensureMutable("add_route", "register route"); err != nil { return err }
    r := a.ensureRouter()
    if r == nil { return wrapCoreError(...) }
    return r.AddRoute(method, path, handler)
}

func (a *App) addNamedRoute(method, name, path string, handler http.Handler) error {
    if handler == nil {
        return contract.WrapError(contract.ErrHandlerNil, "add_route", "core",
            map[string]any{"method": method, "path": path, "name": name})
    }
    if err := a.ensureMutable("add_route", "register route"); err != nil { return err }
    r := a.ensureRouter()
    if r == nil { return wrapCoreError(...) }
    return r.AddRouteWithName(method, path, name, handler)
}
```

The only differences are: the `name` parameter and the final router call.
Divergent copies invite drift — a fix to one (e.g. a new error key, a logging call)
must be remembered in both.

Fix:
- Replace both with a single `func (a *App) registerRoute(method, path, name string, handler http.Handler) error`.
- When `name == ""`, call `r.AddRoute(method, path, handler)`.
- When `name != ""`, call `r.AddRouteWithName(method, path, name, handler)`.
- Update `addRoute` and `addNamedRoute` to be thin one-line delegators (or inline them into their callers).

Non-goals:
- Do not change the public `AddRoute`, `AddRouteWithName`, or shorthand methods.
- Do not change routing behaviour.

Files:
- `core/routing.go`

Tests:
- `go build ./...`
- `go test -timeout 20s ./core/...`
- `go vet ./...`

Docs Sync: —

Done Definition:
- A single private implementation function handles both named and unnamed route registration.
- No logic is duplicated between the two paths.
- All existing public methods delegate to it correctly.
- All tests pass.

Outcome:
- `registerRoute(method, path, name string, handler http.Handler) error` added as single implementation.
- `addRoute` and `addNamedRoute` reduced to one-line delegators to `registerRoute`.
- Handler nil check, `ensureMutable`, and `ensureRouter` logic no longer duplicated.
- `go test -timeout 20s ./...` passes.
