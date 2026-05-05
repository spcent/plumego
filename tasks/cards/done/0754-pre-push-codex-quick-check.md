# 0754 - Pre-push Codex Quick Check

State: done
Priority: P1
Primary module: scripts

## Goal

Ensure Codex-created branches run the same local quick checks as other normal development branches before push.

## Scope

- Include `codex/*` in the non-milestone quick-check branch match in `scripts/pre-push`.
- Keep existing quick-check command order unchanged.

## Non-goals

- Do not change milestone branch gates.
- Do not add new quick-check commands.
- Do not change CI configuration.

## Files

- `scripts/pre-push`

## Tests

- `bash scripts/check-doc-snippets-compile.sh`
- `go test -timeout 20s ./core/...`

## Docs Sync

No docs update required; this is local hook behavior.

## Done Definition

- `scripts/pre-push` runs quick checks for `codex/*` branches.
- Existing doc snippet compile gate still passes.
- Core tests pass.

## Validation

- `bash scripts/check-doc-snippets-compile.sh`
- `go test -timeout 20s ./core/...`
