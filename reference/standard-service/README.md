# Standard Service Reference

`reference/standard-service` is Plumego's canonical reference application.

It is the single source of truth for:

- default application layout
- bootstrap flow in `main.go`
- route registration style
- handler and response conventions
- minimal stable-root-only wiring

Design constraints:

- depends only on `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, and `metrics` plus the standard library
- excludes `x/*` capabilities from the canonical learning path
- keeps all route registration explicit in `internal/app/routes.go`
- keeps app-local configuration in `internal/config`

Canonical files:

- `main.go`
- `internal/app/app.go`
- `internal/app/routes.go`
- `internal/handler/api.go`
- `internal/handler/health.go`
- `internal/config/config.go`

Run it with:

```bash
go run ./reference/standard-service
```
