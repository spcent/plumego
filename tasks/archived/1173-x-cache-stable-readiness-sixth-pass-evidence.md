# 1173 - x/cache stable readiness sixth pass evidence

Status: done
Priority: P2
State: done
Primary module: `x/cache`

## Problem

After the distributed, leaderboard, and Redis adapter follow-up work, x/cache
release evidence must be updated so the remaining stable blockers are current
and actionable.

## Scope

- Update x/cache release evidence with the sixth-pass changes.
- Keep `x/cache/module.yaml` experimental unless all stable evidence is present.
- Record remaining stable blockers by surface.
- Run x/cache tests and module evidence checks.
- Move this card to done with actual validation commands.

## Out of Scope

- Promoting `x/cache` to stable.
- Adding owner sign-off without an actual owner review.
- Adding exported API snapshots beyond documenting the remaining blocker.

## Files

- `docs/extension-evidence/x-cache.md`
- `docs/modules/x-cache/README.md`
- `x/cache/module.yaml`
- `tasks/cards/done/1173-x-cache-stable-readiness-sixth-pass-evidence.md`

## Validation

- `go test -race -timeout 60s ./x/cache/...`
- `go vet ./x/cache/...`
- `go run ./internal/checks/module-manifests`

## Done Definition

Evidence reflects the completed sixth pass, remaining blockers are concrete, and
module status remains aligned with available evidence.

## Outcome

- Updated `docs/extension-evidence/x-cache.md` from fifth-pass to sixth-pass
  validation.
- Updated `docs/modules/x-cache/README.md` to point at sixth-pass evidence.
- Kept `x/cache/module.yaml` at `experimental`.
- Recorded that stable promotion still needs selected surface API snapshots,
  release refs, concrete Redis driver integration evidence, durable distributed
  repair decision, and owner sign-off.

## Validation Run

- `go test -race -timeout 60s ./x/cache/...`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
