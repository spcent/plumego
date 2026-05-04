# 0745 - x/cache stable readiness fourth pass evidence

Status: active
Priority: P2
Primary module: `x/cache`

## Problem

Even after implementation hardening, `x/cache` cannot be treated as stable
without updated evidence that identifies remaining blockers by surface.

## Scope

- Update the x/cache primer with fourth-pass behavior.
- Update extension evidence with remaining distributed, leaderboard, and Redis
  blockers.
- Keep module status experimental unless the release evidence policy is fully
  satisfied.
- Record validation commands actually run during this pass.

## Out of Scope

- Promoting `x/cache` or any child package to beta/stable.
- Generating release API snapshots unless the repo tooling already supports the
  chosen surface without widening scope.

## Files

- `docs/modules/x-cache/README.md`
- `docs/extension-evidence/x-cache.md`
- `x/cache/module.yaml`
- `tasks/cards/active/0745-x-cache-stable-readiness-fourth-pass-evidence.md`

## Validation

- `go test -race -timeout 60s ./x/cache/...`
- `go vet ./x/cache/...`
- `go run ./internal/checks/agent-workflow`

## Done Definition

The docs and evidence ledger reflect the fourth stabilization pass and clearly
state why `x/cache` remains experimental.

## Outcome

Pending.
