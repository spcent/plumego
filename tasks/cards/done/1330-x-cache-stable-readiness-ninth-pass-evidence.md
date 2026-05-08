# 1330 - x/cache stable readiness ninth pass evidence

Status: done
Priority: P2
Primary module: `x/cache`

## Problem

After this stabilization pass, x/cache evidence and API snapshots need to
reflect implemented hardening while preserving truthful stable blockers.

## Scope

- Refresh x/cache evidence with ninth-pass distributed, leaderboard, and Redis
  changes.
- Keep `x/cache/module.yaml` status as `experimental`.
- Regenerate affected current-head API snapshots.
- Preserve release refs, historical snapshot comparisons, real Redis driver
  matrix, distributed repair, and owner sign-off blockers.
- Run x/cache validation and manifest/workflow checks.

## Out of Scope

- Manufacturing release refs.
- Owner sign-off without owner approval.
- Promoting x/cache or a child surface to stable.

## Files

- `docs/extension-evidence/x-cache.md`
- `docs/modules/x-cache/README.md`
- `docs/extension-evidence/snapshots/x-cache`
- `x/cache/module.yaml`
- `tasks/cards/done/1330-x-cache-stable-readiness-ninth-pass-evidence.md`

## Validation

- `go test -race -timeout 60s ./x/cache/...`
- `go vet ./x/cache/...`
- `go run ./internal/checks/module-manifests`

## Done Definition

Evidence and snapshots match the ninth pass, and all remaining stable blockers
remain explicit and truthful.

## Outcome

- Refreshed distributed, leaderboard, and Redis current-head API snapshots.
- Kept `x/cache/module.yaml` status as `experimental`.
- Updated ninth-pass evidence and preserved release refs, historical API
  snapshot comparisons, real Redis driver matrix, distributed repair, and owner
  sign-off as explicit blockers.
- Validation passed:
  - `go test -race -timeout 60s ./x/cache/...`
  - `go vet ./x/cache/...`
  - `go test -timeout 20s ./...`
  - `go vet ./...`
  - `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/dependency-rules`
  - `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/agent-workflow`
  - `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/module-manifests`
  - `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/reference-layout`
  - `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-api-snapshot -module ./x/cache/distributed -out docs/extension-evidence/snapshots/x-cache/x-cache-distributed-head.snapshot`
  - `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-api-snapshot -module ./x/cache/leaderboard -out docs/extension-evidence/snapshots/x-cache/x-cache-leaderboard-head.snapshot`
  - `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-api-snapshot -module ./x/cache/redis -out docs/extension-evidence/snapshots/x-cache/x-cache-redis-head.snapshot`
  - `env GOCACHE=/private/tmp/plumego-gocache go run ./internal/checks/extension-release-evidence -module ./x/cache/... -base HEAD -head HEAD`
