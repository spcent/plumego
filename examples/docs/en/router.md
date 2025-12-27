# Router module

The **router** package implements a trie-based matcher with route groups, parameters, and middleware inheritance.

## Responsibilities
- Register HTTP handlers via `Get`, `Post`, `Put`, `Patch`, `Delete`, `Options`, `Head` and context-aware `*Ctx` variants.
- Compose route groups with `Group(prefix, ...middleware)` to reuse prefixes and middleware stacks.
- Freeze the routing table during `App.Boot()` to catch late registrations.

## Usage patterns
- Prefer RESTful paths such as `GET /users/:id` and `POST /users` to keep dashboards and docs aligned.
- Use `:name` for named parameters and `*filepath` for catch-alls (static assets, docs trees).
- Mount static frontends through `frontend.RegisterFromDir` or embedded assets; attach them to a dedicated group.

## Middleware inheritance
Middleware executes in order: **global → group → route-specific**. Group middleware is appended to the inherited stack, so admin or public surfaces can each have their own guards.

## Debugging
- Enable debug mode to log registered method/path pairs during boot.
- Add a development-only endpoint that iterates `router.Routes()` to expose the final tree.
- Panics on registrations after `Boot()` are expected—fix initialization order instead of suppressing them.

## Example
```go
api := app.Router().Group("/api")
api.Use(middleware.NewConcurrencyLimiter(100))
api.Get("/users/:id", userHandler)
api.Post("/users", createUserHandler)

assets := app.Router().Group("/docs")
assets.Get("/*filepath", docsHandler)
```

## Where to look in the repo
- `router/router.go`: router structure, node matching, and group handling.
- `frontend/register.go`: helpers for mounting static directories or embedded frontend bundles.
- `examples/reference/main.go`: real wiring of API, metrics, health, and frontend routes.
