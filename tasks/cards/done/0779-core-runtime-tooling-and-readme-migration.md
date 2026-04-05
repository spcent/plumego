# Card 0779

Priority: P1
State: done
Primary Module: core
Owned Files:
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/devserver/dashboard.go`
- `README.md`
- `README_CN.md`
- `docs/getting-started.md`
Depends On:
- `0777-core-runtime-activation-phase-collapse.md`
- `0778-core-runtime-activation-reference-migration.md`

Goal:
- Migrate first-party tooling and top-level guides to the collapsed core
  lifecycle and current explicit config/metrics semantics.

Problem:
- The scaffold template and devserver dashboard still call `Start()`.
- Root guides still teach the old multi-phase lifecycle.
- `README.md` and `README_CN.md` also still attribute request-body limits,
  concurrency limits, and env-file fields to `core` even though those concerns
  now live in middleware or app-local tooling.

Scope:
- Update scaffold and devserver bootstraps to remove `Start()`.
- Fix root guides so they describe the current `core` config surface and
  explicit middleware/devtools ownership boundaries.
- Keep examples and generated app layouts aligned with the canonical reference
  apps.

Non-goals:
- Do not redesign CLI UX.
- Do not add compatibility shims for the removed lifecycle phase.
- Do not move middleware or devtools concerns back into `core`.

Files:
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/devserver/dashboard.go`
- `README.md`
- `README_CN.md`
- `docs/getting-started.md`

Tests:
- `go test -timeout 20s ./cmd/plumego/internal/scaffold/... ./cmd/plumego/internal/devserver/...`
- `go test -timeout 20s ./...`
- `go vet ./...`

Docs Sync:
- Root guides must stop describing old core lifecycle phases or non-core
  config/runtime fields.

Done Definition:
- First-party tooling no longer calls `core.Start(...)`.
- Root guides describe the current core surface without stale config claims.
- Generated and documented app bootstrap matches the canonical lifecycle.

Outcome:
- Migrated scaffold and devserver bootstraps off `core.Start(...)`.
- Updated root guides and getting-started docs to the collapsed lifecycle.
- Removed stale README claims that request-body limits and concurrency limits
  are part of `core.AppConfig`.
