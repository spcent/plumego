---
id: 0787
title: x/data idempotency duplicate classifier
status: done
priority: P2
primary_module: x/data/idempotency
---

## Goal

Narrow SQL idempotency duplicate detection so unrelated constraint failures are
not silently converted into idempotent duplicates.

## Scope

- Remove the broad `"constraint"` fallback.
- Keep compatibility for common unique/duplicate wording.
- Prefer configured `DuplicateError` for driver-specific production handling.
- Add negative tests for non-unique constraint errors.

## Non-goals

- Add driver-specific imports.
- Change stable `store/idempotency` contracts.
- Add tenant or HTTP semantics.

## Files

- `x/data/idempotency/sql.go`
- `x/data/idempotency/*_test.go`

## Tests

- `go test -timeout 20s ./x/data/idempotency`
- `go test -race -timeout 60s ./x/data/idempotency`
- `go vet ./x/data/idempotency`

## Docs Sync

Update provider docs if duplicate classifier guidance is present.

## Done Definition

- Non-unique constraint errors return as errors.
- Unique/duplicate key wording still maps to `created=false, nil`.
- Custom classifier remains supported.

## Validation

- `go test -timeout 20s ./x/data/idempotency`
- `go test -race -timeout 60s ./x/data/idempotency`
- `go vet ./x/data/idempotency`
