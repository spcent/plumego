# Standard Service Reference

`reference/standard-service` is Plumego's canonical reference application.

Read this immediately after `docs/getting-started.md`. This directory is the
default answer to "what should a Plumego application look like?"

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

Canonical next reads after this reference:

1. `docs/README.md`
2. `docs/CANONICAL_STYLE_GUIDE.md`
3. `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
4. `specs/task-routing.yaml`
