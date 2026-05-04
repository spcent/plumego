# 0735 - x/cache stability second pass docs

Status: active
Priority: P2
Primary module: `x/cache`

## Problem

After the second stabilization pass, `x/cache` docs and evidence need to reflect
implemented behavior and remaining blockers without claiming beta readiness.

## Scope

- Sync `docs/modules/x-cache/README.md` with current behavior.
- Update `docs/extension-evidence/x-cache.md` blocker inventory.
- Keep `x/cache/module.yaml` status as `experimental`.
- Record remaining promotion blockers precisely.

## Out of Scope

- Promoting `x/cache` to beta or GA.
- Inventing release-history evidence.
- Owner sign-off.

## Files

- `docs/modules/x-cache/README.md`
- `docs/extension-evidence/x-cache.md`
- `x/cache/module.yaml`

## Validation

- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`
- `go test -timeout 20s ./x/cache/...`

## Done Definition

Documentation matches implemented behavior, remaining blockers are explicit, and
the module remains experimental until formal evidence exists.
