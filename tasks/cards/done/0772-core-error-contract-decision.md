# 0772 - Core Error Contract Decision

State: done
Priority: P2
Primary module: core docs

## Goal

Freeze the stable decision that core lifecycle errors remain normal wrapped Go errors without exported lifecycle sentinel types.

## Scope

- Make the non-typed error contract explicit in Go doc and module docs.
- Add or adjust external tests to assert wrapped exported lower-level causes without depending on full strings.
- Keep operation names documented as stable diagnostic context.

## Non-goals

- Do not introduce exported core sentinel errors.
- Do not convert lifecycle errors into custom public error structs.
- Do not freeze complete human-readable error strings.

## Files

- `core/doc.go`
- `core/core_public_test.go`
- `docs/modules/core/README.md`
- `tasks/cards/active/0772-core-error-contract-decision.md`

## Tests

- `go test -timeout 20s ./core/...`
- `bash scripts/check-doc-snippets-compile.sh`

## Docs Sync

Update core module docs and Go package docs only.

## Done Definition

- Stable docs state how callers should handle core errors.
- External tests cover wrapped lower-level cause matching without full-string assertions.

## Validation

- `gofmt -w core/doc.go core/core_public_test.go`
- `go test -timeout 20s ./core/...`
- `bash scripts/check-doc-snippets-compile.sh`
