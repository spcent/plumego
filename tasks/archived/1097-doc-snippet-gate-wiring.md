# 1097 - doc Snippet Gate Wiring

State: done
Priority: P1
Primary Module: docs

## Goal

Wire the core documentation snippet compile check into the primary quality gates.

## Scope

- Run `bash scripts/check-doc-snippets-compile.sh` from `make gates`.
- Run the same check from the local pre-push hook in both full and quick modes.
- Keep the check focused on documentation examples; do not widen runtime behavior.

## Non-goals

- Do not change core runtime code.
- Do not change website build gates.
- Do not redesign the snippet extraction script.

## Files

- `Makefile`
- `scripts/pre-push`

## Tests

- `bash scripts/check-doc-snippets-compile.sh`
- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

## Docs Sync

Not required; this is gate wiring only.

## Done Definition

- Main quality gates and local pre-push hooks run the doc snippet compile check.
- Existing snippet compile check still passes.
- Core tests and vet pass.

## Outcome

- Added `bash scripts/check-doc-snippets-compile.sh` to `make gates`.
- Added the same check to both full milestone and quick local pre-push paths.
- Verified with `bash scripts/check-doc-snippets-compile.sh`,
  `go test -timeout 20s ./core/...`, and `go vet ./core/...`.
