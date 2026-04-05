# Card 0754

Priority: P2
State: active
Primary Module: core
Owned Files:
- `core/config.go`
- `core/introspection.go`
- `x/devtools/devtools.go`
- `cmd/plumego/internal/devserver/config_edit.go`
- `README.md`
Depends On:
- `0753-core-route-registration-surface-convergence.md`

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
