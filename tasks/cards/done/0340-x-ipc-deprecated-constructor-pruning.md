# Card 0340: X IPC Deprecated Constructor Pruning

Priority: P1
State: done
Recipe: specs/change-recipes/symbol-change.yaml
Primary Module: x/ipc
Depends On: —

## Goal

Remove the remaining deprecated `x/ipc` constructor wrappers now that the
functional-options path is the canonical API, so the package no longer carries
duplicate public entrypoints for the same behavior.

## Problem

- `x/ipc/ipc.go` still exports `NewServerWithConfig` and `DialWithConfig` even
  though the package docs already mark them deprecated in favor of
  `NewServer(...opts)` and `Dial(...opts)`.
- Current repo references are limited to `x/ipc` tests, which means the
  deprecated symbols are now mostly compatibility tail rather than active API
  necessity.
- Leaving both constructor families in place keeps the exported surface larger
  than necessary and weakens the repository rule that deprecated symbols should
  be removed once their last caller is migrated.

## Scope

- Enumerate and migrate every in-repo use of `NewServerWithConfig` and
  `DialWithConfig`.
- Remove the deprecated wrapper functions from `x/ipc/ipc.go` in the same
  change.
- Update `x/ipc` tests and package docs/examples to use the functional-options
  path only.

## Non-Goals

- Do not redesign IPC config semantics, reconnect behavior, or platform
  backends.
- Do not change the `Config` type itself unless a test needs a functional
  option equivalent to preserve behavior.
- Do not widen this card into broader IPC API redesign.

## Files

- `x/ipc/ipc.go`
- `x/ipc/ipc_test.go`
- `x/ipc/doc.go` or package comments if they still mention the deprecated
  constructors

## Tests

```bash
rg -n 'NewServerWithConfig|DialWithConfig' x/ipc . -g '*.go'
go test -timeout 20s ./x/ipc/...
go vet ./x/ipc/...
```

## Docs Sync

- any `x/ipc` examples or package comments that still teach the deprecated
  constructor names

## Done Definition

- `NewServerWithConfig` and `DialWithConfig` have zero remaining references and
  are removed from `x/ipc`.
- `x/ipc` tests use only the functional-options constructor path.
- Package comments and examples no longer advertise the deprecated API.
- Focused `x/ipc` validation passes.

## Outcome

- Removed deprecated `NewServerWithConfig` and `DialWithConfig` from `x/ipc`.
- Migrated the remaining example/test callers to the functional-options path by
  introducing a local `optionsFromConfig` helper in `x/ipc/ipc_test.go`.
- Re-ran the exported-symbol grep to confirm the removed constructor names have
  zero remaining in-repo references.

## Validation Run

```bash
rg -n 'NewServerWithConfig|DialWithConfig' . -g '*.go'
go test -timeout 20s ./x/ipc/...
go vet ./x/ipc/...
```
