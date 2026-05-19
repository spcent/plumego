# 1224 - Core ServeHTTP Prepare Failure Contract

State: done
Priority: P1
Primary module: core

## Goal

Make the `ServeHTTP` then `Prepare` failure state explicit and tested as a stable contract.

## Scope

- Add external package coverage for an app that is handler-prepared through `ServeHTTP` and later fails `Prepare` because of server-only config.
- Document that this state is intentionally handler-frozen but not server-prepared.
- Preserve the existing split: `ServeHTTP` stays handler-only and `Prepare` owns server config validation.

## Non-goals

- Do not make `ServeHTTP` validate server-only config.
- Do not alter route freezing semantics.
- Do not add new lifecycle states.

## Files

- `core/core_public_test.go`
- `docs/modules/core/README.md`
- `tasks/cards/done/1224-core-servehttp-prepare-failure-contract.md`

## Tests

- `go test -timeout 20s ./core/...`
- `bash scripts/check-doc-snippets-compile.sh`

## Docs Sync

Update `docs/modules/core/README.md` only.

## Done Definition

- Public tests prove the post-`ServeHTTP` failed-`Prepare` state.
- Core docs describe the behavior in the frozen matrix.

## Validation

- `go test -timeout 20s ./core/...`
- `bash scripts/check-doc-snippets-compile.sh`
