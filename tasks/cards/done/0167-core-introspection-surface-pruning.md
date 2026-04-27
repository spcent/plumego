# Card 0167

Priority: P1
State: done
Primary Module: core
Owned Files:
- `core/introspection.go`
- `core/module.yaml`
- `core/introspection_test.go`
- `x/devtools/devtools.go`
- `x/devtools/devtools_test.go`
Depends On:
- `0164-core-runtime-state-contract-collapse.md`

Goal:
- Prune dead or presentation-oriented introspection helpers from `core` so the
  kernel exposes only structured state it actually owns.

Problem:
- `(*App).MiddlewareNames()` returns a `%T`-formatted string list, which is a
  tooling presentation detail rather than kernel state.
- The method currently has no direct production callers.
- `core` therefore still exports a public helper that is both unowned and
  presentation-shaped, which conflicts with the package's tighter kernel
  boundary.

Scope:
- Remove `MiddlewareNames()` from `core`.
- Move any remaining middleware-name presentation responsibility to the
  consuming tooling layer if it is still needed.
- Tighten tests and module metadata to the reduced introspection surface.

Non-goals:
- Do not add a replacement formatted helper in `core`.
- Do not broaden `core` to expose middleware internals structurally.
- Do not keep dead compatibility wrappers.

Files:
- `core/introspection.go`
- `core/module.yaml`
- `core/introspection_test.go`
- `x/devtools/devtools.go`
- `x/devtools/devtools_test.go`

Tests:
- `go test -race -timeout 60s ./core/...`
- `go test -timeout 20s ./x/devtools/...`
- `go vet ./core/... ./x/devtools/...`

Docs Sync:
- Remove `middleware list` from docs and manifest metadata if the pruned helper
  was the only remaining kernel surface for it.

Done Definition:
- `core` no longer exports `MiddlewareNames()`.
- Middleware-name presentation lives only in tooling that actually needs it, if
  at all.
- Module docs/metadata reflect the pruned introspection surface.

Outcome:
- Removed `(*App).MiddlewareNames()` from `core` because it was an unused
  presentation helper rather than owned kernel state.
- Verified that the symbol has no remaining Go call sites after removal.
- Updated `core/module.yaml` so runtime introspection now only advertises the
  remaining config snapshot surface.
