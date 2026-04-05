# Card 0778

Priority: P1
State: active
Primary Module: core
Owned Files:
- `reference/standard-service/internal/app/app.go`
- `reference/with-gateway/internal/app/app.go`
- `reference/with-messaging/internal/app/app.go`
- `reference/with-webhook/internal/app/app.go`
- `reference/with-websocket/internal/app/app.go`
Depends On:
- `0777-core-runtime-activation-phase-collapse.md`

Goal:
- Migrate the canonical reference applications to the collapsed `core`
  activation path immediately after `Start()` is removed.

Problem:
- Every first-party reference app still calls `Core.Start(ctx)` after
  `Prepare()`.
- Those references are the canonical application layouts for this repository,
  so leaving them on the old phase model would preserve user confusion even if
  the kernel itself is cleaned up.

Scope:
- Update all five reference application bootstraps to use the new collapsed
  activation path.
- Keep route wiring, shutdown ownership, and application-local behavior
  unchanged otherwise.

Non-goals:
- Do not redesign reference application structure.
- Do not introduce app-local wrappers that recreate the old `Start()` concept.
- Do not change feature module behavior in the reference apps.

Files:
- `reference/standard-service/internal/app/app.go`
- `reference/with-gateway/internal/app/app.go`
- `reference/with-messaging/internal/app/app.go`
- `reference/with-webhook/internal/app/app.go`
- `reference/with-websocket/internal/app/app.go`

Tests:
- `go test -timeout 20s ./reference/...`
- `go test -timeout 20s ./...`
- `go vet ./...`

Docs Sync:
- Keep the canonical reference layouts aligned with the collapsed kernel
  lifecycle.

Done Definition:
- No reference app calls `core.Start(...)`.
- The canonical application layouts demonstrate the collapsed lifecycle path.
- Reference tests and repo-wide gates pass with the migrated bootstraps.

Outcome:
