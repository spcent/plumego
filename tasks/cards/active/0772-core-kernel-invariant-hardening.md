# Card 0772

Priority: P2
State: active
Primary Module: core
Owned Files:
- `core/app_helpers.go`
- `core/middleware.go`
- `core/routing.go`
- `core/introspection.go`
- `core/options_test.go`
Depends On:
- `0771-core-handler-server-preparation-split.md`

Goal:
- Make `core.App` internals invariant instead of pseudo-optional, so the kernel
  stops carrying lazy nil-recovery paths for state that `New(...)` already owns.

Problem:
- `New(...)` always initializes config, router, middleware chain, and logger.
- `ensureRouter()` and `ensureMiddlewareChain()` still lazily recreate nil
  internals as if half-constructed `&App{}` values were supported.
- Some tests still exercise options by mutating raw `&App{}` values directly.
- This weakens kernel invariants and forces extra nil branches into normal
  runtime code for a construction mode the package should not treat as valid.

Scope:
- Make `New(...)` the only supported constructor path for live `core.App`
  values.
- Remove lazy recreation fallback for owned router / middleware internals.
- Update tests to build real apps instead of depending on partial `&App{}`
  state.
- Keep nil receiver behavior only where the public API intentionally needs it.

Non-goals:
- Do not remove exported `Option`.
- Do not change route or middleware behavior.
- Do not redesign public logger configuration.

Files:
- `core/app_helpers.go`
- `core/middleware.go`
- `core/routing.go`
- `core/introspection.go`
- `core/options_test.go`

Tests:
- `go test -race -timeout 60s ./core/...`
- `go vet ./core/...`

Docs Sync:
- No user-facing docs required unless public construction guidance changes.

Done Definition:
- Core runtime code no longer lazily rebuilds owned router/middleware state.
- Tests stop modeling raw `&App{}` as a supported live-app construction path.
- Kernel invariants are explicit and simpler to reason about.

Outcome:
