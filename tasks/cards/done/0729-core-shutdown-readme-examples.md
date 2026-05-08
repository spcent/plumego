# 0729 - core Shutdown README Examples

State: done
Priority: P2
Primary Module: core

## Goal

Make top-level README shutdown examples handle or explicitly discard
`Shutdown` errors so examples do not normalize silently ignored lifecycle errors.

## Scope

- Update English and Chinese README snippets that defer `Shutdown`.
- Keep example flow compact and standard-library only.
- Preserve existing documented behavior.

## Non-goals

- Do not rewrite the quickstart structure.
- Do not change runtime behavior.
- Do not touch reference application code in this card.

## Files

- `README.md`
- `README_CN.md`

## Tests

- `go run ./internal/checks/reference-layout`

## Docs Sync

Required in `README.md` and `README_CN.md`.

## Done Definition

- README examples no longer call `defer app.Shutdown(ctx)` directly.
- Reference layout check passes.

## Outcome

- Updated English and Chinese README examples to log shutdown errors from the
  deferred shutdown path.
- Verified with `go run ./internal/checks/reference-layout`.
