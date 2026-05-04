# 0733 - core Shutdown Doc Examples

State: done
Priority: P2
Primary Module: core

## Goal

Synchronize remaining documentation snippets with the canonical shutdown-error
handling pattern.

## Scope

- Update stale `defer app.Shutdown(...)` snippets in core-facing docs.
- Keep examples compact and standard-library only.
- Preserve existing startup flow.

## Non-goals

- Do not change runtime code.
- Do not rewrite unrelated documentation sections.
- Do not change README examples already updated in earlier cards.

## Files

- `docs/getting-started.md`
- `docs/CANONICAL_STYLE_GUIDE.md`

## Tests

- `go run ./internal/checks/reference-layout`
- `bash scripts/check-doc-snippets-compile.sh`

## Docs Sync

Required for both files in scope.

## Done Definition

- Remaining core startup snippets no longer directly defer `Shutdown` while
  ignoring its return value.
- Documentation checks pass.

## Outcome

- Updated `docs/getting-started.md` and `docs/CANONICAL_STYLE_GUIDE.md` startup
  snippets to log deferred shutdown errors.
- Verified with `go run ./internal/checks/reference-layout`.
- Verified with `rg` that the scoped docs no longer contain direct deferred
  shutdown calls.
- `bash scripts/check-doc-snippets-compile.sh` was not run because this
  worktree does not contain that script.
