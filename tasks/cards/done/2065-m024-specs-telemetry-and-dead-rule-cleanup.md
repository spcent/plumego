# Card 2065

Milestone: M-024
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: control-plane
Priority: P1
State: done
Primary Module: specs
Owned Files:
- `specs/dependency-rules.yaml`
- `middleware/module.yaml`
- `docs/modules/middleware/README.md`
Depends On: none

## Goal

Make `middleware/internal/telemetry` visible in machine-readable boundary rules
and remove the dead `plumego.go` forbidden-path residue.

## Scope

Add an internal module entry or equivalent declared-caller rule for
`middleware/internal/telemetry`, document why stable middleware uses it, and
drop the stale special-rule path that no longer protects a real file.

## Non-goals

- Do not move telemetry helpers out of `middleware/internal/telemetry`.
- Do not change runtime middleware behavior.
- Do not expand the cleanup into unrelated dependency-rule renames.

## Files

- `specs/dependency-rules.yaml`
- `middleware/module.yaml`
- `docs/modules/middleware/README.md`

## Acceptance Tests

- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/module-manifests`

## Tests

- `go run ./internal/checks/agent-workflow`

## Docs Sync

- `docs/modules/middleware/README.md`

## Validation

- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/module-manifests`
- `gofmt -l .`

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

## Outcome

Declared `middleware/internal/telemetry` as a machine-readable internal module
with `middleware` as its caller surface, documented its stable observability
role in the middleware manifest and primer, and removed the dead
`plumego.go` forbidden-path residue from dependency rules.
