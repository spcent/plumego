# 1121 - core Metrics Ownership Doc Precision

State: done
Priority: P2
Primary Module: core

## Goal

Make the core primer precise about metrics ownership.

## Scope

- Update `docs/modules/core/README.md` so metrics are described as middleware or
  app-local wiring, not constructor-owned core dependencies.
- Preserve the existing logger dependency wording.

## Non-goals

- Do not change `core.AppDependencies`.
- Do not add metrics dependencies or runtime behavior to core.
- Do not update unrelated module primers.

## Files

- `docs/modules/core/README.md`

## Tests

- `bash scripts/check-doc-snippets-compile.sh`
- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

## Docs Sync

Required in the core primer only.

## Done Definition

- The core primer no longer implies metrics are constructor-owned by core.
- Documentation snippet compile check, core tests, and vet pass.

## Outcome

- Clarified that core constructor wiring owns passive runtime dependencies such
  as logging, while metrics attach through middleware or app-local wiring.
- Verified with `bash scripts/check-doc-snippets-compile.sh`,
  `go test -timeout 20s ./core/...`, and `go vet ./core/...`.
