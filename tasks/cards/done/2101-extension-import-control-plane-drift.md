# Card 2101: Align Extension Import Control Plane With Current Code

Priority: P1
State: done
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: specs
Owned Files:
- specs/dependency-rules.yaml
- x/fileapi/module.yaml
- x/messaging/module.yaml
- internal/checks/dependency-rules/main.go
- internal/checks/checkutil/
Depends On:

## Goal

The machine-readable control plane should not disagree with current extension
imports.  Today the dependency checker passes, but module manifests still give
agents conflicting guidance:

- `x/fileapi/handler.go` imports `x/tenant/core` and the module docs explicitly
  require tenant context, but `x/fileapi/module.yaml` does not list `x/tenant`
  under `allowed_imports`.
- `x/messaging/quota.go` imports `x/tenant/core`, and
  `specs/dependency-rules.yaml` allows `x/tenant` for `x/messaging`, but
  `x/messaging/module.yaml` still forbids `x/tenant/**`.

This is not a compile failure, but it makes the repo control plane unreliable:
Codex can be told by one official file that an import is allowed and by another
that the same import is forbidden.

## Scope

- Decide the canonical policy for extension-to-extension imports in:
  - `x/fileapi -> x/tenant`
  - `x/messaging -> x/tenant`
- Update the module manifests and dependency rules so they agree.
- Add or tighten a check that fails when a module's manifest and
  `specs/dependency-rules.yaml` disagree for the same import family.
- Keep the rule narrowly scoped to declared library modules; do not turn this
  into a full Go import analyzer.

## Non-goals

- Do not move `x/fileapi` tenant extraction into stable middleware or store.
- Do not remove `x/messaging` quota integration unless a separate product
  decision says messaging must become tenant-agnostic.
- Do not add new dependencies.
- Do not change runtime behavior.

## Files

- `specs/dependency-rules.yaml`
- `x/fileapi/module.yaml`
- `x/messaging/module.yaml`
- `internal/checks/dependency-rules/main.go`
- `internal/checks/checkutil/`

## Tests

```bash
go run ./internal/checks/module-manifests
go run ./internal/checks/dependency-rules
go test -timeout 20s ./internal/checks/...
```

## Docs Sync

Update module docs only if the policy changes.  If the implementation already
matches the docs and only the manifests were stale, no docs sync is required.

## Done Definition

- `x/fileapi` import guidance is internally consistent across docs, manifest,
  and dependency rules.
- `x/messaging` import guidance is internally consistent across docs, manifest,
  and dependency rules.
- A future contradiction between module manifest imports and dependency rules is
  caught by a repo check.
- The listed validation commands pass.

## Outcome

Completed.

- Added `x/tenant` to `x/fileapi/module.yaml` allowed imports to match its
  tenant-context handler dependency and existing module docs.
- Added `x/tenant` to `x/messaging/module.yaml` allowed imports and removed
  the stale `x/tenant/**` forbidden entry, matching the existing quota checker.
- Narrowed the `x/fileapi` dependency-rule allow entry from `x/data` to
  `x/data/file` to match the manifest and docs.
- Removed stale `x/observability` from the `x/messaging` dependency-rule allow
  list because production code does not import it.
- Extended `dependency-rules` with an `x/*` manifest-vs-dependency consistency
  check so future extension allow/deny contradictions fail the gate.
- Added focused checkutil coverage for contradiction detection.

Validation:

```bash
go run ./internal/checks/module-manifests
go run ./internal/checks/dependency-rules
go test -timeout 20s ./internal/checks/...
```
