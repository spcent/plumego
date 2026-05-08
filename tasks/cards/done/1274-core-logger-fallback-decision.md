# 1274 - Core Logger Fallback Decision

State: done
Priority: P3
Primary module: core docs

## Goal

Resolve the logger fallback clarity issue without introducing hidden globals or lifecycle ownership.

## Scope

- Document that core resolves missing loggers to discard loggers and does not cache or lifecycle-manage logger fallbacks.
- Keep construction-time dependency resolution explicit.
- Add a focused public test for nil-app logger fallback behavior if needed.

## Non-goals

- Do not introduce package-level logger singletons.
- Do not make core start, stop, flush, or close loggers.
- Do not change `AppDependencies`.

## Files

- `core/core_public_test.go`
- `docs/modules/core/README.md`
- `tasks/cards/done/1274-core-logger-fallback-decision.md`

## Tests

- `go test -timeout 20s ./core/...`
- `bash scripts/check-doc-snippets-compile.sh`

## Docs Sync

Update `docs/modules/core/README.md` only.

## Done Definition

- Logger fallback behavior is explicitly documented.
- Core keeps no hidden global fallback logger.
- Core tests pass.

## Validation

- `go test -timeout 20s ./core/...`
- `bash scripts/check-doc-snippets-compile.sh`
