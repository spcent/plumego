# Card 0768

Priority: P1
State: active
Primary Module: core
Owned Files:
- `core/app.go`
- `core/options.go`
- `core/options_test.go`
- `core/routing.go`
- `reference/with-webhook/internal/app/routes.go`
- `reference/with-websocket/internal/app/routes.go`
- `cmd/plumego/internal/devserver/dashboard.go`
- `x/webhook/in.go`
- `x/webhook/out.go`
- `x/websocket/websocket.go`
- `x/devtools/devtools.go`
- `x/devtools/pubsubdebug/component.go`
- `x/frontend/frontend.go`
- `x/observability/observability.go`
- `docs/modules/core/README.md`
Depends On:

Goal:
- Converge first-party route wiring onto one app-owned registration contract.

Problem:
- `core` offers strict route registration through `App.Get/Post/...`, with
  mutability checks and explicit errors.
- `core` also exposes `(*App).Router()` as a raw mutation escape hatch.
- `core.WithRouter(...)` keeps a second router ownership path even though
  first-party callers no longer need to replace the router instance.
- First-party extensions such as websocket, webhook, devtools, pubsub debug,
  observability, and frontend mounts still register routes by mutating a raw
  `*router.Router`, so `core` currently maintains both a strict app contract
  and a raw router contract for the same wiring task.

Scope:
- Remove custom router injection from `core`.
- Introduce one explicit app-owned route registration boundary for first-party
  extension wiring instead of passing around raw `*router.Router`.
- Migrate first-party route registration helpers to the chosen boundary.
- Update docs/tests so `core` is documented as the router owner.

Non-goals:
- Do not redesign the standalone `router` package.
- Do not change route matching behaviour.
- Do not change handler semantics or response shapes in this card.

Files:
- `core/app.go`
- `core/options.go`
- `core/options_test.go`
- `core/routing.go`
- `reference/with-webhook/internal/app/routes.go`
- `reference/with-websocket/internal/app/routes.go`
- `cmd/plumego/internal/devserver/dashboard.go`
- `x/webhook/in.go`
- `x/webhook/out.go`
- `x/websocket/websocket.go`
- `x/devtools/devtools.go`
- `x/devtools/pubsubdebug/component.go`
- `x/frontend/frontend.go`
- `x/observability/observability.go`
- `docs/modules/core/README.md`

Tests:
- `go test -race -timeout 60s ./core/...`
- `go test -timeout 20s ./reference/...`
- `go test -timeout 20s ./x/webhook/... ./x/websocket/... ./x/devtools/... ./x/frontend/... ./x/observability/...`
- `go test -timeout 20s ./internal/devserver/...`
- `go vet ./core/... ./reference/...`

Docs Sync:
- Update the core primer to describe `core` as the router owner and remove
  first-party reliance on raw `Router()` mutation.

Done Definition:
- First-party route wiring uses one app-owned registration surface.
- `core` no longer keeps an unused custom-router injection path.
- Raw router mutation is no longer part of first-party bootstrap.

Outcome:
