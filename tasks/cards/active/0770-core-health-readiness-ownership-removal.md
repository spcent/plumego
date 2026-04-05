# Card 0770

Priority: P1
State: active
Primary Module: core
Owned Files:
- `core/app.go`
- `core/options.go`
- `core/lifecycle.go`
- `core/options_test.go`
- `docs/modules/core/README.md`
Depends On:

Goal:
- Remove misleading readiness ownership from `core` so the kernel stops claiming
  server-ready state it does not actually control.

Problem:
- `core` no longer owns `ListenAndServe()`; callers do that explicitly on the
  prepared `*http.Server`.
- `(*App).Start(ctx)` still calls `healthManager.MarkReady()` before any serve
  call happens.
- `(*App).Shutdown(ctx)` still calls `MarkNotReady(...)` even though the kernel
  no longer owns the outer serving loop.
- This leaves `WithHealthManager(...)` as a confusing half-contract: `core`
  reports readiness based on runtime-hook start, not on actual traffic
  readiness.

Scope:
- Remove `WithHealthManager(...)` and the `healthManager` field from `core`.
- Stop `core.Start` / `core.Shutdown` from mutating external readiness state.
- Keep health wiring explicit in app-local code or the owning health package.

Non-goals:
- Do not redesign the `health` package.
- Do not add a replacement readiness callback in this card.
- Do not change HTTP serving behavior.

Files:
- `core/app.go`
- `core/options.go`
- `core/lifecycle.go`
- `core/options_test.go`
- `docs/modules/core/README.md`

Tests:
- `go test -race -timeout 60s ./core/...`
- `go vet ./core/...`

Docs Sync:
- Update the core primer so readiness ownership is no longer described as a
  kernel concern.

Done Definition:
- `core` no longer exposes `WithHealthManager(...)`.
- `core.Start` / `core.Shutdown` no longer mutate external readiness state.
- Docs/tests treat readiness as app-local wiring, not kernel-owned behavior.

Outcome:
