# Card 0752

Priority: P1
State: done
Primary Module: core
Owned Files:
- `core/lifecycle.go`
- `core/lifecycle_test.go`
- `cmd/plumego/internal/devserver/dashboard.go`
- `docs/modules/core/README.md`
- `README.md`
Depends On:
- `0751-core-router-option-order-independence.md`

Goal:
- Shrink `core` to one canonical lifecycle path by removing internal reliance on
  the legacy `Boot`/`Run` convenience flow.

Problem:
- `Boot` and `Run` duplicate the explicit `Prepare` + `Start` + `Server` +
  `Shutdown` path that the reference app already uses.
- `Run` also owns signal watching and server start internally, while the
  explicit path leaves serving to the caller. That creates two lifecycle models
  inside the kernel.
- The dev dashboard still uses `Boot`, so the legacy surface remains live in
  first-party code even though `core/lifecycle.go` already labels it legacy.

Scope:
- Migrate the last first-party caller in the dev dashboard off `Boot`.
- Keep one explicit first-party startup pattern centered on `Prepare`,
  `Start`, `Server`, and `Shutdown`.
- Remove `Boot` and `Run` once first-party callers are migrated so `core`
  exposes only one lifecycle model.

Non-goals:
- Do not change HTTP server semantics, shutdown timeouts, or connection
  draining behavior in this card.
- Do not introduce new lifecycle abstractions.
- Do not move signal handling into extensions.

Files:
- `core/lifecycle.go`
- `core/lifecycle_test.go`
- `cmd/plumego/internal/devserver/dashboard.go`
- `docs/modules/core/README.md`
- `README.md`

Tests:
- Update lifecycle tests so the canonical path exercises `Prepare`, `Start`,
  `Server`, and `Shutdown` directly.
- `go test -race -timeout 60s ./core/...`
- `go test -timeout 20s ./cmd/plumego/internal/devserver/...`

Docs Sync:
- Remove `Boot`/`Run` from user-facing startup guidance.
- Keep root README and module primer aligned on the explicit lifecycle path.

Done Definition:
- `Boot` and `Run` are removed from `core`.
- The canonical startup documentation uses only the explicit lifecycle path.
- Lifecycle tests cover the canonical path and pass.

Outcome:
- Removed `Boot` and `Run` from `core/lifecycle.go`, and deleted the now-dead
  internal signal-watcher helper tied to that legacy flow.
- Migrated `cmd/plumego/internal/devserver/dashboard.go` to the explicit
  `Prepare` + `Start` + `Server` + `Shutdown` lifecycle path.
- Reworked lifecycle tests to exercise the canonical explicit path instead of
  `Boot`.
- Updated startup guidance in `README.md`, `README_CN.md`,
  `docs/getting-started.md`, and `docs/modules/core/README.md` to show the
  explicit lifecycle only.
- Cleared lingering `app.Boot()` examples from `x/gateway` comments/docs.
- Validation:
  - `gofmt -w core/lifecycle.go core/lifecycle_test.go cmd/plumego/internal/devserver/dashboard.go`
  - `go test -race -timeout 60s ./core/...`
  - `go test -timeout 20s ./internal/devserver/...` (run from `cmd/plumego/`)
  - `go vet ./core/...`
  - `go vet ./internal/devserver/...` (run from `cmd/plumego/`)
