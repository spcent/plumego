# Card 0787

Priority: P1
State: active
Primary Module: core
Owned Files:
- `core/app.go`
- `core/app_helpers.go`
- `core/http_handler.go`
- `core/routing.go`
- `core/options_test.go`
- `core/app_test.go`
Depends On:
- `0786-core-internal-preparation-state-unification.md`

Goal:
- Remove hidden router-policy synchronization from read and helper paths so
  `core` stops mutating owned router state during lookup or preparation.

Problem:
- `ensureRouter()` currently calls `syncRouterConfig(...)` every time it is
  used.
- That means read-only paths such as `Routes()`, `URL(...)`, and handler/server
  preparation still perform hidden policy mutation on the owned router.
- Since `core` already owns both the config and router lifecycle, this
  continuous re-sync is unnecessary and makes simple reads carry side effects.

Scope:
- Apply router policy at one canonical ownership point instead of every
  `ensureRouter()` call.
- Make route lookup / introspection paths side-effect free.
- Keep the owned router policy aligned with typed config without hidden repair.

Non-goals:
- Do not reintroduce raw router injection or replacement.
- Do not redesign router policy fields.
- Do not broaden `core` introspection.

Files:
- `core/app.go`
- `core/app_helpers.go`
- `core/http_handler.go`
- `core/routing.go`
- `core/options_test.go`
- `core/app_test.go`

Tests:
- `go test -race -timeout 60s ./core/...`
- `go vet ./core/...`

Docs Sync:
- Docs only if the canonical ownership point for router policy becomes
  user-visible.

Done Definition:
- Router policy is applied once at an explicit ownership point.
- `Routes()` / `URL(...)` / preparation helpers do not mutate router policy as
  a side effect.
- Tests assert the explicit router-policy ownership path.
