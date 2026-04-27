# Card 0137

Priority: P2
State: done
Primary Module: core
Owned Files:
- `core/config.go`
- `core/introspection.go`
- `x/devtools/devtools.go`
- `cmd/plumego/internal/devserver/config_edit.go`
- `README.md`
Depends On:
- `0136-core-route-registration-surface-convergence.md`

Goal:
- Collapse duplicated `core` introspection paths into one stable snapshot
  contract that both devtools and the dev dashboard consume consistently.

Problem:
- `DebugEnabled`, `EnvPath`, and `ConfigSnapshot` all expose overlapping app
  state in different shapes.
- `core/app_helpers.go` also keeps a second private `configSnapshot()` path,
  so config state already exists in multiple representations.
- `ConfigSnapshot` returns an untyped `map[string]any` with informal keys
  consumed by the dashboard, but it omits lifecycle fields such as shutdown and
  drain settings that still belong to the kernel runtime contract.
- This makes `core` introspection grep-unfriendly and easy to drift as new
  fields are added.

Scope:
- Define one typed snapshot source in `core` for exported runtime/config
  introspection.
- Keep any `map[string]any` shape as an adapter at the integration boundary
  where devtools or the dashboard still need it.
- Remove redundant getters instead of preserving overlapping access patterns.
- Document which snapshot fields are stable for first-party tooling.

Non-goals:
- Do not move devtools back into `core`.
- Do not redesign dashboard APIs beyond the config/introspection contract.
- Do not add feature-specific state to the kernel snapshot.

Files:
- `core/config.go`
- `core/introspection.go`
- `x/devtools/devtools.go`
- `cmd/plumego/internal/devserver/config_edit.go`
- `README.md`

Tests:
- Add or update tests to cover the stable snapshot shape consumed by devtools
  and the dashboard.
- `go test -race -timeout 60s ./core/...`
- `go test -timeout 20s ./x/devtools/... ./cmd/plumego/internal/devserver/...`

Docs Sync:
- Update the root README so `WithDebug`, `WithEnvPath`, and the `/_debug`
  contract are described consistently with `x/devtools`.

Done Definition:
- `core` has one canonical snapshot source for runtime/config introspection.
- Redundant exported getters that overlap with the snapshot contract are
  removed.
- First-party tooling no longer depends on undocumented snapshot drift.
- Root documentation matches the actual `core` + `x/devtools` contract.

Outcome:
- Added a single typed `core.RuntimeSnapshot` contract and replaced the old
  overlapping getter set with `(*App).RuntimeSnapshot()`.
- Removed redundant `DebugEnabled`, `EnvPath`, `ConfigSnapshot`, and the
  private duplicate `configSnapshot()` path from `core`.
- Updated `x/devtools` to consume the typed snapshot hook and keep the legacy
  JSON object shape as an explicit adapter at the debug endpoint boundary.
- Updated the dev dashboard analyzer/config editor to consume the typed
  snapshot instead of reading undocumented map keys directly.
- Added snapshot-focused tests in `core`, `x/devtools`, and
  `cmd/plumego/internal/devserver`.
- Updated `README.md` and `README_CN.md` to describe the stable `/_debug/config`
  snapshot contract consistently with `WithDebug`, `WithEnvPath`, and
  `x/devtools`.
- Validation:
  - `gofmt -w core/config.go core/introspection.go core/app_helpers.go core/introspection_test.go x/devtools/devtools.go x/devtools/devtools_test.go cmd/plumego/internal/devserver/analyzer.go cmd/plumego/internal/devserver/analyzer_test.go cmd/plumego/internal/devserver/config_edit.go cmd/plumego/internal/devserver/dashboard.go`
  - `go test -race -timeout 60s ./core/...`
  - `go test -timeout 20s ./x/devtools/...`
  - `go test -timeout 20s ./internal/devserver/...` (run from `cmd/plumego/`)
  - `go vet ./core/... ./x/devtools/...`
  - `go vet ./internal/devserver/...` (run from `cmd/plumego/`)
