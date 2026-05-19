# 1287 - x/cache stable readiness eighth pass evidence

Status: active
Priority: P2
State: done
Primary module: `x/cache`

## Problem

After this stabilization pass, x/cache evidence must distinguish implemented
hardening from remaining promotion blockers. The module must remain
experimental unless release refs, API snapshots, and owner sign-off actually
exist.

## Scope

- Refresh `docs/extension-evidence/x-cache.md` with eighth-pass outcomes.
- Keep `x/cache/module.yaml` status as `experimental`.
- Generate current-head exported API snapshots for the selected candidate
  surfaces where appropriate.
- Preserve release-history and owner-signoff blockers.
- Run x/cache validation and manifest/workflow checks.

## Out of Scope

- Manufacturing release refs.
- Adding owner sign-off without owner approval.
- Promoting `x/cache` or any sub-surface to stable.

## Files

- `docs/extension-evidence/x-cache.md`
- `docs/modules/x-cache/README.md`
- `docs/extension-evidence/snapshots/x-cache`
- `x/cache/module.yaml`
- `tasks/cards/done/1287-x-cache-stable-readiness-eighth-pass-evidence.md`

## Validation

- `go test -race -timeout 60s ./x/cache/...`
- `go vet ./x/cache/...`
- `go run ./internal/checks/module-manifests`

## Done Definition

Evidence records the eighth pass accurately, current-head snapshots are present
where useful, and all remaining stable blockers are truthful and actionable.

## Outcome

- Refreshed x/cache evidence from seventh-pass to eighth-pass validation.
- Generated current-head API snapshots for distributed, leaderboard, and Redis
  surfaces under `docs/extension-evidence/snapshots/x-cache/`.
- Kept `x/cache/module.yaml` at `experimental`.
- Preserved release refs, historical snapshot comparisons, real Redis driver
  evidence, durable distributed repair, and owner sign-off as blockers.
- Updated the module primer to reference eighth-pass evidence and snapshot
  limits.

## Validation Run

- `go test -race -timeout 60s ./x/cache/...`
- `go test -timeout 20s ./x/cache/...`
- `go vet ./x/cache/...`
- `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/module-manifests`
- `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/agent-workflow`
- `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-api-snapshot -module ./x/cache/distributed -out docs/extension-evidence/snapshots/x-cache/x-cache-distributed-head.snapshot`
- `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-api-snapshot -module ./x/cache/leaderboard -out docs/extension-evidence/snapshots/x-cache/x-cache-leaderboard-head.snapshot`
- `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-api-snapshot -module ./x/cache/redis -out docs/extension-evidence/snapshots/x-cache/x-cache-redis-head.snapshot`
