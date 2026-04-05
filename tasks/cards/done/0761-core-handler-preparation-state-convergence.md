# Card 0761

Priority: P1
State: done
Primary Module: core
Owned Files:
- `core/http_handler.go`
- `core/lifecycle.go`
- `core/introspection.go`
- `core/app_test.go`
- `core/lifecycle_test.go`
- `docs/modules/core/README.md`
Depends On:

Goal:
- Make `core` freeze/build handler state through one explicit internal
  transition so `ServeHTTP` and `Prepare` cannot leave the app in divergent
  lifecycle states.

Problem:
- `ServeHTTP` currently calls `ensureHandler()` directly and freezes config and
  router on first request.
- `Prepare` calls `setupServer()`, which also calls `ensureHandler()` before
  constructing `http.Server`.
- That means `core` has two preparation entrypoints: one handler-only path and
  one server-backed path.
- The runtime snapshot reports `ConfigFrozen` and `ServerPrepared` separately,
  so the handler-only path can freeze the app while still looking “not
  prepared” to introspection and tooling.
- Tests already exercise both paths independently, which keeps the split live.

Scope:
- Define one internal handler-preparation transition used by both `ServeHTTP`
  and `Prepare`.
- Make lifecycle/introspection state mean the same thing regardless of whether
  the app is used as an `http.Handler` or as a prepared server.
- Stop testing private lifecycle entrypoints directly when the same behaviour
  can be exercised through the canonical surface.

Non-goals:
- Do not remove `net/http` compatibility from `core.App`.
- Do not redesign route freezing semantics in `router`.
- Do not add a new public lifecycle abstraction.

Files:
- `core/http_handler.go`
- `core/lifecycle.go`
- `core/introspection.go`
- `core/app_test.go`
- `core/lifecycle_test.go`
- `docs/modules/core/README.md`

Tests:
- Add coverage that the app reaches the same immutable/prepared state whether
  it is first touched via `ServeHTTP` or `Prepare`.
- Remove direct reliance on private setup helpers where public lifecycle APIs
  can assert the same behaviour.
- `go test -race -timeout 60s ./core/...`
- `go vet ./core/...`

Docs Sync:
- Update the core module primer so it describes one preparation/freeze model
  instead of implying separate handler and server state machines.

Done Definition:
- `ServeHTTP` and `Prepare` share one internal preparation transition.
- Runtime/introspection state no longer drifts based on which entrypoint was
  touched first.
- Core tests assert lifecycle behaviour through the canonical public surface.

Outcome:
- Added one shared internal `ensurePrepared()` transition used by both
  `Prepare()` and `ServeHTTP()`.
- The shared preparation path now freezes config/router state, builds the
  handler, and constructs the backing `http.Server` in one place.
- Updated core tests to assert the canonical public preparation surface instead
  of relying on private setup helpers for equivalent behaviour.
- Updated the core module primer to describe one preparation/freeze model.
