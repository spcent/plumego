# 1227 - x/cache stable readiness seventh pass evidence

Status: done
Priority: P2
State: done
Primary module: `x/cache`

## Problem

After the seventh x/cache stabilization pass, release evidence must be refreshed
to reflect the remaining blockers accurately without promoting the module.

## Scope

- Update x/cache evidence with seventh-pass lifecycle, callback, Redis, and
  leaderboard documentation changes.
- Keep `x/cache/module.yaml` experimental.
- Record remaining blockers by surface.
- Run x/cache tests and manifest/workflow checks.
- Move this card to done with actual validation commands.

## Out of Scope

- Promoting `x/cache` to stable.
- Adding owner sign-off without owner review.
- Creating API snapshots.

## Files

- `docs/extension-evidence/x-cache.md`
- `docs/modules/x-cache/README.md`
- `x/cache/module.yaml`
- `tasks/cards/done/1227-x-cache-stable-readiness-seventh-pass-evidence.md`

## Validation

- `go test -race -timeout 60s ./x/cache/...`
- `go vet ./x/cache/...`
- `go run ./internal/checks/module-manifests`

## Done Definition

Evidence reflects the seventh pass, remaining blockers are concrete, and module
status remains aligned with available evidence.

## Outcome

- Updated x/cache evidence from sixth-pass to seventh-pass validation.
- Kept `x/cache/module.yaml` at `experimental`.
- Recorded seventh-pass distributed lifecycle/callback, Redis compatibility
  field, and leaderboard documentation coverage.
- Left remaining stable blockers focused on surface selection, API snapshots,
  release refs, durable distributed repair decisions, concrete Redis driver
  integration evidence, and owner sign-off.

## Validation Run

- `go test -race -timeout 60s ./x/cache/...`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
