# 0730 - x/cache stability docs and evidence

Status: active
Priority: P2
Primary module: `x/cache`

## Problem

`x/cache` has a primer and manifest, but stable-readiness blockers are not
recorded with enough specificity. After correctness cleanup, docs must describe
implemented behavior only and keep promotion blockers explicit.

## Scope

- Sync `docs/modules/x-cache/README.md` with implemented distributed,
  leaderboard, and Redis behavior.
- Update maturity/evidence notes if blocker state changes.
- Keep `x/cache/module.yaml` status as `experimental` unless release evidence
  and owner sign-off exist.
- Record remaining promotion blockers without inventing release history.

## Out of Scope

- Promoting `x/cache` to beta or GA.
- Owner sign-off.
- Broad `x/data` documentation rewrite.

## Files

- `docs/modules/x-cache/README.md`
- `x/cache/module.yaml`
- `docs/extension-evidence/x-data.md` if cache readiness is tracked there
- `specs/extension-beta-evidence.yaml` only if ledger state changes

## Validation

- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`
- `go test -timeout 20s ./x/cache/...`

## Done Definition

Docs and manifest state match implemented behavior, remaining stable blockers
are explicit, and no promotion status is claimed without evidence.
